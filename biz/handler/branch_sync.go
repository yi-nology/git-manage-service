package handler

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal"
	"github.com/yi-nology/git-manage-service/biz/model"
	"github.com/yi-nology/git-manage-service/biz/pkg/response"
	"github.com/yi-nology/git-manage-service/biz/service"
)

type PushBranchReq struct {
	Remotes []string `json:"remotes"` // List of remote names
}

// @Summary Push branch to remotes
// @Tags Branches
// @Param id path int true "Repo ID"
// @Param name path string true "Branch Name"
// @Param request body PushBranchReq true "Remotes"
// @Success 200 {object} response.Response
// @Router /api/repos/{id}/branches/{name}/push [post]
func PushBranch(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	branch := c.Param("name")
	
	var req PushBranchReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var repo model.Repo
	if err := dal.DB.First(&repo, id).Error; err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := service.NewGitService()
	
	var errors []string
	for _, remote := range req.Remotes {
		if err := gitSvc.PushBranch(repo.Path, remote, branch); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", remote, err))
		}
	}
	
	if len(errors) > 0 {
		response.InternalServerError(c, strings.Join(errors, "; "))
		return
	}

	service.AuditSvc.Log(c, "PUSH_BRANCH", "repo:"+repo.Key, map[string]interface{}{
		"branch": branch,
		"remotes": req.Remotes,
	})
	response.Success(c, map[string]string{"message": "pushed"})
}

// @Summary Pull/Sync branch from upstream
// @Tags Branches
// @Param id path int true "Repo ID"
// @Param name path string true "Branch Name"
// @Success 200 {object} response.Response
// @Router /api/repos/{id}/branches/{name}/pull [post]
func PullBranch(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)
	branch := c.Param("name")

	var repo model.Repo
	if err := dal.DB.First(&repo, id).Error; err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := service.NewGitService()
	
	// Get Upstream
	// We need to know which remote is upstream. 
	// The `branch.<name>.remote` config tells us.
	// Or we can just try `git pull` if it's the current branch, but for non-current branches it's tricky.
	// `git pull` only works on current HEAD.
	// So we must check if `branch` is current.
	
	branches, _ := gitSvc.ListBranchesWithInfo(repo.Path)
	var isCurrent bool
	var upstreamRemote string
	
	for _, b := range branches {
		if b.Name == branch {
			isCurrent = b.IsCurrent
			if b.Upstream != "" {
				parts := strings.Split(b.Upstream, "/")
				if len(parts) > 0 {
					upstreamRemote = parts[0]
				}
			}
			break
		}
	}
	
	if !isCurrent {
		// Can't pull non-checked out branch easily without fetching + rebase/merge manually.
		// For now, let's just Fetch origin:branch.
		// Actually requirement says "Sync function... execute git pull --rebase".
		// This implies we are working on the workspace.
		// If it's not checked out, we should probably tell user to checkout first or we just fetch.
		response.BadRequest(c, "Can only sync currently checked out branch")
		return
	}
	
	if upstreamRemote == "" {
		response.BadRequest(c, "No upstream configured for this branch")
		return
	}

	if err := gitSvc.PullBranch(repo.Path, upstreamRemote, branch); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	service.AuditSvc.Log(c, "PULL_BRANCH", "repo:"+repo.Key, map[string]string{
		"branch": branch,
		"remote": upstreamRemote,
	})
	response.Success(c, map[string]string{"message": "synced"})
}
