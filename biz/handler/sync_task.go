package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/audit"
	"github.com/yi-nology/git-manage-service/biz/service/sync"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

// @Summary Create a sync task
// @Description Create a new automated synchronization task between two repositories or branches.
// @Tags Tasks
// @Accept json
// @Produce json
// @Param request body api.SyncTaskDTO true "Task info"
// @Success 200 {object} response.Response{data=api.SyncTaskDTO}
// @Failure 400 {object} response.Response "Bad Request"
// @Failure 500 {object} response.Response "Internal Server Error"
// @Router /api/sync/tasks [post]
func CreateTask(ctx context.Context, c *app.RequestContext) {
	var req po.SyncTask
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	req.Key = uuid.New().String()
	// Should validate Repo existence

	if err := db.NewSyncTaskDAO().Create(&req); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	sync.CronSvc.UpdateTask(req)
	audit.AuditSvc.Log(c, "CREATE", "task:"+req.Key, req)
	response.Success(c, api.NewSyncTaskDTO(req))
}

// @Summary List sync tasks for a repo
// @Description List all sync tasks, optionally filtered by repository key.
// @Tags Tasks
// @Param repo_key query string false "Repo Key"
// @Produce json
// @Success 200 {object} response.Response{data=[]api.SyncTaskDTO}
// @Router /api/sync/tasks [get]
func ListTasks(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	var tasks []po.SyncTask
	var err error

	taskDAO := db.NewSyncTaskDAO()

	if repoKey != "" {
		tasks, err = taskDAO.FindByRepoKey(repoKey)
	} else {
		tasks, err = taskDAO.FindAllWithRepos()
	}

	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	var dtos []api.SyncTaskDTO
	for _, t := range tasks {
		dtos = append(dtos, api.NewSyncTaskDTO(t))
	}
	response.Success(c, dtos)
}

// @Summary Get a sync task
// @Description Get details of a specific sync task by its unique key.
// @Tags Tasks
// @Param key path string true "Task Key"
// @Produce json
// @Success 200 {object} response.Response{data=api.SyncTaskDTO}
// @Failure 404 {object} response.Response "Task not found"
// @Router /api/sync/tasks/{key} [get]
func GetTask(ctx context.Context, c *app.RequestContext) {
	key := c.Param("key")

	task, err := db.NewSyncTaskDAO().FindByKey(key)
	if err != nil {
		response.NotFound(c, "task not found")
		return
	}
	response.Success(c, api.NewSyncTaskDTO(*task))
}

// @Summary Update a sync task
// @Description Update configuration of an existing sync task.
// @Tags Tasks
// @Param key path string true "Task Key"
// @Param request body api.SyncTaskDTO true "Task info"
// @Produce json
// @Success 200 {object} response.Response{data=api.SyncTaskDTO}
// @Failure 404 {object} response.Response "Task not found"
// @Failure 400 {object} response.Response "Bad Request"
// @Router /api/sync/tasks/{key} [put]
func UpdateTask(ctx context.Context, c *app.RequestContext) {
	key := c.Param("key")

	var req po.SyncTask
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	taskDAO := db.NewSyncTaskDAO()
	task, err := taskDAO.FindByKey(key)
	if err != nil {
		response.NotFound(c, "task not found")
		return
	}

	// Update fields
	task.SourceRepoKey = req.SourceRepoKey
	task.SourceRemote = req.SourceRemote
	task.SourceBranch = req.SourceBranch
	task.TargetRepoKey = req.TargetRepoKey
	task.TargetRemote = req.TargetRemote
	task.TargetBranch = req.TargetBranch
	task.PushOptions = req.PushOptions
	task.Cron = req.Cron
	task.Enabled = req.Enabled

	// Reset webhook token if needed or requested?
	// For now keep existing or allow update if passed?

	if err := taskDAO.Save(task); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	sync.CronSvc.UpdateTask(*task)
	audit.AuditSvc.Log(c, "UPDATE", "task:"+task.Key, task)

	response.Success(c, api.NewSyncTaskDTO(*task))
}

// @Summary Delete a sync task
// @Description Delete a sync task.
// @Tags Tasks
// @Param key path string true "Task Key"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response "Task not found"
// @Router /api/sync/tasks/{key} [delete]
func DeleteTask(ctx context.Context, c *app.RequestContext) {
	key := c.Param("key")

	taskDAO := db.NewSyncTaskDAO()
	task, err := taskDAO.FindByKey(key)
	if err != nil {
		response.NotFound(c, "task not found")
		return
	}

	taskDAO.Delete(task)
	sync.CronSvc.RemoveTask(task.ID)
	audit.AuditSvc.Log(c, "DELETE", "task:"+task.Key, nil)

	response.Success(c, map[string]string{"message": "deleted"})
}

// @Summary Trigger a sync task manually
// @Description Manually trigger a configured sync task to run immediately.
// @Tags Sync
// @Accept json
// @Produce json
// @Param request body api.RunSyncReq true "Task Key"
// @Success 200 {object} response.Response "Status started"
// @Failure 400 {object} response.Response "Bad Request"
// @Router /api/sync/run [post]
func RunSync(ctx context.Context, c *app.RequestContext) {
	var req api.RunSyncReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	go func() {
		svc := sync.NewSyncService()
		svc.RunTask(req.TaskKey)
	}()

	audit.AuditSvc.Log(c, "SYNC", "task_key:"+req.TaskKey, nil)
	response.Success(c, map[string]string{"status": "started"})
}

// @Summary Execute an ad-hoc sync
// @Tags Sync
// @Accept json
// @Produce json
// @Param request body api.ExecuteSyncReq true "Sync info"
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /api/sync/execute [post]
func ExecuteSync(ctx context.Context, c *app.RequestContext) {
	var req api.ExecuteSyncReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	// Construct a temporary task
	task := po.SyncTask{
		Key:           uuid.New().String(),
		SourceRepoKey: repo.Key,
		SourceRepo:    *repo,
		SourceRemote:  req.SourceRemote,
		SourceBranch:  req.SourceBranch,
		TargetRepoKey: repo.Key, // Same repo for ad-hoc sync (usually)
		TargetRepo:    *repo,
		TargetRemote:  req.TargetRemote,
		TargetBranch:  req.TargetBranch,
		PushOptions:   req.PushOptions,
	}

	go func() {
		svc := sync.NewSyncService()
		svc.ExecuteSync(&task)
	}()

	audit.AuditSvc.Log(c, "SYNC_ADHOC", "task:"+task.Key, task)
	response.Success(c, map[string]string{"status": "started", "task_key": task.Key})
}

// @Summary Get sync execution history
// @Description Get the history of sync executions, optionally filtered by repository.
// @Tags History
// @Param repo_key query string false "Repo Key"
// @Produce json
// @Success 200 {object} response.Response{data=[]api.SyncRunDTO}
// @Router /api/sync/history [get]
func ListHistory(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	var runs []po.SyncRun
	var err error

	runDAO := db.NewSyncRunDAO()

	if repoKey != "" {
		// Find tasks related to this repo
		taskKeys, _ := db.NewSyncTaskDAO().GetKeysByRepoKey(repoKey)

		if len(taskKeys) > 0 {
			runs, err = runDAO.FindByTaskKeys(taskKeys, 50)
		} else {
			// No tasks found, return empty history
			response.Success(c, []api.SyncRunDTO{})
			return
		}
	} else {
		runs, err = runDAO.FindLatest(50)
	}

	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	var dtos []api.SyncRunDTO
	for _, r := range runs {
		dtos = append(dtos, api.NewSyncRunDTO(r))
	}
	response.Success(c, dtos)
}

// @Summary Delete a sync history record
// @Tags History
// @Param id path int true "History ID"
// @Success 200 {object} response.Response
// @Router /api/sync/history/{id} [delete]
func DeleteHistory(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	if err := db.NewSyncRunDAO().Delete(uint(id)); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, map[string]string{"message": "deleted"})
}
