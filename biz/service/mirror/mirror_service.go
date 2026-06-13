package mirror

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-manage-service/pkg/lock"
	"github.com/yi-nology/git-manage-service/pkg/queue"
	"github.com/yi-nology/git-platform-sdk/gitbackend"
	"github.com/yi-nology/git-platform-sdk/pkg/branchfilter"
)

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
		if ok, _ := s.lockSvc.Up(ctx, lockKey, 10*time.Minute); !ok {
			return
		}
		defer s.lockSvc.Down(ctx, lockKey)
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

	var execResult interface {
		getBranchesSynced() int
		getCommits() int
	}
	var syncErr error

	switch mirror.MirrorType {
	case po.MirrorTypePull:
		r, err := s.pullExec.Execute(ctx, mirror, logf)
		syncErr = err
		if r != nil {
			syncLog.BranchesSynced = r.BranchesSynced
			syncLog.CommitsPushed = r.CommitsPulled
			_ = execResult
		}
	case po.MirrorTypePush:
		r, err := s.pushExec.Execute(ctx, mirror, logf)
		syncErr = err
		if r != nil {
			syncLog.BranchesSynced = r.BranchesSynced
			syncLog.CommitsPushed = r.CommitsPushed
			_ = execResult
		}
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
	return fmt.Sprintf("mirror_%d_%d", time.Now().UnixNano(), time.Now().Nanosecond())
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

func (s *MirrorService) ValidateCredentialForMirror(mirrorID uint) error {
	mirror, err := s.mirrorDAO.FindByID(mirrorID)
	if err != nil {
		return err
	}

	if mirror.Credential == nil {
		return nil
	}

	_ = branchfilter.New(mirror.BranchFilter)
	return nil
}
