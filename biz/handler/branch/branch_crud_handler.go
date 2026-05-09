package branch

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/branch"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/service/branchrule"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func List(ctx context.Context, c *app.RequestContext) {
	var req branch.ListBranchRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.GetRepoKey() == "" {
		response.BadRequest(c, "repo_key is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	branches, err := gitSvc.ListBranchesWithInfo(repo.Path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	branchTypeFilter := c.Query("type")
	if branchTypeFilter != "" && (branchTypeFilter == "local" || branchTypeFilter == "remote") {
		var filtered []domain.BranchInfo
		for _, b := range branches {
			if b.Type == branchTypeFilter {
				filtered = append(filtered, b)
			}
		}
		branches = filtered
	}

	if req.GetKeyword() != "" {
		var filtered []domain.BranchInfo
		keyword := strings.ToLower(req.GetKeyword())
		for _, b := range branches {
			if strings.Contains(strings.ToLower(b.Name), keyword) ||
				strings.Contains(strings.ToLower(b.Author), keyword) {
				filtered = append(filtered, b)
			}
		}
		branches = filtered
	}

	page := int(req.GetPage())
	if page < 1 {
		page = 1
	}
	pageSize := int(req.GetPageSize())
	if pageSize < 1 {
		pageSize = 100
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(branches) {
		start = len(branches)
	}
	if end > len(branches) {
		end = len(branches)
	}

	paged := branches[start:end]

	for i := range paged {
		b := &paged[i]
		if b.Upstream != "" {
			ahead, behind, _ := gitSvc.GetBranchSyncStatus(repo.Path, b.Name, b.Upstream)
			b.Ahead = ahead
			b.Behind = behind
		}
	}

	response.Success(c, map[string]interface{}{
		"total": len(branches),
		"list":  paged,
	})
}

func Create(ctx context.Context, c *app.RequestContext) {
	var req branch.CreateBranchRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	skipRules := string(c.GetHeader("X-Skip-Branch-Rules")) == "true"
	validation, err := branchrule.ValidateBranchName(req.GetRepoKey(), req.GetName(), req.GetBaseRef(), skipRules)
	if err == nil && !validation.Valid {
		response.BadRequest(c, fmt.Sprintf("分支规则校验失败: %s", validation.Message))
		return
	}

	gitSvc := git.NewGitService()
	if err := gitSvc.CreateBranch(repo.Path, req.GetName(), req.GetBaseRef()); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "created"})
}

func Delete(ctx context.Context, c *app.RequestContext) {
	var req branch.DeleteBranchRequest
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
	if err := gitSvc.DeleteBranch(repo.Path, req.GetName(), req.GetForce()); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "deleted"})
}

func Update(ctx context.Context, c *app.RequestContext) {
	var req branch.UpdateBranchRequest
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
	currentName := req.GetName()

	if req.GetNewName() != "" && req.GetNewName() != currentName {
		if err := gitSvc.RenameBranch(repo.Path, currentName, req.GetNewName()); err != nil {
			response.InternalServerError(c, err.Error())
			return
		}
		currentName = req.GetNewName()
	}

	if req.GetDesc() != "" {
		if err := gitSvc.SetBranchDescription(repo.Path, currentName, req.GetDesc()); err != nil {
		}
	}

	response.Success(c, map[string]string{"message": "updated"})
}

func Checkout(ctx context.Context, c *app.RequestContext) {
	var req branch.CheckoutBranchRequest
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
	if err := gitSvc.CheckoutBranch(repo.Path, req.GetName()); err != nil {
		response.BadRequest(c, "Checkout failed: "+err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "checked out " + req.GetName()})
}
