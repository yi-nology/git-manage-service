package branch

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/branch"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func CherryPick(ctx context.Context, c *app.RequestContext) {
	var req branch.CherryPickRequest
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
	newCommit, conflicts, err := gitSvc.CherryPick(repo.Path, req.GetCommitHash(), false)
	if err != nil {
		if len(conflicts) > 0 {
			c.JSON(200, response.Response{
				Code: 409,
				Msg:  "Cherry-pick conflict",
				Data: map[string]interface{}{
					"conflicts": conflicts,
				},
			})
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]interface{}{
		"success":    true,
		"new_commit": newCommit,
	})
}

func Rebase(ctx context.Context, c *app.RequestContext) {
	var req branch.RebaseRequest
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
	success, conflicts, err := gitSvc.Rebase(repo.Path, req.GetUpstream(), "")
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	if !success {
		c.JSON(200, response.Response{
			Code: 409,
			Msg:  "Rebase conflict",
			Data: map[string]interface{}{
				"success":     false,
				"in_progress": true,
				"conflicts":   conflicts,
			},
		})
		return
	}

	response.Success(c, map[string]interface{}{
		"success":     true,
		"in_progress": false,
	})
}

func RebaseAbort(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	if repoKey == "" {
		response.BadRequest(c, "repo_key is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	if err := gitSvc.RebaseAbort(repo.Path); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "rebase aborted"})
}

func RebaseContinue(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	if repoKey == "" {
		response.BadRequest(c, "repo_key is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	success, conflicts, err := gitSvc.RebaseContinue(repo.Path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	if !success {
		c.JSON(200, response.Response{
			Code: 409,
			Msg:  "Rebase conflict",
			Data: map[string]interface{}{
				"success":     false,
				"in_progress": true,
				"conflicts":   conflicts,
			},
		})
		return
	}

	response.Success(c, map[string]interface{}{
		"success":     true,
		"in_progress": false,
	})
}

func Merge(ctx context.Context, c *app.RequestContext) {
	var req branch.MergeBranchRequest
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

	check, err := gitSvc.MergeDryRun(repo.Path, req.GetSource(), req.GetTarget())
	if err != nil {
		response.InternalServerError(c, "Pre-merge check failed: "+err.Error())
		return
	}
	if !check.Success {
		mergeID := uuid.New().String()
		reportURL := fmt.Sprintf("/merge_report.html?repo_key=%s&source=%s&target=%s&merge_id=%s", repo.Key, req.GetSource(), req.GetTarget(), mergeID)

		c.JSON(200, response.Response{
			Code: 409,
			Msg:  "Merge conflict detected",
			Data: map[string]interface{}{
				"conflicts":  check.Conflicts,
				"report_url": reportURL,
				"merge_id":   mergeID,
			},
		})
		return
	}

	if err := gitSvc.Merge(repo.Path, req.GetSource(), req.GetTarget(), req.GetMessage(), req.GetNoFf(), req.GetSquash()); err != nil {
		response.InternalServerError(c, "Merge execution failed: "+err.Error())
		return
	}

	c.Set("audit_target", "repo:"+repo.Key)
	c.Set("audit_details", map[string]string{"source": req.GetSource(), "target": req.GetTarget()})

	response.Success(c, map[string]string{"status": "merged"})
}

func MergeCheck(ctx context.Context, c *app.RequestContext) {
	var req branch.MergeCheckRequest
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
	result, err := gitSvc.MergeDryRun(repo.Path, req.GetBase(), req.GetTarget())
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}
