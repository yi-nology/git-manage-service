package branch

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/branch"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func Compare(ctx context.Context, c *app.RequestContext) {
	var req branch.CompareBranchRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()

	stat, err := gitSvc.GetDiffStat(repo.Path, req.GetBase(), req.GetTarget())
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	files, err := gitSvc.GetDiffFiles(repo.Path, req.GetBase(), req.GetTarget())
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]interface{}{
		"stat":  stat,
		"files": files,
	})
}

func GetDiff(ctx context.Context, c *app.RequestContext) {
	var req branch.GetDiffRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	content, err := gitSvc.GetRawDiff(repo.Path, req.GetBase(), req.GetTarget(), req.GetFile())
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]string{"diff": content})
}

func GetPatch(ctx context.Context, c *app.RequestContext) {
	var req branch.GetPatchRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	patch, err := gitSvc.GetPatch(repo.Path, req.GetBase(), req.GetTarget())
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s-%s-%s.patch", repo.Name, req.GetBase(), time.Now().Format("20060102")))
	c.Write([]byte(patch))
}
