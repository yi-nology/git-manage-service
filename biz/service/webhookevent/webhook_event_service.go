package webhookevent

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/codereview"
	syncv2 "github.com/yi-nology/git-manage-service/biz/service/sync/v2"
	"github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-platform-sdk/provider"
)

func List(eventType, source, status string, page, pageSize int) ([]api.WebhookEventDTO, int, error) {
	dao := db.NewWebhookEventDAO()
	events, total, err := dao.List(eventType, source, status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	dtos := make([]api.WebhookEventDTO, 0, len(events))
	for _, e := range events {
		dtos = append(dtos, toEventDTO(&e))
	}
	return dtos, int(total), nil
}

func Retry(eventID uint) error {
	dao := db.NewWebhookEventDAO()
	var event po.WebhookEvent
	if err := db.DB.First(&event, eventID).Error; err != nil {
		return fmt.Errorf("event not found: %w", err)
	}
	event.Status = "received"
	event.ErrorMessage = ""
	return dao.Save(&event)
}

func ProcessIncomingEvent(event *provider.NormalizedEvent, providerCfgID uint) error {
	dao := db.NewWebhookEventDAO()
	eventID := event.ID

	_, err := dao.FindByEventID(eventID)
	if err == nil {
		return nil
	}

	var repoID uint
	if event.Repo != nil {
		if parts := strings.SplitN(event.Repo.FullName, "/", 2); len(parts) == 2 {
			if r, err := db.NewRepoDAO().FindByPlatformOwnerRepo(parts[0], parts[1]); err == nil {
				repoID = r.ID
			}
		}
	}

	var crID uint
	var platformCRNum int
	if event.CR != nil {
		platformCRNum = event.CR.Number
		if repoID > 0 {
			crDAO := db.NewChangeRequestDAO()
			if localCR, err := crDAO.FindByRepoAndNumber(repoID, event.CR.Number); err == nil {
				crID = localCR.ID
			}
		}
	}

	actorName := ""
	actorUsername := ""
	if event.Actor != nil {
		actorName = event.Actor.Name
		actorUsername = event.Actor.Username
	}

	payload := map[string]interface{}{
		"type":   event.Type,
		"source": string(event.Source),
	}
	if event.Branch != "" {
		payload["branch"] = event.Branch
	}
	if event.Tag != "" {
		payload["tag"] = event.Tag
	}
	if event.CommitSHA != "" {
		payload["commit_sha"] = event.CommitSHA
	}

	whEvent := &po.WebhookEvent{
		EventID:          eventID,
		ProviderConfigID: providerCfgID,
		EventType:        event.Type,
		Source:           string(event.Source),
		RepoID:           repoID,
		CRID:             crID,
		PlatformCRNumber: platformCRNum,
		ActorName:        actorName,
		ActorUsername:    actorUsername,
		Payload:          payload,
		Status:           "received",
	}
	if err := dao.Create(whEvent); err != nil {
		return err
	}

	go applyRules(whEvent)

	return nil
}

func applyRules(event *po.WebhookEvent) {
	ruleDAO := db.NewWebhookRuleDAO()
	rules, err := ruleDAO.FindByProviderConfigID(event.ProviderConfigID)
	if err != nil {
		log.Printf("[webhook] failed to load rules for provider_config %d (event %s): %v", event.ProviderConfigID, event.EventID, err)
		return
	}

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}
		if !matchPattern(rule.EventTypePattern, event.EventType) {
			continue
		}
		if rule.RepoPattern != "" && rule.RepoPattern != "*" {
			repoDAO := db.NewRepoDAO()
			if repo, err := repoDAO.FindByID(event.RepoID); err == nil {
				fullName := repo.PlatformOwner + "/" + repo.PlatformRepo
				if !matchPattern(rule.RepoPattern, fullName) {
					continue
				}
			}
		}

		switch rule.Action {
		case "sync":
			triggerSync(rule.ActionConfig)
		case "notify":
			log.Printf("Webhook rule %s: notify action triggered for event %s", rule.Name, event.EventID)
		case "code_review":
			triggerCodeReview(event)
		}
	}

	now := time.Now()
	event.Status = "processed"
	event.ProcessedAt = &now
	eventDAO := db.NewWebhookEventDAO()
	if err := eventDAO.Save(event); err != nil {
		log.Printf("[webhook] failed to mark event %s processed: %v", event.EventID, err)
	}
}

// triggerSync runs the configured sync task via git-sync-service. Previously
// this was a log-only stub, so the "触发同步任务" webhook rule action was
// silently broken; it now delegates to syncv2.RunTask (the real engine).
func triggerSync(config map[string]interface{}) {
	taskKey, _ := config["sync_task_key"].(string)
	if taskKey == "" {
		return
	}
	svc := syncv2.GetService()
	if svc.GetCore() == nil {
		log.Printf("Webhook sync trigger skipped: sync service not initialized (task %s)", taskKey)
		return
	}
	if err := svc.RunTask(context.Background(), taskKey); err != nil {
		log.Printf("Webhook sync trigger failed for task %s: %v", taskKey, err)
	}
}

func triggerCodeReview(event *po.WebhookEvent) {
	if !configs.GetCodeReviewConfig().Enabled {
		return
	}
	if !configs.GetCodeReviewConfig().AutoReviewOnMR {
		return
	}

	switch event.EventType {
	case "cr.open", "cr.update", "cr.push":
		triggerCRReview(event)
	case "push":
		triggerPushReview(event)
	case "cr.note":
		triggerNoteReview(event)
	default:
		if strings.HasPrefix(event.EventType, "cr.") {
			triggerCRReview(event)
		}
	}
}

func triggerCRReview(event *po.WebhookEvent) {
	if event.RepoID == 0 || event.PlatformCRNumber == 0 {
		return
	}

	repo, err := db.NewRepoDAO().FindByID(event.RepoID)
	if err != nil {
		log.Printf("[CodeReview] trigger: repo %d not found: %v", event.RepoID, err)
		return
	}

	if !shouldReviewBranch(repo, event) {
		return
	}

	if repo.ProviderConfigID != 0 && repo.PlatformOwner != "" && repo.PlatformRepo != "" {
		repoCfg, err := db.NewReviewRepoConfigDAO().FindByRemoteRepo(repo.ProviderConfigID, repo.PlatformOwner, repo.PlatformRepo)
		if err == nil && !repoCfg.AutoReviewOnMR {
			return
		}
	}

	mrIID := fmt.Sprintf("%d", event.PlatformCRNumber)
	commitSHA := ""
	if event.Payload != nil {
		if sha, ok := event.Payload["commit_sha"].(string); ok {
			commitSHA = sha
		}
	}

	_, err = codereview.CreateTask(context.Background(), repo.Key, event.ProviderConfigID, mrIID, commitSHA, "webhook")
	if err != nil {
		log.Printf("[CodeReview] trigger: failed to create task: %v", err)
	}
}

func triggerPushReview(event *po.WebhookEvent) {
	if event.RepoID == 0 {
		return
	}
	repo, err := db.NewRepoDAO().FindByID(event.RepoID)
	if err != nil {
		return
	}
	if !shouldReviewBranch(repo, event) {
		return
	}
	branch := ""
	if event.Payload != nil {
		if b, ok := event.Payload["branch"].(string); ok {
			branch = b
		}
	}
	if branch == "" {
		return
	}
	log.Printf("[CodeReview] Push event review triggered for repo %s branch %s (skipped: no MR to attach)", repo.Key, branch)
}

func triggerNoteReview(event *po.WebhookEvent) {
	if event.RepoID == 0 || event.PlatformCRNumber == 0 {
		return
	}

	noteText := ""
	if event.Payload != nil {
		if t, ok := event.Payload["note"].(string); ok {
			noteText = t
		}
	}
	if noteText == "" {
		return
	}
	if !strings.Contains(strings.ToLower(noteText), "/review") {
		return
	}

	repo, err := db.NewRepoDAO().FindByID(event.RepoID)
	if err != nil {
		return
	}

	mrIID := fmt.Sprintf("%d", event.PlatformCRNumber)
	commitSHA := ""
	if event.Payload != nil {
		if sha, ok := event.Payload["commit_sha"].(string); ok {
			commitSHA = sha
		}
	}
	_, err = codereview.CreateTask(context.Background(), repo.Key, event.ProviderConfigID, mrIID, commitSHA, "note_command")
	if err != nil {
		log.Printf("[CodeReview] note trigger: failed to create task: %v", err)
	}
}

func shouldReviewBranch(repo *po.Repo, event *po.WebhookEvent) bool {
	branch := ""
	if event.Payload != nil {
		if b, ok := event.Payload["branch"].(string); ok {
			branch = b
		}
	}
	if branch == "" && event.EventType == "push" {
		return false
	}

	if repo.ProviderConfigID != 0 && repo.PlatformOwner != "" && repo.PlatformRepo != "" {
		repoCfg, err := db.NewReviewRepoConfigDAO().FindByRemoteRepo(repo.ProviderConfigID, repo.PlatformOwner, repo.PlatformRepo)
		if err != nil {
			return true
		}
		scopeNote := repoCfg.ScopeNote
		if scopeNote == "" {
			return true
		}
		patterns := strings.Split(scopeNote, ",")
		for _, p := range patterns {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			if matchPattern(p, branch) {
				return true
			}
		}
		return false
	}
	return true
}

var regexCache sync.Map // map[string]*regexp.Regexp — avoids recompiling per event

func matchPattern(pattern, value string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	if !strings.Contains(pattern, "*") {
		return pattern == value
	}
	regexStr := "^" + strings.ReplaceAll(regexp.QuoteMeta(pattern), "\\*", ".*") + "$"
	var re *regexp.Regexp
	if v, ok := regexCache.Load(regexStr); ok {
		re = v.(*regexp.Regexp)
	} else {
		compiled, err := regexp.Compile(regexStr)
		if err != nil {
			return false
		}
		re = compiled
		regexCache.Store(regexStr, re)
	}
	return re.MatchString(value)
}

func toEventDTO(e *po.WebhookEvent) api.WebhookEventDTO {
	return api.WebhookEventDTO{
		ID:               e.ID,
		EventID:          e.EventID,
		EventType:        e.EventType,
		Source:           e.Source,
		RepoID:           e.RepoID,
		CRID:             e.CRID,
		PlatformCRNumber: e.PlatformCRNumber,
		ActorName:        e.ActorName,
		ActorUsername:    e.ActorUsername,
		Status:           e.Status,
		ProcessedAt:      e.ProcessedAt,
		CreatedAt:        e.CreatedAt,
		ErrorMessage:     e.ErrorMessage,
	}
}
