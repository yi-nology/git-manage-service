package workspace

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	workspace "github.com/yi-nology/git-manage-service/biz/model/workspace"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func GetWorkspaceStatus(ctx context.Context, c *app.RequestContext) {
	var req workspace.GetWorkspaceStatusReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	result, err := gitSvc.GetWorkspaceStatus(repo.Path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}

func GetWorkspaceDiff(ctx context.Context, c *app.RequestContext) {
	var req workspace.GetWorkspaceDiffReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	result, err := gitSvc.GetWorkspaceDiff(repo.Path, req.File, req.StagedOnly)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}

func StageFiles(ctx context.Context, c *app.RequestContext) {
	var req workspace.StageFilesReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	err = gitSvc.StageFiles(repo.Path, req.Files, req.StageAll)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func UnstageFiles(ctx context.Context, c *app.RequestContext) {
	var req workspace.UnstageFilesReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	err = gitSvc.UnstageFiles(repo.Path, req.Files, req.UnstageAll)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func CommitChanges(ctx context.Context, c *app.RequestContext) {
	var req workspace.CommitChangesReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Message == "" {
		response.BadRequest(c, "commit message is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	result, err := gitSvc.CommitChanges(repo.Path, req.Files, req.StageAll, req.Message, req.AuthorName, req.AuthorEmail, req.Push, req.PushRemote)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}

func PullWithResolve(ctx context.Context, c *app.RequestContext) {
	var req workspace.PullWithResolveReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	result, err := gitSvc.PullWithResolve(repo.Path, req.Remote, req.Branch, req.FetchOnly)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}

func GetConflictDetail(ctx context.Context, c *app.RequestContext) {
	var req workspace.GetConflictDetailReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	result, err := gitSvc.GetConflictDetail(repo.Path, req.File)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}

func MarkConflictResolved(ctx context.Context, c *app.RequestContext) {
	var req workspace.MarkConflictResolvedReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	err = gitSvc.MarkConflictResolved(repo.Path, req.File, req.ResolvedContent, req.Stage)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func AIResolveConflict(ctx context.Context, c *app.RequestContext) {
	var req workspace.AIResolveConflictReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	aiSvc := git.NewWorkspaceAIService()
	result, err := aiSvc.AIResolveConflict(req.OursContent, req.TheirsContent, req.BaseContent, req.File, req.Hint)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}
