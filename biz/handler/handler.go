package handler

import (
	"context"
	"git-sync-tool/biz/dal"
	"git-sync-tool/biz/model"
	"git-sync-tool/biz/service"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// Repo Handlers
func RegisterRepo(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	// Validate path
	gitSvc := service.NewGitService()
	if !gitSvc.IsGitRepo(req.Path) {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": "path is not a valid git repository"})
		return
	}

	repo := model.Repo{
		Name: req.Name,
		Path: req.Path,
	}
	if err := dal.DB.Create(&repo).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	c.JSON(consts.StatusOK, repo)
}

func ListRepos(ctx context.Context, c *app.RequestContext) {
	var repos []model.Repo
	dal.DB.Find(&repos)
	c.JSON(consts.StatusOK, repos)
}

// Task Handlers
func CreateTask(ctx context.Context, c *app.RequestContext) {
	var req model.SyncTask
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	if err := dal.DB.Create(&req).Error; err != nil {
		c.JSON(consts.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	service.CronSvc.UpdateTask(req)
	c.JSON(consts.StatusOK, req)
}

func ListTasks(ctx context.Context, c *app.RequestContext) {
	var tasks []model.SyncTask
	dal.DB.Preload("SourceRepo").Preload("TargetRepo").Find(&tasks)
	c.JSON(consts.StatusOK, tasks)
}

func GetTask(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var task model.SyncTask
	if err := dal.DB.Preload("SourceRepo").Preload("TargetRepo").First(&task, id).Error; err != nil {
		c.JSON(consts.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}
	c.JSON(consts.StatusOK, task)
}

func UpdateTask(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var req model.SyncTask
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	var task model.SyncTask
	if err := dal.DB.First(&task, id).Error; err != nil {
		c.JSON(consts.StatusNotFound, map[string]string{"error": "task not found"})
		return
	}

	// Update fields
	task.SourceRepoID = req.SourceRepoID
	task.SourceRemote = req.SourceRemote
	task.SourceBranch = req.SourceBranch
	task.TargetRepoID = req.TargetRepoID
	task.TargetRemote = req.TargetRemote
	task.TargetBranch = req.TargetBranch
	task.PushOptions = req.PushOptions
	task.Cron = req.Cron
	task.Enabled = req.Enabled

	dal.DB.Save(&task)
	service.CronSvc.UpdateTask(task)

	c.JSON(consts.StatusOK, task)
}

func RunSync(ctx context.Context, c *app.RequestContext) {
	var req struct {
		TaskID uint `json:"task_id"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(consts.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	go func() {
		svc := service.NewSyncService()
		svc.RunTask(req.TaskID)
	}()

	c.JSON(consts.StatusOK, map[string]string{"status": "started"})
}

func ListHistory(ctx context.Context, c *app.RequestContext) {
	var runs []model.SyncRun
	dal.DB.Order("start_time desc").Limit(50).Preload("Task").Find(&runs)
	c.JSON(consts.StatusOK, runs)
}
