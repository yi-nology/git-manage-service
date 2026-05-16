package adapter

import (
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

// GitSyncTask 内部定义的同步任务模型（与 git-sync-service 兼容）
type GitSyncTask struct {
	Key            string
	SourceRepoKey  string
	SourceRemote   string
	SourceBranch   string
	TargetRepoKey  string
	TargetRemote   string
	TargetBranch   string
	CronExpression string
	SyncMode       string
	Enabled        bool
	PushOptions    string
	GitTags        bool
	GitForce       bool
	GitPrune       bool
	GitNoVerify    bool
}

// GitSyncRun 内部定义的同步运行模型（与 git-sync-service 兼容）
type GitSyncRun struct {
	ID            uint64
	TaskKey       string
	TriggerSource string
	Status        string
	CommitRange   string
	ErrorMessage  string
	Details       string
	StartTime     time.Time
	EndTime       time.Time
	CreatedAt     time.Time
}

// GitSyncRepo 内部定义的仓库模型（与 git-sync-service 兼容）
type GitSyncRepo struct {
	Key         string
	Name        string
	Platform    string
	CloneURL    string
	AccessToken string
	Path        string
}

// TriggerSource 触发源类型
type TriggerSource string

const (
	TriggerSourceManual  TriggerSource = "manual"
	TriggerSourceCron    TriggerSource = "cron"
	TriggerSourceWebhook TriggerSource = "webhook"
)

// SyncStatus 同步状态
type SyncStatus string

const (
	SyncStatusRunning  SyncStatus = "running"
	SyncStatusSuccess  SyncStatus = "success"
	SyncStatusFailed   SyncStatus = "failed"
	SyncStatusConflict SyncStatus = "conflict"
)

// ToGitSyncTask 转换为兼容格式
func ToGitSyncTask(task *po.SyncTask) *GitSyncTask {
	if task == nil {
		return nil
	}

	return &GitSyncTask{
		Key:            task.Key,
		SourceRepoKey:  task.SourceRepoKey,
		SourceRemote:   task.SourceRemote,
		SourceBranch:   task.SourceBranch,
		TargetRepoKey:  task.TargetRepoKey,
		TargetRemote:   task.TargetRemote,
		TargetBranch:   task.TargetBranch,
		CronExpression: task.Cron,
		SyncMode:       convertSyncMode(task.SyncMode),
		Enabled:        task.Enabled,
		PushOptions:    task.PushOptions,
		GitTags:        task.GitTags,
		GitForce:       task.GitForce,
		GitPrune:       task.GitPrune,
		GitNoVerify:    task.GitNoVerify,
	}
}

// FromGitSyncTask 从兼容格式转换回来
func FromGitSyncTask(task *GitSyncTask) *po.SyncTask {
	if task == nil {
		return nil
	}

	return &po.SyncTask{
		Key:           task.Key,
		SourceRepoKey: task.SourceRepoKey,
		SourceRemote:  task.SourceRemote,
		SourceBranch:  task.SourceBranch,
		TargetRepoKey: task.TargetRepoKey,
		TargetRemote:  task.TargetRemote,
		TargetBranch:  task.TargetBranch,
		Cron:          task.CronExpression,
		SyncMode:      task.SyncMode,
		Enabled:       task.Enabled,
		PushOptions:   task.PushOptions,
		GitTags:       task.GitTags,
		GitForce:      task.GitForce,
		GitPrune:      task.GitPrune,
		GitNoVerify:   task.GitNoVerify,
	}
}

// ToGitSyncRun 转换为兼容格式
func ToGitSyncRun(run *po.SyncRun) *GitSyncRun {
	if run == nil {
		return nil
	}

	return &GitSyncRun{
		ID:            uint64(run.ID),
		TaskKey:       run.TaskKey,
		TriggerSource: run.TriggerSource,
		Status:        run.Status,
		CommitRange:   run.CommitRange,
		ErrorMessage:  run.ErrorMessage,
		Details:       run.Details,
		StartTime:     run.StartTime,
		EndTime:       run.EndTime,
		CreatedAt:     run.CreatedAt,
	}
}

// FromGitSyncRun 从兼容格式转换回来
func FromGitSyncRun(run *GitSyncRun) *po.SyncRun {
	if run == nil {
		return nil
	}

	return &po.SyncRun{
		TaskKey:       run.TaskKey,
		TriggerSource: run.TriggerSource,
		Status:        run.Status,
		CommitRange:   run.CommitRange,
		ErrorMessage:  run.ErrorMessage,
		Details:       run.Details,
		StartTime:     run.StartTime,
		EndTime:       run.EndTime,
	}
}

// ToGitSyncRepo 转换为兼容格式
func ToGitSyncRepo(repo *po.Repo) *GitSyncRepo {
	if repo == nil {
		return nil
	}

	return &GitSyncRepo{
		Key:         repo.Key,
		Name:        repo.Name,
		Platform:    detectPlatform(repo.RemoteURL),
		CloneURL:    repo.RemoteURL,
		AccessToken: getRepoAccessToken(repo),
		Path:        repo.Path,
	}
}

// ConvertTriggerSource 转换触发源
func ConvertTriggerSource(source string) TriggerSource {
	switch source {
	case po.TriggerSourceManual:
		return TriggerSourceManual
	case po.TriggerSourceCron:
		return TriggerSourceCron
	case po.TriggerSourceWebhook:
		return TriggerSourceWebhook
	default:
		return TriggerSourceManual
	}
}

// ConvertSyncStatus 转换同步状态
func ConvertSyncStatus(status string) string {
	switch SyncStatus(status) {
	case SyncStatusRunning:
		return "running"
	case SyncStatusSuccess:
		return "success"
	case SyncStatusFailed:
		return "failed"
	case SyncStatusConflict:
		return "conflict"
	default:
		return status
	}
}

// SyncPreviewResult 预览结果转换
type SyncPreviewResult struct {
	NeedSync      bool
	IsFastForward bool
	SourceHash    string
	TargetHash    string
	CommitCount   int
	CommitRange   string
}

func convertSyncMode(mode string) string {
	switch mode {
	case "single":
		return "single"
	case "all-branch":
		return "all-branch"
	default:
		return "single"
	}
}

func detectPlatform(cloneURL string) string {
	if cloneURL == "" {
		return "generic"
	}
	if strings.Contains(cloneURL, "github.com") {
		return "github"
	}
	if strings.Contains(cloneURL, "gitlab.com") || strings.Contains(cloneURL, "gitlab.") {
		return "gitlab"
	}
	if strings.Contains(cloneURL, "gitea.com") || strings.Contains(cloneURL, "gitea.") {
		return "gitea"
	}
	return "generic"
}

func getRepoAccessToken(repo *po.Repo) string {
	if repo.AuthType == "token" {
		return repo.AuthSecret
	}
	return ""
}
