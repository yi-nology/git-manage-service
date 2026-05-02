package codereview

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/provider"
	"github.com/yi-nology/git-manage-service/pkg/configs"
)

func CreateTask(ctx context.Context, repoKey string, providerConfigID uint, mrIID, commitSHA, triggerType string) (*po.ReviewTask, error) {
	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		return nil, fmt.Errorf("repo not found: %w", err)
	}

	if providerConfigID == 0 {
		b, err := resolveBinding(repo.ID)
		if err != nil {
			return nil, fmt.Errorf("no provider binding for repo %s: %w", repoKey, err)
		}
		providerConfigID = b.ProviderConfigID
	}

	task := &po.ReviewTask{
		RepoID:           repo.ID,
		ProviderConfigID: providerConfigID,
		Platform:         "gitlab",
		EventType:        "manual",
		MRIID:            mrIID,
		CommitSHA:        commitSHA,
		TriggerType:      triggerType,
		Status:           "pending",
	}
	if err := db.NewReviewTaskDAO().Create(task); err != nil {
		return nil, fmt.Errorf("failed to create review task: %w", err)
	}

	go func() {
		_ = RunReview(context.Background(), task.ID)
	}()

	return task, nil
}

func RetryTask(ctx context.Context, id uint) (*po.ReviewTask, error) {
	taskDAO := db.NewReviewTaskDAO()
	task, err := taskDAO.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("review task not found: %w", err)
	}

	task.Status = "pending"
	task.ErrorMessage = ""
	task.RiskLevel = ""
	task.Summary = ""
	task.StartedAt = nil
	task.FinishedAt = nil
	if err := taskDAO.Save(task); err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	go func() {
		_ = RunReview(context.Background(), id)
	}()

	return task, nil
}

func RunReview(ctx context.Context, taskID uint) error {
	taskDAO := db.NewReviewTaskDAO()
	task, err := taskDAO.FindByID(taskID)
	if err != nil {
		return err
	}

	now := time.Now()
	task.Status = "running"
	task.StartedAt = &now
	taskDAO.Save(task)

	defer func() {
		finished := time.Now()
		task.FinishedAt = &finished
		taskDAO.Save(task)
	}()

	repo, p, owner, repoName, _, err := resolveRepoProvider(task.RepoID, task.ProviderConfigID)
	if err != nil {
		task.Status = "failed"
		task.ErrorMessage = err.Error()
		taskDAO.Save(task)
		return err
	}

	mrNum, _ := strconv.Atoi(task.MRIID)

	mergeDiff, err := p.GetCRDiff(ctx, owner, repoName, mrNum)
	if err != nil {
		task.Status = "failed"
		task.ErrorMessage = fmt.Sprintf("failed to fetch diff: %v", err)
		taskDAO.Save(task)
		return err
	}

	files := ParseDiff(mergeDiff.RawDiff)

	ruleCtx := &RuleContext{
		Files:    files,
		Provider: task.Platform,
		RepoKey:  repo.Key,
		Owner:    owner,
		Repo:     repoName,
		MRIID:    task.MRIID,
	}

	var allFindings []*Finding
	for _, rule := range GetRules() {
		findings, rErr := rule.Check(ruleCtx)
		if rErr != nil {
			log.Printf("[CodeReview] Rule %s error: %v", rule.ID(), rErr)
			continue
		}
		allFindings = append(allFindings, findings...)
	}

	llmFindings := runLLMReview(ctx, files, repoName, owner)
	allFindings = append(allFindings, llmFindings...)

	totalAdd, totalDel := 0, 0
	for _, f := range files {
		totalAdd += f.Additions
		totalDel += f.Deletions
	}

	cfg := GetMergedConfig(task.RepoID)
	result := Aggregate(allFindings, totalAdd, totalDel, len(files), cfg.BlockOnHigh)

	task.RiskLevel = string(result.RiskLevel)

	if err := persistFindings(taskID, result.Findings); err != nil {
		log.Printf("[CodeReview] Failed to persist findings: %v", err)
	}

	if err := publishComments(ctx, p, owner, repoName, mrNum, task.ID, result); err != nil {
		log.Printf("[CodeReview] Failed to publish comments: %v", err)
	}

	if task.CommitSHA != "" {
		statusState := "success"
		statusDesc := fmt.Sprintf("Review passed (%d findings)", len(result.Findings))
		if result.Blocked {
			statusState = "failed"
			statusDesc = result.BlockReason
		}
		_ = p.CreateCommitStatus(ctx, owner, repoName, task.CommitSHA, provider.CommitStatusOptions{
			State:       statusState,
			Context:     "code-review/git-manage-service",
			Description: statusDesc,
		})
	}

	if err := ApplyPolicy(result, task); err != nil {
		log.Printf("[CodeReview] Failed to apply policy: %v", err)
	}

	task.Summary = BuildSummaryComment(result)
	task.Status = "success"
	if result.Blocked {
		task.Status = "blocked"
	}
	return taskDAO.Save(task)
}

func CheckMerge(ctx context.Context, repoKey, mrIID, commitSHA string) (*api.MergeCheckDTO, error) {
	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		return nil, fmt.Errorf("repo not found: %w", err)
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
	_, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		return "", fmt.Errorf("repo not found: %w", err)
	}
	return "# .cr-service.yaml placeholder\nversion: 1\nreview:\n  enabled: true\n", nil
}

func UpdateReviewConfig(repoKey, configYAML string) error {
	_, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		return fmt.Errorf("repo not found: %w", err)
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

func resolveRepoProvider(repoID, providerConfigID uint) (*po.Repo, provider.Provider, string, string, uint, error) {
	repo, err := db.NewRepoDAO().FindByID(repoID)
	if err != nil {
		return nil, nil, "", "", 0, fmt.Errorf("repo not found: %w", err)
	}

	p, err := provider.GetManager().GetProvider(providerConfigID)
	if err != nil {
		return nil, nil, "", "", 0, fmt.Errorf("provider not found: %w", err)
	}

	owner := repo.PlatformOwner
	repoName := repo.PlatformRepo
	if owner == "" || repoName == "" {
		return nil, nil, "", "", 0, fmt.Errorf("repo %d missing platform owner/repo", repoID)
	}
	return repo, p, owner, repoName, providerConfigID, nil
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
	gc := configs.GlobalConfig.CodeReview
	return reviewConfig{
		BlockOnHigh:    gc.BlockOnHigh,
		MaxFiles:       gc.MaxFiles,
		MaxDiffLines:   gc.MaxDiffLines,
		AutoReviewOnMR: gc.AutoReviewOnMR,
	}
}
