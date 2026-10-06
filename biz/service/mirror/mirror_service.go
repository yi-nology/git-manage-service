package mirror

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-manage-service/pkg/lock"
	"github.com/yi-nology/git-manage-service/pkg/queue"
	"github.com/yi-nology/go-git-platform/gitbackend"
)

// mirrorLockTTL 单次镜像同步的互斥锁时长；ProcessSyncRequest 里的看门狗按
// TTL/3 周期续期，正常退出时由 Down 释放。
const mirrorLockTTL = 10 * time.Minute

var GlobalMirrorService *MirrorService

type MirrorService struct {
	mirrorDAO  *db.MirrorDAO
	syncLogDAO *db.MirrorSyncLogDAO
	lockSvc    lock.DistLock
	pullExec   *PullExecutor
	pushExec   *PushExecutor
	retry      *RetryStrategy
	queue      queue.UniqueQueue
	backend    gitbackend.GitBackend
	cfg        configs.MirrorConfig
}

func NewMirrorService(
	mirrorDAO *db.MirrorDAO,
	syncLogDAO *db.MirrorSyncLogDAO,
	lockSvc lock.DistLock,
	backend gitbackend.GitBackend,
	q queue.UniqueQueue,
	cfg configs.MirrorConfig,
) *MirrorService {
	return &MirrorService{
		mirrorDAO:  mirrorDAO,
		syncLogDAO: syncLogDAO,
		lockSvc:    lockSvc,
		pullExec:   NewPullExecutor(backend),
		pushExec:   NewPushExecutor(backend),
		retry:      NewRetryStrategy(cfg.MaxRetry),
		queue:      q,
		backend:    backend,
		cfg:        cfg,
	}
}

func (s *MirrorService) CreateMirror(mirror *po.Mirror) error {
	if mirror.SyncInterval == 0 {
		mirror.SyncInterval = s.cfg.DefaultSyncInterval
	}
	if mirror.Status == "" {
		mirror.Status = po.MirrorStatusActive
	}
	if mirror.WebhookToken == "" {
		mirror.WebhookToken = generateToken()
	}

	now := time.Now().Add(time.Duration(mirror.SyncInterval) * time.Second)
	mirror.NextSyncAt = &now

	return s.mirrorDAO.Create(mirror)
}

func (s *MirrorService) UpdateMirror(mirror *po.Mirror) error {
	if mirror.SyncInterval > 0 {
		now := time.Now().Add(time.Duration(mirror.SyncInterval) * time.Second)
		mirror.NextSyncAt = &now
		mirror.RetryCount = 0
		if mirror.Status == po.MirrorStatusPaused && mirror.Enabled {
			mirror.Status = po.MirrorStatusActive
		}
	}
	return s.mirrorDAO.Save(mirror)
}

func (s *MirrorService) DeleteMirror(id uint) error {
	return s.mirrorDAO.Delete(id)
}

func (s *MirrorService) GetMirror(id uint) (*po.Mirror, error) {
	return s.mirrorDAO.FindByID(id)
}

func (s *MirrorService) ListMirrors() ([]po.Mirror, error) {
	return s.mirrorDAO.FindAll()
}

func (s *MirrorService) ListMirrorsByRepo(repoID uint) ([]po.Mirror, error) {
	return s.mirrorDAO.FindByRepoID(repoID)
}

func (s *MirrorService) TriggerSync(mirrorID uint, triggerType string) error {
	mirror, err := s.mirrorDAO.FindByID(mirrorID)
	if err != nil {
		return fmt.Errorf("mirror not found: %w", err)
	}

	if !CanStartSync(mirror.Status) {
		return fmt.Errorf("mirror %d is in status %s, cannot start sync", mirrorID, mirror.Status)
	}

	return s.queue.Push(queue.SyncRequest{
		MirrorID:    mirrorID,
		TriggerType: triggerType,
		RequestedAt: time.Now(),
	})
}

func (s *MirrorService) BatchTriggerSync(mirrorIDs []uint, triggerType string) error {
	for _, id := range mirrorIDs {
		if err := s.TriggerSync(id, triggerType); err != nil {
			return err
		}
	}
	return nil
}

func (s *MirrorService) ProcessSyncRequest(req queue.SyncRequest) {
	ctx := context.Background()

	mirror, err := s.mirrorDAO.FindByID(req.MirrorID)
	if err != nil {
		return
	}

	if !CanStartSync(mirror.Status) {
		return
	}

	lockKey := fmt.Sprintf("mirror:sync:%d", mirror.ID)
	if s.lockSvc != nil {
		if ok, _ := s.lockSvc.Up(ctx, lockKey, mirrorLockTTL); !ok {
			return
		}
		defer s.lockSvc.Down(ctx, lockKey)
		// 大仓库同步可能超过单个 TTL；锁在任务结束前过期会让其他实例并发
		// 操作同一个本地仓库。看门狗按 TTL/3 周期续期，Down 前先停。
		stopRenew := make(chan struct{})
		defer close(stopRenew)
		go func() {
			ticker := time.NewTicker(mirrorLockTTL / 3)
			defer ticker.Stop()
			for {
				select {
				case <-stopRenew:
					return
				case <-ticker.C:
					if ok, err := s.lockSvc.Refresh(ctx, lockKey, mirrorLockTTL); err != nil {
						log.Printf("[Mirror] lock renew failed for mirror %d: %v", mirror.ID, err)
					} else if !ok {
						log.Printf("[Mirror] lock for mirror %d was lost, sync may run concurrently", mirror.ID)
					}
				}
			}
		}()
	}

	status := NewMirrorStatus(mirror.Status)
	if err := status.TransitionTo(po.MirrorStatusSyncing); err != nil {
		return
	}
	s.mirrorDAO.UpdateStatus(mirror.ID, po.MirrorStatusSyncing)

	syncLog := &po.MirrorSyncLog{
		MirrorID:    mirror.ID,
		TriggerType: req.TriggerType,
		Status:      po.SyncLogStatusRunning,
		StartedAt:   timePtr(time.Now()),
	}
	s.syncLogDAO.Create(syncLog)

	var logs strings.Builder
	logf := func(format string, args ...interface{}) {
		msg := fmt.Sprintf(format, args...)
		logs.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05"), msg))
	}

	var syncErr error

	switch mirror.MirrorType {
	case po.MirrorTypePull:
		r, err := s.pullExec.Execute(ctx, mirror, logf)
		syncErr = err
		if r != nil {
			syncLog.BranchesSynced = r.BranchesSynced
			syncLog.CommitsPushed = r.CommitsPulled
		}
	case po.MirrorTypePush:
		r, err := s.pushExec.Execute(ctx, mirror, logf)
		syncErr = err
		if r != nil {
			syncLog.BranchesSynced = r.BranchesSynced
			syncLog.CommitsPushed = r.CommitsPushed
		}
	default:
		// 未知类型不能标记成功，否则会错误地重置重试计数、掩盖故障
		syncErr = fmt.Errorf("unknown mirror type: %s", mirror.MirrorType)
	}

	now := time.Now()
	syncLog.FinishedAt = &now
	syncLog.DurationMs = now.Sub(*syncLog.StartedAt).Milliseconds()
	syncLog.DetailLog = logs.String()

	if syncErr != nil {
		syncLog.Status = po.SyncLogStatusFailed
		syncLog.ErrorMessage = syncErr.Error()
		s.handleSyncFailure(mirror)
	} else {
		syncLog.Status = po.SyncLogStatusSuccess
		s.handleSyncSuccess(mirror)
	}

	s.syncLogDAO.Save(syncLog)
}

func (s *MirrorService) handleSyncSuccess(mirror *po.Mirror) {
	now := time.Now()

	s.mirrorDAO.UpdateStatus(mirror.ID, po.MirrorStatusActive)
	s.mirrorDAO.ResetRetryCount(mirror.ID)

	updated, err := s.mirrorDAO.FindByID(mirror.ID)
	if err == nil {
		updated.LastSyncAt = &now
		updated.LastError = ""
		nextSync := now.Add(time.Duration(updated.SyncInterval) * time.Second)
		updated.NextSyncAt = &nextSync
		s.mirrorDAO.Save(updated)
	}
}

func (s *MirrorService) handleSyncFailure(mirror *po.Mirror) {
	newRetryCount := mirror.RetryCount + 1
	s.mirrorDAO.IncrementRetryCount(mirror.ID)

	if s.retry.ShouldPause(newRetryCount) {
		s.mirrorDAO.UpdateStatus(mirror.ID, po.MirrorStatusPaused)
		return
	}

	nextSync := s.retry.GetNextSyncAt(newRetryCount)
	s.mirrorDAO.UpdateStatus(mirror.ID, po.MirrorStatusFailed)
	s.mirrorDAO.UpdateNextSyncAt(mirror.ID, nextSync)
}

func (s *MirrorService) PauseMirror(id uint) error {
	return s.mirrorDAO.UpdateStatus(id, po.MirrorStatusPaused)
}

func (s *MirrorService) ResumeMirror(id uint) error {
	mirror, err := s.mirrorDAO.FindByID(id)
	if err != nil {
		return err
	}
	s.mirrorDAO.ResetRetryCount(id)
	s.mirrorDAO.UpdateStatus(id, po.MirrorStatusActive)

	now := time.Now().Add(time.Duration(mirror.SyncInterval) * time.Second)
	s.mirrorDAO.UpdateNextSyncAt(id, now)

	return nil
}

func (s *MirrorService) ListSyncLogs(mirrorID uint, limit int) ([]po.MirrorSyncLog, error) {
	return s.syncLogDAO.FindByMirrorID(mirrorID, limit)
}

func (s *MirrorService) GetSyncLog(id uint) (*po.MirrorSyncLog, error) {
	return s.syncLogDAO.FindByID(id)
}

func (s *MirrorService) DeleteSyncLog(id uint) error {
	return s.syncLogDAO.Delete(id)
}

func (s *MirrorService) PreviewSync(ctx context.Context, id uint) (string, error) {
	mirror, err := s.mirrorDAO.FindByID(id)
	if err != nil {
		return "", fmt.Errorf("mirror not found: %w", err)
	}

	switch mirror.MirrorType {
	case po.MirrorTypePull:
		return PreviewPull(ctx, s.backend, mirror)
	case po.MirrorTypePush:
		return PreviewPush(ctx, s.backend, mirror)
	default:
		return "", fmt.Errorf("unknown mirror type: %s", mirror.MirrorType)
	}
}

func (s *MirrorService) GetMirrorByWebhookToken(token string) (*po.Mirror, error) {
	return s.mirrorDAO.FindByWebhookToken(token)
}

func (s *MirrorService) HandleWebhook(token string) error {
	mirror, err := s.mirrorDAO.FindByWebhookToken(token)
	if err != nil {
		return fmt.Errorf("invalid webhook token")
	}
	return s.TriggerSync(mirror.ID, po.TriggerTypeWebhook)
}

func (s *MirrorService) CleanupOldLogs() error {
	return s.mirrorDAO.CleanupOldLogs(s.cfg.LogRetentionDays)
}

func (s *MirrorService) AnalyzeRemote(ctx context.Context, remoteURL string) (*AnalyzeResult, error) {
	return analyzeRemote(ctx, s.backend, remoteURL)
}

type AnalyzeResult struct {
	Reachable     bool     `json:"reachable"`
	Branches      []string `json:"branches"`
	DefaultBranch string   `json:"defaultBranch"`
	Protocol      string   `json:"protocol"`
}

func timePtr(t time.Time) *time.Time {
	return &t
}

func generateToken() string {
	return uuid.NewString()
}

func analyzeRemote(ctx context.Context, backend gitbackend.GitBackend, remoteURL string) (*AnalyzeResult, error) {
	result := &AnalyzeResult{
		Reachable: false,
		Protocol:  "unknown",
	}

	if strings.HasPrefix(remoteURL, "git@") || strings.HasPrefix(remoteURL, "ssh://") {
		result.Protocol = "ssh"
	} else if strings.HasPrefix(remoteURL, "https://") {
		result.Protocol = "https"
	} else if strings.HasPrefix(remoteURL, "http://") {
		result.Protocol = "http"
	}

	return result, nil
}
