package sync

import (
	"context"
	"fmt"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/sync/adapter"
)

// SyncServiceV2 使用 git-sync-service 核心库的同步服务
type SyncServiceV2 struct {
	gitSyncSvc *adapter.GitSyncService
}

// NewSyncServiceV2 创建同步服务
func NewSyncServiceV2() *SyncServiceV2 {
	svc := &SyncServiceV2{
		gitSyncSvc: adapter.GetGitSyncService(),
	}

	svc.initializeGitSyncService()

	return svc
}

// initializeGitSyncService 初始化 git-sync-service 核心库
func (s *SyncServiceV2) initializeGitSyncService() {
	gsCfg := adapter.DefaultConfig()
	if err := s.gitSyncSvc.Initialize(gsCfg); err != nil {
		fmt.Printf("[WARNING] git-sync-service initialization failed: %v\n", err)
		return
	}

	fmt.Printf("[INFO] git-sync-service initialized successfully\n")
}

// RunTask 执行同步任务
func (s *SyncServiceV2) RunTask(taskKey string) error {
	return s.RunTaskWithTrigger(taskKey, po.TriggerSourceManual)
}

// RunTaskWithTrigger 带触发源的执行
func (s *SyncServiceV2) RunTaskWithTrigger(taskKey string, triggerSource string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	_, err := s.gitSyncSvc.RunTask(ctx, taskKey, triggerSource)
	return err
}

// ExecuteSync 直接执行同步
func (s *SyncServiceV2) ExecuteSync(task *po.SyncTask) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	_, err := s.gitSyncSvc.ExecuteSync(ctx, task)
	return err
}

// PreviewSync 同步预览
func (s *SyncServiceV2) PreviewSync(task *po.SyncTask) (*adapter.SyncPreviewResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	return s.gitSyncSvc.PreviewSync(ctx, task)
}

// BatchSync 批量同步
func (s *SyncServiceV2) BatchSync(taskKeys []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	_, err := s.gitSyncSvc.BatchSync(ctx, taskKeys)
	return err
}

// RegisterRepo 注册仓库
func (s *SyncServiceV2) RegisterRepo(repo *po.Repo) error {
	ctx := context.Background()
	return s.gitSyncSvc.RegisterRepo(ctx, repo)
}

// UnregisterRepo 注销仓库
func (s *SyncServiceV2) UnregisterRepo(repoKey string) error {
	ctx := context.Background()
	return s.gitSyncSvc.UnregisterRepo(ctx, repoKey)
}

// ReceiveWebhook 接收 Webhook 事件
func (s *SyncServiceV2) ReceiveWebhook(repoKey string, payload []byte, headers map[string]string) error {
	ctx := context.Background()
	return s.gitSyncSvc.ReceiveWebhook(ctx, repoKey, payload, headers)
}

// IsReady 检查服务是否就绪
func (s *SyncServiceV2) IsReady() bool {
	return s.gitSyncSvc.IsReady()
}

// Start 启动后台调度
func (s *SyncServiceV2) Start() {
	if s.gitSyncSvc.IsReady() {
		s.gitSyncSvc.Start()
	}
}

// Stop 停止后台调度
func (s *SyncServiceV2) Stop() {
	if s.gitSyncSvc.IsReady() {
		s.gitSyncSvc.Stop()
	}
}

// GetImplementationInfo 获取实现信息
func (s *SyncServiceV2) GetImplementationInfo() map[string]interface{} {
	return map[string]interface{}{
		"service_ready":  s.gitSyncSvc.IsReady(),
		"implementation": "git-sync-service",
	}
}
