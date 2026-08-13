package syncv2

import (
	"context"
	"log"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	syncv2 "github.com/yi-nology/git-manage-service/biz/service/sync/v2"
	"github.com/yi-nology/git-manage-service/pkg/response"
	gitsyncmodel "github.com/yi-nology/git-sync-service/sync/model"
)

var svc = syncv2.GetService()

// ==================== Stats API ====================

// GetStats 获取统计数据
func GetStats(ctx context.Context, c *app.RequestContext) {
	stats, err := svc.GetStats(ctx)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, stats)
}

// ==================== Task API ====================

// ListTasks 获取任务列表
func ListTasks(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")

	tasks, err := svc.ListTasks(ctx, repoKey)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, tasks)
}

// GetTask 获取任务详情
func GetTask(ctx context.Context, c *app.RequestContext) {
	key := c.Query("key")
	if key == "" {
		response.BadRequest(c, "key is required")
		return
	}

	task, err := svc.GetTask(ctx, key)
	if err != nil {
		response.NotFound(c, "task not found")
		return
	}

	response.Success(c, task)
}

// CreateTask 创建任务
func CreateTask(ctx context.Context, c *app.RequestContext) {
	var req gitsyncmodel.CreateTaskRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	task, err := svc.CreateTask(ctx, &req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, task)
}

// UpdateTask 更新任务
func UpdateTask(ctx context.Context, c *app.RequestContext) {
	var req gitsyncmodel.UpdateTaskRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	task, err := svc.UpdateTask(ctx, &req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, task)
}

// DeleteTask 删除任务
func DeleteTask(ctx context.Context, c *app.RequestContext) {
	key := c.Query("key")
	if key == "" {
		response.BadRequest(c, "key is required")
		return
	}

	if err := svc.DeleteTask(ctx, key); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "deleted"})
}

// RunTask 运行任务
func RunTask(ctx context.Context, c *app.RequestContext) {
	key := c.Query("key")
	if key == "" {
		response.BadRequest(c, "key is required")
		return
	}

	go func() {
		if err := svc.RunTask(context.Background(), key); err != nil {
			log.Printf("[sync] async RunTask %s failed: %v", key, err)
		}
	}()

	response.Success(c, map[string]string{"status": "started"})
}

// BatchRunTasks 批量运行任务
func BatchRunTasks(ctx context.Context, c *app.RequestContext) {
	var req struct {
		TaskKeys []string `json:"task_keys"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if len(req.TaskKeys) == 0 {
		response.BadRequest(c, "task_keys is required")
		return
	}

	go func() {
		if err := svc.BatchRunTasks(context.Background(), req.TaskKeys); err != nil {
			log.Printf("[sync] async BatchRunTasks failed: %v", err)
		}
	}()

	response.Success(c, map[string]interface{}{"status": "started", "count": len(req.TaskKeys)})
}

// PreviewSync 预览同步
func PreviewSync(ctx context.Context, c *app.RequestContext) {
	var req gitsyncmodel.PreviewSyncRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	result, err := svc.PreviewSync(ctx, &req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// ==================== History API ====================

// ListHistory 获取执行历史
func ListHistory(ctx context.Context, c *app.RequestContext) {
	taskKey := c.Query("task_key")
	limitStr := c.Query("limit")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	var runs []*gitsyncmodel.SyncRun
	var err error

	if taskKey != "" {
		runs, err = svc.ListHistoryByTask(ctx, taskKey, limit)
	} else {
		runs, err = svc.ListRecentHistory(ctx, limit)
	}

	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, runs)
}

// DeleteHistory 删除历史记录
func DeleteHistory(ctx context.Context, c *app.RequestContext) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := svc.DeleteHistory(ctx, uint(id)); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "deleted"})
}

// ==================== Webhook Rule API ====================

// ListRules 获取 Webhook 规则列表
func ListRules(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")

	rules, err := svc.ListRules(ctx, repoKey)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, rules)
}

// GetRule 获取规则详情
func GetRule(ctx context.Context, c *app.RequestContext) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	rule, err := svc.GetRule(ctx, uint(id))
	if err != nil {
		response.NotFound(c, "rule not found")
		return
	}

	response.Success(c, rule)
}

// CreateRule 创建规则
func CreateRule(ctx context.Context, c *app.RequestContext) {
	var req gitsyncmodel.CreateRuleRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	rule, err := svc.CreateRule(ctx, &req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, rule)
}

// UpdateRule 更新规则
func UpdateRule(ctx context.Context, c *app.RequestContext) {
	var req gitsyncmodel.UpdateRuleRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	rule, err := svc.UpdateRule(ctx, &req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, rule)
}

// DeleteRule 删除规则
func DeleteRule(ctx context.Context, c *app.RequestContext) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := svc.DeleteRule(ctx, uint(id)); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "deleted"})
}

// ==================== Webhook Event API ====================

// ListEvents 获取 Webhook 事件列表
func ListEvents(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	limitStr := c.Query("limit")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	events, err := svc.ListEvents(ctx, repoKey, limit)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, events)
}

// RetryEvent 重试事件
func RetryEvent(ctx context.Context, c *app.RequestContext) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	go func() {
		_ = svc.RetryEvent(context.Background(), uint(id))
	}()

	response.Success(c, map[string]string{"status": "retried"})
}
