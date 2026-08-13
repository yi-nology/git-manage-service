package syncv2

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/yi-nology/git-manage-service/pkg/configs"
	gitsync "github.com/yi-nology/git-sync-service/sync"
	gitsyncmodel "github.com/yi-nology/git-sync-service/sync/model"
)

var (
	instance *SyncServiceV2
	once     sync.Once
)

// SyncStats 同步统计数据
type SyncStats struct {
	TotalTasks   int `json:"totalTasks"`
	EnabledTasks int `json:"enabledTasks"`
	TodayRuns    int `json:"todayRuns"`
	FailedRuns   int `json:"failedRuns"`
	RunningTasks int `json:"runningTasks"`
}

// SyncConfigItem 同步配置项
type SyncConfigItem struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Scope       string `json:"scope"`
	Description string `json:"description"`
}

// SyncServiceV2 基于 git-sync-service 的 V2 同步服务
type SyncServiceV2 struct {
	core *gitsync.Service
}

// GetService 获取单例
func GetService() *SyncServiceV2 {
	once.Do(func() {
		instance = &SyncServiceV2{}
	})
	return instance
}

// Initialize 初始化服务
func (s *SyncServiceV2) Initialize(appCfg *configs.Config) error {
	// 构建 git-sync-service 配置
	cfg := &gitsyncmodel.Config{
		Database: gitsyncmodel.DatabaseConfig{
			Driver:       string(appCfg.Database.Type),
			DSN:          buildDSN(appCfg),
			MaxIdleConns: 10,
			MaxOpenConns: 100,
		},
		Redis: gitsyncmodel.RedisConfig{
			Addr:     appCfg.Redis.Addr,
			Password: appCfg.Redis.Password,
			DB:       appCfg.Redis.DB,
		},
		Git: gitsyncmodel.GitConfig{
			TempDir: "/tmp/git-manage-sync",
		},
		Sync: gitsyncmodel.SyncConfig{
			MaxConcurrent:  5,
			DefaultTimeout: 300,
			RetryCount:     3,
		},
	}

	// 初始化核心服务
	core, err := gitsync.NewService(cfg)
	if err != nil {
		return fmt.Errorf("failed to create git-sync-service: %w", err)
	}

	s.core = core

	// 启动核心服务
	if err := s.core.Start(); err != nil {
		return fmt.Errorf("failed to start git-sync-service: %w", err)
	}

	return nil
}

// Stop 停止服务
func (s *SyncServiceV2) Stop() {
	if s.core != nil {
		s.core.Stop()
	}
}

// GetCore 获取底层核心服务
func (s *SyncServiceV2) GetCore() *gitsync.Service {
	return s.core
}

// ==================== Task API ====================

// ListTasks 获取任务列表
func (s *SyncServiceV2) ListTasks(ctx context.Context, repoKey string) ([]*gitsyncmodel.SyncTask, error) {
	tasks, _, err := s.core.ListTasks(ctx, repoKey, 0, 1000)
	return tasks, err
}

// GetTask 获取任务详情
func (s *SyncServiceV2) GetTask(ctx context.Context, key string) (*gitsyncmodel.SyncTask, error) {
	return s.core.GetTask(ctx, key)
}

// FindTaskByWebhookToken finds a sync task by its auto-generated webhook trigger
// token. git-sync-service mints a unique WebhookToken per task but exposes no
// direct lookup, so this scans the task list. Returns (nil, nil) when nothing
// matches; returns an error only when the service is uninitialized or the list
// call fails.
func (s *SyncServiceV2) FindTaskByWebhookToken(ctx context.Context, token string) (*gitsyncmodel.SyncTask, error) {
	if s.core == nil {
		return nil, fmt.Errorf("sync service not initialized")
	}
	tasks, _, err := s.core.ListTasks(ctx, "", 0, 1000)
	if err != nil {
		return nil, err
	}
	for _, t := range tasks {
		if t.WebhookToken == token {
			return t, nil
		}
	}
	return nil, nil
}

// CreateTask 创建任务
func (s *SyncServiceV2) CreateTask(ctx context.Context, req *gitsyncmodel.CreateTaskRequest) (*gitsyncmodel.SyncTask, error) {
	return s.core.CreateTask(ctx, req)
}

// UpdateTask 更新任务
func (s *SyncServiceV2) UpdateTask(ctx context.Context, req *gitsyncmodel.UpdateTaskRequest) (*gitsyncmodel.SyncTask, error) {
	return s.core.UpdateTask(ctx, req)
}

// DeleteTask 删除任务
func (s *SyncServiceV2) DeleteTask(ctx context.Context, key string) error {
	return s.core.DeleteTask(ctx, key)
}

// RunTask 运行任务
func (s *SyncServiceV2) RunTask(ctx context.Context, taskKey string) error {
	return s.core.RunTask(ctx, taskKey)
}

// BatchRunTasks 批量运行任务
func (s *SyncServiceV2) BatchRunTasks(ctx context.Context, taskKeys []string) error {
	for _, key := range taskKeys {
		go func(k string) {
			_ = s.core.RunTask(context.Background(), k)
		}(key)
	}
	return nil
}

// PreviewSync 预览同步
func (s *SyncServiceV2) PreviewSync(ctx context.Context, req *gitsyncmodel.PreviewSyncRequest) (*gitsyncmodel.PreviewSyncResult, error) {
	return s.core.PreviewSync(ctx, req)
}

// ==================== History API ====================

// ListHistoryByTask 获取指定任务的执行历史
func (s *SyncServiceV2) ListHistoryByTask(ctx context.Context, taskKey string, limit int) ([]*gitsyncmodel.SyncRun, error) {
	runs, _, err := s.core.ListHistory(ctx, taskKey, 0, limit)
	return runs, err
}

// ListRecentHistory 获取最近的执行历史
func (s *SyncServiceV2) ListRecentHistory(ctx context.Context, limit int) ([]*gitsyncmodel.SyncRun, error) {
	runs, _, err := s.core.ListHistory(ctx, "", 0, limit)
	return runs, err
}

// DeleteHistory 删除历史记录
func (s *SyncServiceV2) DeleteHistory(ctx context.Context, id uint) error {
	return s.core.DeleteHistory(ctx, id)
}

// ==================== Stats API ====================

// GetStats 获取统计数据
func (s *SyncServiceV2) GetStats(ctx context.Context) (*SyncStats, error) {
	tasks, _, err := s.core.ListTasks(ctx, "", 0, 1000)
	if err != nil {
		return nil, err
	}

	runs, _, err := s.core.ListHistory(ctx, "", 0, 100)
	if err != nil {
		return nil, err
	}

	stats := &SyncStats{
		TotalTasks:   len(tasks),
		EnabledTasks: countEnabled(tasks),
		TodayRuns:    countTodayRuns(runs),
		FailedRuns:   countFailedRuns(runs),
		RunningTasks: countRunningTasks(runs),
	}

	return stats, nil
}

// ==================== Webhook Rule API ====================

// ListRules 获取 Webhook 规则列表
func (s *SyncServiceV2) ListRules(ctx context.Context, repoKey string) ([]*gitsyncmodel.WebhookRule, error) {
	return s.core.ListRules(ctx, repoKey)
}

// GetRule 获取规则详情
func (s *SyncServiceV2) GetRule(ctx context.Context, id uint) (*gitsyncmodel.WebhookRule, error) {
	return s.core.GetRule(ctx, id)
}

// CreateRule 创建规则
func (s *SyncServiceV2) CreateRule(ctx context.Context, req *gitsyncmodel.CreateRuleRequest) (*gitsyncmodel.WebhookRule, error) {
	return s.core.CreateRule(ctx, req)
}

// UpdateRule 更新规则
func (s *SyncServiceV2) UpdateRule(ctx context.Context, req *gitsyncmodel.UpdateRuleRequest) (*gitsyncmodel.WebhookRule, error) {
	return s.core.UpdateRule(ctx, req)
}

// DeleteRule 删除规则
func (s *SyncServiceV2) DeleteRule(ctx context.Context, id uint) error {
	return s.core.DeleteRule(ctx, id)
}

// ==================== Webhook Event API ====================

// ListEvents 获取 Webhook 事件列表
func (s *SyncServiceV2) ListEvents(ctx context.Context, repoKey string, limit int) ([]*gitsyncmodel.WebhookEvent, error) {
	events, _, err := s.core.ListEvents(ctx, repoKey, 0, limit)
	return events, err
}

// RetryEvent 重试事件
func (s *SyncServiceV2) RetryEvent(ctx context.Context, id uint) error {
	return s.core.RetryEvent(ctx, id)
}

// ==================== Helper Functions ====================

func buildDSN(cfg *configs.Config) string {
	return "data/git_sync_v2.db"
}

func countEnabled(tasks []*gitsyncmodel.SyncTask) int {
	count := 0
	for _, t := range tasks {
		if t.Enabled {
			count++
		}
	}
	return count
}

func countTodayRuns(runs []*gitsyncmodel.SyncRun) int {
	today := time.Now().Truncate(24 * time.Hour)
	count := 0
	for _, r := range runs {
		if r.StartTime.After(today) {
			count++
		}
	}
	return count
}

func countFailedRuns(runs []*gitsyncmodel.SyncRun) int {
	count := 0
	for _, r := range runs {
		if r.Status == "failed" {
			count++
		}
	}
	return count
}

func countRunningTasks(runs []*gitsyncmodel.SyncRun) int {
	count := 0
	for _, r := range runs {
		if r.Status == "running" {
			count++
		}
	}
	return count
}
