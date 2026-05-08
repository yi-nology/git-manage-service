package webhookevent

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/codereview"
	"github.com/yi-nology/git-manage-service/biz/service/provider"
	"github.com/yi-nology/git-manage-service/pkg/configs"
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
		repoDAO := db.NewRepoDAO()
		repos, rErr := repoDAO.FindAll()
		if rErr == nil {
			for _, r := range repos {
				if r.PlatformOwner+"/"+r.PlatformRepo == event.Repo.FullName {
					repoID = r.ID
					break
				}
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
	eventDAO.Save(event)
}

func triggerSync(config map[string]interface{}) {
	taskKey, _ := config["sync_task_key"].(string)
	if taskKey == "" {
		return
	}
	log.Printf("Webhook rule triggered sync for task: %s", taskKey)
}

func triggerCodeReview(event *po.WebhookEvent) {
	if !configs.GlobalConfig.CodeReview.Enabled {
		return
	}
	if !configs.GlobalConfig.CodeReview.AutoReviewOnMR {
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

func matchPattern(pattern, value string) bool {
	if pattern == "" || pattern == "*" {
		return true
	}
	if strings.Contains(pattern, "*") {
		regex := "^" + strings.ReplaceAll(regexp.QuoteMeta(pattern), "\\*", ".*") + "$"
		matched, _ := regexp.MatchString(regex, value)
		return matched
	}
	return pattern == value
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
