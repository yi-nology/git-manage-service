package system

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/service/branchrule"
	pkgresponse "github.com/yi-nology/git-manage-service/pkg/response"
)

func GetBranchRules(ctx context.Context, c *app.RequestContext) {
	result, err := branchrule.GetGlobalRules()
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func UpdateBranchRules(ctx context.Context, c *app.RequestContext) {
	var req api.BranchRuleSetDTO
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	result, err := branchrule.UpdateGlobalRules(req)
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func GetRemoteRepoBranchRules(ctx context.Context, c *app.RequestContext) {
	providerID, err := parseIDParam(c.Param("provider_id"))
	if err != nil {
		pkgresponse.BadRequest(c, "invalid provider_id")
		return
	}
	owner := c.Param("owner")
	repo := c.Param("repo")
	if owner == "" || repo == "" {
		pkgresponse.BadRequest(c, "owner and repo are required")
		return
	}
	result, err := branchrule.GetRemoteRepoRules(providerID, owner, repo)
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func UpdateRemoteRepoBranchRules(ctx context.Context, c *app.RequestContext) {
	providerID, err := parseIDParam(c.Param("provider_id"))
	if err != nil {
		pkgresponse.BadRequest(c, "invalid provider_id")
		return
	}
	owner := c.Param("owner")
	repo := c.Param("repo")
	if owner == "" || repo == "" {
		pkgresponse.BadRequest(c, "owner and repo are required")
		return
	}
	var req api.RemoteRepoBranchRulesDTO
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	result, err := branchrule.UpdateRemoteRepoRules(providerID, owner, repo, req)
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func ValidateBranchName(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	branchName := c.Query("branch_name")
	baseRef := c.Query("base_ref")
	skipRules := c.Query("skip_rules") == "true"
	if repoKey == "" || branchName == "" {
		pkgresponse.BadRequest(c, "repo_key and branch_name are required")
		return
	}
	result, err := branchrule.ValidateBranchName(repoKey, branchName, baseRef, skipRules)
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func init() {
	_ = strconv.ErrSyntax
}
