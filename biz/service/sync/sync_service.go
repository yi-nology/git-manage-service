package sync

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/auth"
	"github.com/yi-nology/git-manage-service/biz/service/commit_analyzer"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	notificationSvc "github.com/yi-nology/git-manage-service/biz/service/notification"
	"github.com/yi-nology/git-manage-service/pkg/lock"
)

type SyncService struct {
	git            *git.GitService
	authSvc        *auth.AuthService
	syncTaskDAO    *db.SyncTaskDAO
	syncRunDAO     *db.SyncRunDAO
	lockSvc        lock.DistLock
	commitAnalyzer *commit_analyzer.AnalyzerService
}

func NewSyncService() *SyncService {
	service := &SyncService{
		git:            git.NewGitService(),
		authSvc:        auth.NewAuthService(),
		syncTaskDAO:    db.NewSyncTaskDAO(),
		syncRunDAO:     db.NewSyncRunDAO(),
		commitAnalyzer: commit_analyzer.NewAnalyzerService(),
	}

	return service
}

// SetLockService 设置锁服务（用于依赖注入）
func (s *SyncService) SetLockService(lockSvc lock.DistLock) {
	s.lockSvc = lockSvc
}

func (s *SyncService) RunTask(taskKey string) error {
	return s.RunTaskWithTrigger(taskKey, po.TriggerSourceManual)
}

func (s *SyncService) RunTaskWithTrigger(taskKey string, triggerSource string) error {
	task, err := s.syncTaskDAO.FindByKey(taskKey)
	if err != nil {
		return err
	}
	return s.ExecuteSyncWithTrigger(task, triggerSource)
}

func (s *SyncService) ExecuteSync(task *po.SyncTask) error {
	return s.ExecuteSyncWithTrigger(task, po.TriggerSourceManual)
}

func (s *SyncService) ExecuteSyncWithTrigger(task *po.SyncTask, triggerSource string) error {
	ctx := context.Background()
	var err error

	// 获取分布式锁保护同步任务
	if s.lockSvc != nil {
		lockKey := fmt.Sprintf("sync:task:%s", task.Key)
		if err := s.lockSvc.UpWait(ctx, lockKey, 5*time.Minute, 30*time.Second); err != nil {
			return fmt.Errorf("failed to acquire lock for task %s: %w", task.Key, err)
		}
		defer func() {
			if err := s.lockSvc.Down(ctx, lockKey); err != nil {
				// 记录解锁失败，但不影响主流程
				_ = err // 暂时使用下划线忽略错误，避免空分支
			}
		}()
	}

	run := po.SyncRun{
		TaskKey:       task.Key,
		TriggerSource: triggerSource,
		StartTime:     time.Now(),
		Status:        "running",
	}
	if err := s.syncRunDAO.Create(&run); err != nil {
		// 记录创建失败，但不影响主流程
		_ = err // 暂时使用下划线忽略错误，避免空分支
	}

	repoPath := task.SourceRepo.Path

	// Capture logs
	var logs strings.Builder
	logf := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		logs.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05"), msg))
	}

	// 根据同步模式选择执行方法
	syncMode := task.SyncMode
	if syncMode == "" {
		syncMode = "single"
	}

	var commitRange string
	if syncMode == "all-branch" {
		commitRange, err = s.doSyncAllBranches(repoPath, task, logf)
	} else {
		commitRange, err = s.doSyncSingleBranch(repoPath, task, logf)
	}

	run.CommitRange = commitRange
	run.Details = logs.String()
	run.EndTime = time.Now()

	if err != nil {
		run.Status = "failed"
		// Check if it was conflict
		if err.Error() == "conflict" {
			run.Status = "conflict"
		}
		run.ErrorMessage = err.Error()
		logf("Sync failed: %v", err)
	} else {
		run.Status = "success"
		logf("Sync completed successfully")
	}
	// Save final details
	run.Details = logs.String()
	if err := s.syncRunDAO.Save(&run); err != nil {
		// 记录保存失败，但不影响主流程
		_ = err // 暂时使用下划线忽略错误，避免空分支
	}

	// 发送通知
	s.sendNotification(task, &run)

	return err
}

// getAuthInfoForRemote 获取指定远程的认证信息（旧系统）
func getAuthInfoForRemote(repo po.Repo, remoteName string) domain.AuthInfo {
	if repo.RemoteAuths != nil {
		if authInfo, ok := repo.RemoteAuths[remoteName]; ok {
			return authInfo
		}
	}
	return domain.AuthInfo{
		Type:   repo.AuthType,
		Key:    repo.AuthKey,
		Secret: repo.AuthSecret,
		Source: "local",
	}
}

// resolveAuthForRemote 解析指定远程的认证方法（支持新凭证系统 + 旧系统回退）
func (s *SyncService) resolveAuthForRemote(repo po.Repo, remoteName string) (transport.AuthMethod, bool, error) {
	return s.authSvc.ResolveCredentialForRemote(
		repo.RemoteCredentials,
		repo.DefaultCredentialID,
		repo.RemoteAuths,
		remoteName,
		repo.AuthType, repo.AuthKey, repo.AuthSecret,
	)
}

// fetchRemote 统一的 fetch 操作，自动处理认证方式选择
func (s *SyncService) fetchRemote(path string, repo po.Repo, remoteName, remoteURL string, refSpecs string, progressWriter io.Writer, logf func(string, ...interface{})) error {
	authMethod, isDBKey, err := s.resolveAuthForRemote(repo, remoteName)
	if err != nil {
		logf("Warning: failed to resolve auth for %s: %v", remoteName, err)
	}

	hasAuth := authMethod != nil || isDBKey

	if remoteURL != "" && hasAuth {
		if isDBKey {
			// 获取凭证 ID 或旧系统的 SSHKeyID
			credID := auth.GetCredentialIDForRemote(repo.RemoteCredentials, repo.DefaultCredentialID, remoteName)
			if credID > 0 {
				privateKey, passphrase, keyErr := s.authSvc.GetCredentialKeyContent(credID)
				if keyErr != nil {
					// 回退到旧系统
					authInfo := getAuthInfoForRemote(repo, remoteName)
					if authInfo.SSHKeyID > 0 {
						privateKey, passphrase, keyErr = s.authSvc.GetDBSSHKeyContent(authInfo.SSHKeyID)
					}
				}
				if keyErr == nil && privateKey != "" {
					logf("Fetching %s using DB SSH key...", remoteName)
					return s.git.FetchWithDBKey(path, remoteURL, privateKey, passphrase, progressWriter, refSpecs)
				}
			}
			// 回退到旧系统
			authInfo := getAuthInfoForRemote(repo, remoteName)
			if authInfo.SSHKeyID > 0 {
				privateKey, passphrase, keyErr := s.authSvc.GetDBSSHKeyContent(authInfo.SSHKeyID)
				if keyErr != nil {
					return fmt.Errorf("failed to load SSH key: %v", keyErr)
				}
				logf("Fetching %s using DB SSH key (legacy)...", remoteName)
				return s.git.FetchWithDBKey(path, remoteURL, privateKey, passphrase, progressWriter, refSpecs)
			}
		}
		logf("Fetching %s with auth...", remoteName)
		return s.git.FetchWithAuthMethod(path, remoteURL, authMethod, progressWriter, refSpecs)
	}

	logf("Fetching %s (no auth)...", remoteName)
	return s.git.Fetch(path, remoteName, progressWriter)
}

// pushRemote 统一的 push 操作，自动处理认证方式选择
func (s *SyncService) pushRemote(path string, repo po.Repo, remoteName, remoteURL, sourceHash, targetBranch string, pushOpts []string, progressWriter io.Writer, logf func(string, ...interface{})) error {
	authMethod, isDBKey, err := s.resolveAuthForRemote(repo, remoteName)
	if err != nil {
		logf("Warning: failed to resolve auth for push to %s: %v", remoteName, err)
	}

	hasAuth := authMethod != nil || isDBKey

	if remoteURL != "" && hasAuth {
		if isDBKey {
			credID := auth.GetCredentialIDForRemote(repo.RemoteCredentials, repo.DefaultCredentialID, remoteName)
			if credID > 0 {
				privateKey, passphrase, keyErr := s.authSvc.GetCredentialKeyContent(credID)
				if keyErr != nil {
					authInfo := getAuthInfoForRemote(repo, remoteName)
					if authInfo.SSHKeyID > 0 {
						privateKey, passphrase, keyErr = s.authSvc.GetDBSSHKeyContent(authInfo.SSHKeyID)
					}
				}
				if keyErr == nil && privateKey != "" {
					logf("Pushing to %s using DB SSH key...", remoteName)
					return s.git.PushWithDBKey(path, remoteURL, sourceHash, targetBranch, privateKey, passphrase, pushOpts, progressWriter)
				}
			}
			authInfo := getAuthInfoForRemote(repo, remoteName)
			if authInfo.SSHKeyID > 0 {
				privateKey, passphrase, keyErr := s.authSvc.GetDBSSHKeyContent(authInfo.SSHKeyID)
				if keyErr != nil {
					return fmt.Errorf("failed to load SSH key for push: %v", keyErr)
				}
				logf("Pushing to %s using DB SSH key (legacy)...", remoteName)
				return s.git.PushWithDBKey(path, remoteURL, sourceHash, targetBranch, privateKey, passphrase, pushOpts, progressWriter)
			}
		}
		logf("Pushing to %s with auth...", remoteName)
		return s.git.PushWithAuthMethod(path, remoteURL, sourceHash, targetBranch, authMethod, pushOpts, progressWriter)
	}

	logf("Pushing to %s (no auth)...", remoteName)
	return s.git.Push(path, remoteName, sourceHash, targetBranch, pushOpts, progressWriter)
}



type logWriter struct {
	logf func(string, ...interface{})
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	str := strings.TrimSpace(string(p))
	if str != "" {
		w.logf("[Git] %s", str)
	}
	return len(p), nil
}

// sendNotification 根据同步结果发送通知
func (s *SyncService) sendNotification(task *po.SyncTask, run *po.SyncRun) {
	var triggerEvent string
	var status string

	switch run.Status {
	case "success":
		triggerEvent = po.TriggerSyncSuccess
		status = "success"
	case "failed":
		triggerEvent = po.TriggerSyncFailure
		status = "failure"
	case "conflict":
		triggerEvent = po.TriggerSyncConflict
		status = "failure"
	default:
		return
	}

	// 计算耗时
	duration := ""
	if !run.EndTime.IsZero() && !run.StartTime.IsZero() {
		d := run.EndTime.Sub(run.StartTime)
		duration = d.Round(time.Millisecond).String()
	}

	data := &notificationSvc.TemplateData{
		TaskKey:      task.Key,
		TaskName:     task.Key,
		Status:       run.Status,
		EventType:    triggerEvent,
		SourceRemote: task.SourceRemote,
		SourceBranch: task.SourceBranch,
		TargetRemote: task.TargetRemote,
		TargetBranch: task.TargetBranch,
		RepoKey:      task.SourceRepoKey,
		RepoName:     task.SourceRepo.Name,
		ErrorMessage: run.ErrorMessage,
		CommitRange:  run.CommitRange,
		Duration:     duration,
		SyncMode:     task.SyncMode,
	}
	if task.Cron != "" {
		data.CronExpression = task.Cron
	}

	// 使用模板渲染的默认标题和内容作为 fallback
	title, content := notificationSvc.RenderTitleAndContent("", "", data)

	notificationSvc.NotifySvc.Send(&notificationSvc.NotificationMessage{
		Title:        title,
		Content:      content,
		Status:       status,
		TriggerEvent: triggerEvent,
		TaskKey:      task.Key,
		RepoKey:      task.SourceRepoKey,
		Data:         data,
	})
}
