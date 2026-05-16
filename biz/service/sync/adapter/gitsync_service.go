package adapter

import (
	"context"
	"fmt"
	"sync"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

var (
	gitSyncService *GitSyncService
	once           sync.Once
)

// SyncService 定义 git-sync-service 核心服务的接口
// 当实际集成时，由 git-sync-service 包提供实现
type SyncService interface {
	RunTaskWithTrigger(ctx context.Context, taskKey string, triggerSource TriggerSource) (*GitSyncRun, error)
	ExecuteSync(ctx context.Context, task *GitSyncTask) (*GitSyncRun, error)
	PreviewSync(ctx context.Context, task *GitSyncTask) (*GitSyncPreviewResult, error)
	BatchSync(ctx context.Context, taskKeys []string) ([]*GitSyncRun, error)
	RegisterRepo(ctx context.Context, repo *GitSyncRepo) error
	UnregisterRepo(ctx context.Context, repoKey string) error
	ReceiveWebhook(ctx context.Context, repoKey string, payload []byte, headers map[string]string) error
	Start()
	Stop()
}

// GitSyncPreviewResult 预览结果
type GitSyncPreviewResult struct {
	NeedSync      bool
	IsFastForward bool
	SourceHash    string
	TargetHash    string
	CommitCount   int
	CommitRange   string
}

// GitSyncService git-sync-service 包装服务
type GitSyncService struct {
	coreService SyncService
	initialized bool
	mu          sync.RWMutex
}

// GetGitSyncService 获取单例实例
func GetGitSyncService() *GitSyncService {
	once.Do(func() {
		gitSyncService = &GitSyncService{}
	})
	return gitSyncService
}

// Initialize 初始化 git-sync-service 核心服务
// 当集成真实的 git-sync-service 时，在此处创建并赋值 coreService
func (s *GitSyncService) Initialize(cfg *Config) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.initialized {
		return nil
	}

	// TODO: 在此处初始化真实的 git-sync-service
	// 示例:
	// svc, err := gitsync.NewService(gitsync.Config{...})
	// if err != nil { return err }
	// s.coreService = svc

	s.initialized = true
	return fmt.Errorf("git-sync-service core implementation pending integration")
}

// SetCoreService 手动设置核心服务实现（用于依赖注入/测试）
func (s *GitSyncService) SetCoreService(svc SyncService) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.coreService = svc
	s.initialized = (svc != nil)
}

// Start 启动后台服务
func (s *GitSyncService) Start() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.coreService != nil {
		s.coreService.Start()
	}
}

// Stop 停止后台服务
func (s *GitSyncService) Stop() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.coreService != nil {
		s.coreService.Stop()
	}
}

// RunTask 执行同步任务
func (s *GitSyncService) RunTask(ctx context.Context, taskKey string, triggerSource string) (*po.SyncRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.coreService == nil {
		return nil, fmt.Errorf("git-sync-service not initialized")
	}

	gsTrigger := ConvertTriggerSource(triggerSource)

	run, err := s.coreService.RunTaskWithTrigger(ctx, taskKey, gsTrigger)
	if err != nil {
		return nil, err
	}

	return FromGitSyncRun(run), nil
}

// ExecuteSync 直接执行同步
func (s *GitSyncService) ExecuteSync(ctx context.Context, task *po.SyncTask) (*po.SyncRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.coreService == nil {
		return nil, fmt.Errorf("git-sync-service not initialized")
	}

	gsTask := ToGitSyncTask(task)

	run, err := s.coreService.ExecuteSync(ctx, gsTask)
	if err != nil {
		return nil, err
	}

	return FromGitSyncRun(run), nil
}

// PreviewSync 预览同步
func (s *GitSyncService) PreviewSync(ctx context.Context, task *po.SyncTask) (*SyncPreviewResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.coreService == nil {
		return nil, fmt.Errorf("git-sync-service not initialized")
	}

	gsTask := ToGitSyncTask(task)

	preview, err := s.coreService.PreviewSync(ctx, gsTask)
	if err != nil {
		return nil, err
	}

	return &SyncPreviewResult{
		NeedSync:      preview.NeedSync,
		IsFastForward: preview.IsFastForward,
		SourceHash:    preview.SourceHash,
		TargetHash:    preview.TargetHash,
		CommitCount:   preview.CommitCount,
		CommitRange:   preview.CommitRange,
	}, nil
}

// BatchSync 批量同步
func (s *GitSyncService) BatchSync(ctx context.Context, taskKeys []string) ([]*po.SyncRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.coreService == nil {
		return nil, fmt.Errorf("git-sync-service not initialized")
	}

	runs, err := s.coreService.BatchSync(ctx, taskKeys)
	if err != nil {
		return nil, err
	}

	result := make([]*po.SyncRun, len(runs))
	for i, run := range runs {
		result[i] = FromGitSyncRun(run)
	}

	return result, nil
}

// RegisterRepo 注册仓库
func (s *GitSyncService) RegisterRepo(ctx context.Context, repo *po.Repo) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.coreService == nil {
		return fmt.Errorf("git-sync-service not initialized")
	}

	gsRepo := ToGitSyncRepo(repo)
	return s.coreService.RegisterRepo(ctx, gsRepo)
}

// UnregisterRepo 注销仓库
func (s *GitSyncService) UnregisterRepo(ctx context.Context, repoKey string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.coreService == nil {
		return fmt.Errorf("git-sync-service not initialized")
	}

	return s.coreService.UnregisterRepo(ctx, repoKey)
}

// ReceiveWebhook 接收 Webhook 事件
func (s *GitSyncService) ReceiveWebhook(ctx context.Context, repoKey string, payload []byte, headers map[string]string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.coreService == nil {
		return fmt.Errorf("git-sync-service not initialized")
	}

	return s.coreService.ReceiveWebhook(ctx, repoKey, payload, headers)
}

// IsReady 检查服务是否就绪
func (s *GitSyncService) IsReady() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.coreService != nil
}

// GetCoreService 获取底层核心服务
func (s *GitSyncService) GetCoreService() interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.coreService
}
