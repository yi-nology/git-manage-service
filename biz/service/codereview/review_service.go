package codereview

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/notification"
	"github.com/yi-nology/git-manage-service/biz/service/provider_manager"
	"github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-manage-service/pkg/logger"
	servicePkg "github.com/yi-nology/git-manage-service/pkg/service"
	"github.com/yi-nology/git-platform-sdk/provider"
)

var runningTasks sync.Map

func runReviewAsync(taskID uint) {
	go func() {
		_ = RunReview(context.Background(), taskID)
	}()
}

type reviewParams struct {
	p          provider.Provider
	owner      string
	repo       string
	repoKey    string
	repoID     uint
	providerID uint
}

func CreateTask(ctx context.Context, repoKey string, providerConfigID uint, mrIID, commitSHA, triggerType string) (*po.ReviewTask, error) {
	repo, err := servicePkg.GetRepoByKey(repoKey)
	if err != nil {
		return nil, err
	}

	if providerConfigID == 0 {
		b, err := resolveBinding(repo.ID)
		if err != nil {
			return nil, fmt.Errorf("no provider binding for repo %s: %w", repoKey, err)
		}
		providerConfigID = b.ProviderConfigID
	}

	cfg := GetMergedConfig(repo.ID)
	cfgSnapshot, _ := json.Marshal(cfg)

	task := &po.ReviewTask{
		RepoID:           repo.ID,
		ProviderConfigID: providerConfigID,
		Platform:         "gitlab",
		EventType:        "manual",
		MRIID:            mrIID,
		CommitSHA:        commitSHA,
		TriggerType:      triggerType,
		Status:           "pending",
		ConfigSnapshot:   string(cfgSnapshot),
	}
	if err := db.NewReviewTaskDAO().Create(task); err != nil {
		return nil, fmt.Errorf("failed to create review task: %w", err)
	}

	runReviewAsync(task.ID)

	return task, nil
}

func CreateTaskByProvider(ctx context.Context, providerConfigID uint, owner, repo, mrIID, commitSHA, triggerType string) (*po.ReviewTask, error) {
	if providerConfigID == 0 {
		return nil, fmt.Errorf("provider_config_id is required")
	}

	cfg := getConfig()
	cfgSnapshot, _ := json.Marshal(cfg)

	task := &po.ReviewTask{
		ProviderConfigID: providerConfigID,
		Platform:         "remote",
		PlatformOwner:    owner,
		PlatformRepo:     repo,
		EventType:        "manual",
		MRIID:            mrIID,
		CommitSHA:        commitSHA,
		TriggerType:      triggerType,
		Status:           "pending",
		ConfigSnapshot:   string(cfgSnapshot),
	}
	if err := db.NewReviewTaskDAO().Create(task); err != nil {
		return nil, fmt.Errorf("failed to create review task: %w", err)
	}

	runReviewAsync(task.ID)

	return task, nil
}

func RetryTask(ctx context.Context, id uint, owner, repo string) (*po.ReviewTask, error) {
	taskDAO := db.NewReviewTaskDAO()
	task, err := taskDAO.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("review task not found: %w", err)
	}

	if task.Status == "running" {
		if task.StartedAt != nil && time.Since(*task.StartedAt) < 2*time.Minute {
			return nil, fmt.Errorf("task %d is already running", id)
		}
		logger.Warn("Force-retrying stuck running task", logrus.Fields{
			"task_id": id, "started_at": task.StartedAt,
		})
	}

	task.Status = "pending"
	task.ErrorMessage = ""
	task.RiskLevel = ""
	task.Summary = ""
	task.StartedAt = nil
	task.FinishedAt = nil

	if owner != "" && task.PlatformOwner == "" {
		task.PlatformOwner = owner
	}
	if repo != "" && task.PlatformRepo == "" {
		task.PlatformRepo = repo
	}

	if err := taskDAO.Save(task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	runReviewAsync(id)

	return task, nil
}

func RunReview(ctx context.Context, taskID uint) (retErr error) {
	taskKey := strconv.FormatUint(uint64(taskID), 10)
	if _, loaded := runningTasks.LoadOrStore(taskKey, struct{}{}); loaded {
		return fmt.Errorf("review task %d is already running", taskID)
	}
	defer runningTasks.Delete(taskKey)

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	taskDAO := db.NewReviewTaskDAO()
	task, err := taskDAO.FindByID(taskID)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			logger.ErrorWithErr("RunReview panic recovered", fmt.Errorf("%v", r), logrus.Fields{"task_id": taskID})
			task.Status = "failed"
			task.ErrorMessage = fmt.Sprintf("internal error: %v", r)
			taskDAO.Save(task)
			retErr = fmt.Errorf("panic: %v", r)
		}
	}()

	now := time.Now()
	task.Status = "running"
	task.StartedAt = &now
	taskDAO.Save(task)

	repoKey := ""
	if task.RepoID > 0 {
		if r, rErr := db.NewRepoDAO().FindByID(task.RepoID); rErr == nil {
			repoKey = r.Key
		}
	}
	broadcastReviewStatus(task, repoKey, "running")

	failTask := func(err error) error {
		task.Status = "failed"
		task.ErrorMessage = err.Error()
		taskDAO.Save(task)
		broadcastReviewStatus(task, repoKey, "failed")
		return err
	}

	params, err := resolveReviewParams(task)
	if err != nil {
		return failTask(err)
	}

	result, rawDiff, err := executeReview(ctx, task, params, taskDAO)
	if err != nil {
		return failTask(err)
	}

	finalizeReview(ctx, task, result, rawDiff, params, taskDAO, repoKey)
	return nil
}

func resolveReviewParams(task *po.ReviewTask) (*reviewParams, error) {
	if task.RepoID == 0 && task.ProviderConfigID == 0 {
		return nil, fmt.Errorf("task %d has no repo_id and no provider_config_id", task.ID)
	}

	var owner, name, repoKey string
	var repoID uint
	if task.RepoID > 0 {
		repo, err := db.NewRepoDAO().FindByID(task.RepoID)
		if err != nil {
			return nil, fmt.Errorf("repo not found: %w", err)
		}
		if repo.PlatformOwner == "" || repo.PlatformRepo == "" {
			return nil, fmt.Errorf("repo %d missing platform owner/repo", task.RepoID)
		}
		owner, name, repoKey, repoID = repo.PlatformOwner, repo.PlatformRepo, repo.Key, repo.ID
	} else {
		owner, name = task.PlatformOwner, task.PlatformRepo
	}

	p, err := provider_manager.GetManager().GetProvider(task.ProviderConfigID)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}
	return &reviewParams{
		p: p, owner: owner, repo: name,
		repoKey: repoKey, repoID: repoID, providerID: task.ProviderConfigID,
	}, nil
}

func executeReview(ctx context.Context, task *po.ReviewTask, params *reviewParams, taskDAO *db.ReviewTaskDAO) (*AggregatedResult, string, error) {
	var processLog []*ProcessStep
	addStep := func(name, status, detail string) {
		processLog = append(processLog, &ProcessStep{Name: name, Status: status, Detail: detail})
	}

	addStep("Resolve Provider", "ok", fmt.Sprintf("owner=%s, repo=%s", params.owner, params.repo))

	mergeDiff, err := params.p.GetCRDiff(ctx, params.owner, params.repo, task.MRIID)
	if err != nil {
		task.Status = "failed"
		task.ErrorMessage = fmt.Sprintf("failed to fetch diff: %v", err)
		taskDAO.Save(task)
		return nil, "", err
	}
	rawDiff := mergeDiff.RawDiff

	addStep("Fetch Diff", "ok", fmt.Sprintf("fetched %d changed files from remote, RawDiff length=%d", len(mergeDiff.Files), len(mergeDiff.RawDiff)))

	files := ParseDiff(mergeDiff.RawDiff)
	if len(files) == 0 && len(mergeDiff.Files) > 0 {
		addStep("Parse Diff", "warn", fmt.Sprintf("ParseDiff(RawDiff) returned 0 files, but provider reported %d files; RawDiff may be empty or unsupported format", len(mergeDiff.Files)))
	} else {
		addStep("Parse Diff", "ok", fmt.Sprintf("parsed %d files from diff", len(files)))
	}

	ruleCtx := &RuleContext{
		Files:    files,
		Provider: task.Platform,
		RepoKey:  params.repoKey,
		Owner:    params.owner,
		Repo:     params.repo,
		MRIID:    task.MRIID,
	}

	var allFindings []*Finding
	enabledIDs, _ := db.NewReviewRuleDAO().FindEnabledIDs()
	for _, rule := range GetRules() {
		name := fmt.Sprintf("Rule: %s", rule.ID())
		if enabledIDs != nil && !enabledIDs[rule.ID()] {
			addStep(name, "skip", "disabled in settings")
			continue
		}
		findings, rErr := rule.Check(ruleCtx)
		if rErr != nil {
			logger.ErrorWithErr("Rule check failed", rErr, logrus.Fields{"rule": rule.ID()})
			addStep(name, "error", rErr.Error())
			continue
		}
		addStep(name, "ok", fmt.Sprintf("%d findings", len(findings)))
		allFindings = append(allFindings, findings...)
	}

	cfg := GetMergedConfig(params.repoID)

	var llmFindings []*Finding
	if len(files) == 0 {
		addStep("LLM Review", "skip", "skipped: no diff files to review")
	} else {
		var llmStep *ProcessStep
		llmFindings, llmStep = runLLMReview(ctx, files, params.repo, params.owner, cfg.LLMProvider, params.repoID, resolveRepoConfig(params))
		processLog = append(processLog, llmStep)
	}
	allFindings = append(allFindings, llmFindings...)

	result := Aggregate(allFindings, mergeDiff.TotalAdd, mergeDiff.TotalDel, len(files), cfg.BlockOnHigh, processLog)
	return result, rawDiff, nil
}

func finalizeReview(ctx context.Context, task *po.ReviewTask, result *AggregatedResult, rawDiff string, params *reviewParams, taskDAO *db.ReviewTaskDAO, repoKey string) {
	task.RiskLevel = string(result.RiskLevel)

	findingIDMap, err := persistFindings(task.ID, result.Findings)
	if err != nil {
		logger.ErrorWithErr("Failed to persist findings", err, logrus.Fields{"task_id": task.ID})
	}

	if task.MRIID != "" {
		cleanupOldComments(ctx, params.p, params.owner, params.repo, task.MRIID, task.ProviderConfigID, task.MRIID)
		if pErr := publishComments(ctx, params.p, params.owner, params.repo, task.MRIID, task.ID, task.CommitSHA, result, findingIDMap); pErr != nil {
			logger.ErrorWithErr("Failed to publish comments", pErr, logrus.Fields{"task_id": task.ID})
		}
	}

	if csm, ok := params.p.(provider.CommitStatusManager); ok && task.CommitSHA != "" {
		statusState := "success"
		statusDesc := fmt.Sprintf("Review passed (%d findings)", len(result.Findings))
		if result.Blocked {
			statusState = "failed"
			statusDesc = result.BlockReason
		}
		_ = csm.CreateCommitStatus(ctx, params.owner, params.repo, task.CommitSHA, provider.CommitStatusOptions{
			State:       statusState,
			Context:     "code-review/git-manage-service/" + params.repoKey,
			Description: statusDesc,
		})
	}

	if err := ApplyPolicy(result, task); err != nil {
		logger.ErrorWithErr("Failed to apply policy", err, logrus.Fields{"task_id": task.ID})
	}

	task.Summary = BuildSummaryComment(result)
	task.RawDiff = rawDiff
	if logJSON, jErr := json.Marshal(result.ProcessLog); jErr == nil {
		task.ProcessLog = string(logJSON)
	}
	task.Status = "success"
	if result.Blocked {
		task.Status = "blocked"
	}

	finished := time.Now()
	task.FinishedAt = &finished
	if err := taskDAO.Save(task); err != nil {
		logger.ErrorWithErr("Failed to save final task state", err, logrus.Fields{"task_id": task.ID})
	}

	sendReviewNotification(task, result, repoKey)
	broadcastReviewStatus(task, repoKey, task.Status)
}

func sendReviewNotification(task *po.ReviewTask, result *AggregatedResult, repoKey string) {
	status := "success"
	if task.Status == "blocked" || task.Status == "failed" {
		status = "failure"
	}
	notification.NotifySvc.Send(&notification.NotificationMessage{
		Title:        fmt.Sprintf("代码审查完成 - MR !%s - %s", task.MRIID, task.RiskLevel),
		Content:      fmt.Sprintf("风险等级: %s, 发现问题: %d", task.RiskLevel, len(result.Findings)),
		Status:       status,
		TriggerEvent: "code_review_completed",
		RepoKey:      repoKey,
	})
}

func broadcastReviewStatus(task *po.ReviewTask, repoKey, status string) {
	// WebSocket push was removed (the ws server was never started, so this was a
	// no-op). Call sites are retained as status-change hooks for a future
	// real-time implementation.
}

func CheckMerge(ctx context.Context, repoKey, mrIID, commitSHA string) (*api.MergeCheckDTO, error) {
	repo, err := servicePkg.GetRepoByKey(repoKey)
	if err != nil {
		return nil, err
	}

	checkDAO := db.NewMergeCheckResultDAO()
	result, err := checkDAO.FindLatest(repo.ID, mrIID, commitSHA)
	if err != nil {
		return &api.MergeCheckDTO{Mergeable: true, Checks: []api.MergeCheckItemDTO{}}, nil
	}

	checks := []api.MergeCheckItemDTO{{
		CheckType: result.CheckType,
		Status:    result.Status,
		RiskLevel: result.RiskLevel,
		Message:   result.Message,
	}}

	mergeable := result.Status == "success"
	return &api.MergeCheckDTO{Mergeable: mergeable, Checks: checks}, nil
}

func GetReviewConfig(repoKey string) (string, error) {
	_, err := servicePkg.GetRepoByKey(repoKey)
	if err != nil {
		return "", err
	}
	return "# .cr-service.yaml placeholder\nversion: 1\nreview:\n  enabled: true\n", nil
}

func UpdateReviewConfig(repoKey, configYAML string) error {
	_, err := servicePkg.GetRepoByKey(repoKey)
	if err != nil {
		return err
	}
	return nil
}

func GetMergedConfig(repoID uint) reviewConfig {
	repo, err := db.NewRepoDAO().FindByID(repoID)
	if err != nil {
		return getConfig()
	}

	cfg := getConfig()

	if repo.ProviderConfigID == 0 || repo.PlatformOwner == "" || repo.PlatformRepo == "" {
		return cfg
	}

	repoCfg, err := db.NewReviewRepoConfigDAO().FindByRemoteRepo(repo.ProviderConfigID, repo.PlatformOwner, repo.PlatformRepo)
	if err != nil {
		return cfg
	}

	if !repoCfg.Enabled {
		cfg.BlockOnHigh = false
		cfg.AutoReviewOnMR = false
		return cfg
	}

	cfg.BlockOnHigh = repoCfg.BlockOnHigh
	cfg.AutoReviewOnMR = repoCfg.AutoReviewOnMR
	if repoCfg.MaxFiles > 0 {
		cfg.MaxFiles = repoCfg.MaxFiles
	}
	if repoCfg.MaxDiffLines > 0 {
		cfg.MaxDiffLines = repoCfg.MaxDiffLines
	}
	if repoCfg.LLMProvider != "" {
		cfg.LLMProvider = repoCfg.LLMProvider
	}
	return cfg
}

func resolveBinding(repoID uint) (*po.RepoProviderBinding, error) {
	bindingDAO := db.NewRepoProviderBindingDAO()
	b, err := bindingDAO.FindPrimaryByRepoID(repoID)
	if err == nil && b != nil {
		return b, nil
	}

	repo, err := db.NewRepoDAO().FindByID(repoID)
	if err != nil {
		return nil, err
	}
	if repo.ProviderConfigID == 0 {
		return nil, fmt.Errorf("no provider binding for repo %d", repoID)
	}
	return &po.RepoProviderBinding{ProviderConfigID: repo.ProviderConfigID}, nil
}

type reviewConfig struct {
	BlockOnHigh    bool
	MaxFiles       int
	MaxDiffLines   int
	LLMProvider    string
	AutoReviewOnMR bool
}

func getConfig() reviewConfig {
	gc := configs.GetCodeReviewConfig()
	return reviewConfig{
		BlockOnHigh:    gc.BlockOnHigh,
		MaxFiles:       gc.MaxFiles,
		MaxDiffLines:   gc.MaxDiffLines,
		AutoReviewOnMR: gc.AutoReviewOnMR,
	}
}

func resolveRepoConfig(params *reviewParams) *po.ReviewRepoConfig {
	if params.providerID == 0 || params.owner == "" || params.repo == "" {
		return nil
	}
	cfg, err := db.NewReviewRepoConfigDAO().FindByRemoteRepo(params.providerID, params.owner, params.repo)
	if err != nil {
		return nil
	}
	return cfg
}
