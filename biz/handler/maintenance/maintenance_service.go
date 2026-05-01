package maintenance

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	maintenance "github.com/yi-nology/git-manage-service/biz/model/maintenance"
	"github.com/yi-nology/git-manage-service/biz/service/audit"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func Health(ctx context.Context, c *app.RequestContext) {
	var req maintenance.HealthRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := git.NewMaintenanceService()
	report, err := svc.AnalyzeHealth(repo.Path)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	largeFiles, err := svc.FindLargeFiles(repo.Path, 1*1024*1024)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	report.LargeFiles = largeFiles
	if report.LargeFiles == nil {
		report.LargeFiles = []api.LargeFileEntry{}
	}

	response.Success(c, report)
}

func Slim(ctx context.Context, c *app.RequestContext) {
	var req maintenance.SlimRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	paths := req.GetPaths()
	if len(paths) == 0 {
		response.BadRequest(c, "paths is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	taskID := uuid.New().String()
	git.GlobalTaskManager.AddTask(taskID)

	go func() {
		svc := git.NewMaintenanceService()
		if err := svc.SlimHistory(repo.Path, paths, req.GetAddGitignore(), taskID); err != nil {
			git.GlobalTaskManager.UpdateStatus(taskID, "failed", err.Error())
			return
		}
		git.GlobalTaskManager.UpdateStatus(taskID, "success", "")
	}()

	audit.AuditSvc.Log(c, "REPO_SLIM", "repo:"+repo.Key, map[string]string{
		"files": strings.Join(paths, ","),
	})

	response.Success(c, api.MaintenanceTaskResponse{TaskID: taskID})
}

func GC(ctx context.Context, c *app.RequestContext) {
	var req maintenance.HealthRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	taskID := uuid.New().String()
	git.GlobalTaskManager.AddTask(taskID)

	go func() {
		svc := git.NewMaintenanceService()
		if err := svc.GarbageCollect(repo.Path, taskID); err != nil {
			git.GlobalTaskManager.UpdateStatus(taskID, "failed", err.Error())
			return
		}
		git.GlobalTaskManager.UpdateStatus(taskID, "success", "")
	}()

	audit.AuditSvc.Log(c, "REPO_GC", "repo:"+repo.Key, nil)

	response.Success(c, api.MaintenanceTaskResponse{TaskID: taskID})
}

func Gitignore(ctx context.Context, c *app.RequestContext) {
	var req maintenance.GitignoreRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	paths := req.GetPaths()
	if len(paths) == 0 {
		response.BadRequest(c, "paths is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := git.NewMaintenanceService()
	if err := svc.AddToGitignore(repo.Path, paths); err != nil {
		response.InternalError(c, err)
		return
	}

	audit.AuditSvc.Log(c, "REPO_GITIGNORE", "repo:"+repo.Key, map[string]string{
		"files": strings.Join(paths, ","),
	})

	response.Success(c, api.MessageResponse{Message: "已添加到 .gitignore"})
}
