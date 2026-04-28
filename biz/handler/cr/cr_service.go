package cr

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/service/crservice"
	pkgresponse "github.com/yi-nology/git-manage-service/pkg/response"
)

func Create(ctx context.Context, c *app.RequestContext) {
	var req api.CreateCRReq
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if req.RepoKey == "" || req.Title == "" || req.SourceBranch == "" || req.TargetBranch == "" {
		pkgresponse.BadRequest(c, "repo_key, title, source_branch and target_branch are required")
		return
	}
	cr, err := crservice.CreateCR(ctx, &req)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to create CR: "+err.Error())
		return
	}
	pkgresponse.Success(c, cr)
}

func Get(ctx context.Context, c *app.RequestContext) {
	var req api.GetCRReq
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if req.RepoKey == "" || req.CRNumber == 0 {
		pkgresponse.BadRequest(c, "repo_key and cr_number are required")
		return
	}
	cr, err := crservice.GetCR(ctx, req.RepoKey, req.CRNumber)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to get CR: "+err.Error())
		return
	}
	pkgresponse.Success(c, cr)
}

func List(ctx context.Context, c *app.RequestContext) {
	var req api.ListCRsReq
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if req.RepoKey == "" {
		pkgresponse.BadRequest(c, "repo_key is required")
		return
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	crs, total, err := crservice.ListCRs(ctx, req.RepoKey, req.State, req.SourceBranch, req.TargetBranch, req.Page, req.PageSize)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to list CRs: "+err.Error())
		return
	}
	pkgresponse.Success(c, map[string]interface{}{
		"items": crs,
		"total": total,
	})
}

func Merge(ctx context.Context, c *app.RequestContext) {
	var req api.MergeCRReq
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if req.RepoKey == "" || req.CRNumber == 0 {
		pkgresponse.BadRequest(c, "repo_key and cr_number are required")
		return
	}
	cr, err := crservice.MergeCR(ctx, req.RepoKey, req.CRNumber, req.MergeCommitMessage, req.Squash, req.RemoveSourceBranch)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to merge CR: "+err.Error())
		return
	}
	pkgresponse.Success(c, cr)
}

func Close(ctx context.Context, c *app.RequestContext) {
	var req api.CloseCRReq
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if req.RepoKey == "" || req.CRNumber == 0 {
		pkgresponse.BadRequest(c, "repo_key and cr_number are required")
		return
	}
	cr, err := crservice.CloseCR(ctx, req.RepoKey, req.CRNumber)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to close CR: "+err.Error())
		return
	}
	pkgresponse.Success(c, cr)
}

func Sync(ctx context.Context, c *app.RequestContext) {
	var req api.SyncCRsReq
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if req.RepoKey == "" {
		pkgresponse.BadRequest(c, "repo_key is required")
		return
	}
	count, err := crservice.SyncCRs(ctx, req.RepoKey, req.State)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to sync CRs: "+err.Error())
		return
	}
	pkgresponse.Success(c, map[string]interface{}{"synced_count": count})
}

func Detect(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	if repoKey == "" {
		pkgresponse.BadRequest(c, "repo_key is required")
		return
	}
	repoDAO := db.NewRepoDAO()
	repo, err := repoDAO.FindByKey(repoKey)
	if err != nil {
		pkgresponse.NotFound(c, "Repo not found")
		return
	}
	result := map[string]interface{}{
		"provider_config_id": repo.ProviderConfigID,
		"platform_owner":     repo.PlatformOwner,
		"platform_repo":      repo.PlatformRepo,
	}
	pkgresponse.Success(c, result)
}

func ListByProvider(ctx context.Context, c *app.RequestContext) {
	providerID := c.Query("provider_id")
	owner := c.Query("owner")
	repoName := c.Query("repo")
	if providerID == "" || owner == "" || repoName == "" {
		pkgresponse.BadRequest(c, "provider_id, owner and repo are required")
		return
	}
	state := c.Query("state")
	page := c.GetInt("page")
	perPage := c.GetInt("per_page")
	if page == 0 {
		page = 1
	}
	if perPage == 0 {
		perPage = 20
	}
	crs, total, err := crservice.ListCRsByProvider(ctx, providerID, owner, repoName, state, page, perPage)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to list CRs: "+err.Error())
		return
	}
	pkgresponse.Success(c, map[string]interface{}{
		"items": crs,
		"total": total,
	})
}

func CreateByProvider(ctx context.Context, c *app.RequestContext) {
	var req api.CreateCRByProviderReq
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if req.ProviderID == 0 || req.Owner == "" || req.Repo == "" || req.Title == "" || req.SourceBranch == "" || req.TargetBranch == "" {
		pkgresponse.BadRequest(c, "provider_id, owner, repo, title, source_branch and target_branch are required")
		return
	}
	cr, err := crservice.CreateCRByProvider(ctx, &req)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to create CR: "+err.Error())
		return
	}
	pkgresponse.Success(c, cr)
}

func MergeByProvider(ctx context.Context, c *app.RequestContext) {
	var req struct {
		ProviderID         uint   `json:"provider_id"`
		Owner              string `json:"owner"`
		Repo               string `json:"repo"`
		CRNumber           int    `json:"cr_number"`
		MergeCommitMessage string `json:"merge_commit_message"`
		Squash             bool   `json:"squash"`
		RemoveSourceBranch bool   `json:"remove_source_branch"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if req.ProviderID == 0 || req.Owner == "" || req.Repo == "" || req.CRNumber == 0 {
		pkgresponse.BadRequest(c, "provider_id, owner, repo and cr_number are required")
		return
	}
	cr, err := crservice.MergeCRByProvider(ctx, req.ProviderID, req.Owner, req.Repo, req.CRNumber, req.MergeCommitMessage, req.Squash, req.RemoveSourceBranch)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to merge CR: "+err.Error())
		return
	}
	pkgresponse.Success(c, cr)
}

func CloseByProvider(ctx context.Context, c *app.RequestContext) {
	var req struct {
		ProviderID uint   `json:"provider_id"`
		Owner      string `json:"owner"`
		Repo       string `json:"repo"`
		CRNumber   int    `json:"cr_number"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if req.ProviderID == 0 || req.Owner == "" || req.Repo == "" || req.CRNumber == 0 {
		pkgresponse.BadRequest(c, "provider_id, owner, repo and cr_number are required")
		return
	}
	cr, err := crservice.CloseCRByProvider(ctx, req.ProviderID, req.Owner, req.Repo, req.CRNumber)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to close CR: "+err.Error())
		return
	}
	pkgresponse.Success(c, cr)
}
