package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	providerModel "github.com/yi-nology/git-manage-service/biz/model/provider"
	"github.com/yi-nology/git-manage-service/biz/service/provider_manager"
	"github.com/yi-nology/git-manage-service/pkg/handler"
	pkgresponse "github.com/yi-nology/git-manage-service/pkg/response"
	"github.com/yi-nology/git-platform-sdk/provider"
)

func toProtoProviderConfig(cfg *po.ProviderConfig) *providerModel.ProviderConfig {
	return &providerModel.ProviderConfig{
		Id:            uint64(cfg.ID),
		Name:          cfg.Name,
		Platform:      cfg.Platform,
		BaseUrl:       cfg.BaseURL,
		CredentialId:  uint64(cfg.CredentialID),
		WebhookSecret: cfg.WebhookSecret,
		SkipTls:       cfg.SkipTLS,
		CreatedAt:     cfg.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     cfg.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toProviderConfigDTO(cfg *po.ProviderConfig) api.ProviderConfigDTO {
	return api.ProviderConfigDTO{
		ID: cfg.ID, Name: cfg.Name, Platform: cfg.Platform,
		BaseURL: cfg.BaseURL, CredentialID: cfg.CredentialID,
		HasWebhookSecret: cfg.WebhookSecret != "",
		WebhookEndpoint:  cfg.WebhookEndpoint,
		SkipTLS:          cfg.SkipTLS,
		CreatedAt:        cfg.CreatedAt, UpdatedAt: cfg.UpdatedAt,
	}
}

// List .
// @router /api/v1/providers [GET]
func List(ctx context.Context, c *app.RequestContext) {
	dao := db.NewProviderConfigDAO()
	configs, err := dao.FindAll()
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to fetch providers: "+err.Error())
		return
	}
	// Batch-resolve credential names in one query (avoids N+1 per provider).
	credIDs := make([]uint, 0, len(configs))
	for _, cfg := range configs {
		if cfg.CredentialID > 0 {
			credIDs = append(credIDs, cfg.CredentialID)
		}
	}
	credNames, _ := db.NewCredentialDAO().FindNamesMap(credIDs)

	result := make([]*providerModel.ProviderConfig, 0, len(configs))
	for _, cfg := range configs {
		dto := toProtoProviderConfig(&cfg)
		if cfg.CredentialID > 0 {
			dto.CredentialName = credNames[cfg.CredentialID]
		}
		result = append(result, dto)
	}
	pkgresponse.Success(c, result)
}

// Get .
// @router /api/v1/providers/:id [GET]
func Get(ctx context.Context, c *app.RequestContext) {
	id, ok := pkgresponse.ParseIDParam(c, "id")
	if !ok {
		return
	}
	dao := db.NewProviderConfigDAO()
	cfg, err := dao.FindByID(id)
	if err != nil {
		pkgresponse.NotFound(c, "Provider config not found")
		return
	}
	dto := toProtoProviderConfig(cfg)
	if cfg.CredentialID > 0 {
		credDAO := db.NewCredentialDAO()
		if cred, err := credDAO.FindByID(cfg.CredentialID); err == nil {
			dto.CredentialName = cred.Name
		}
	}
	pkgresponse.Success(c, dto)
}

// Create .
// @router /api/v1/providers [POST]
func Create(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.CreateProviderConfigReq) (*providerModel.ProviderConfig, error) {
		if req.Name == "" || req.Platform == "" {
			return nil, handler.ErrBadRequest("name and platform are required")
		}
		if req.Platform != "gitlab" && req.Platform != "github" && req.Platform != "gitea" && req.Platform != "forgejo" && req.Platform != "tencent_code" && req.Platform != "gitee" {
			return nil, handler.ErrBadRequest("platform must be gitlab, github, gitea, forgejo, tencent_code or gitee")
		}
		if req.CredentialID == 0 {
			return nil, handler.ErrBadRequest("credential_id is required")
		}
		credDAO := db.NewCredentialDAO()
		if _, err := credDAO.FindByID(req.CredentialID); err != nil {
			return nil, handler.ErrBadRequest("credential not found")
		}
		dao := db.NewProviderConfigDAO()
		cfg := &po.ProviderConfig{
			Name: req.Name, Platform: req.Platform, BaseURL: req.BaseURL,
			CredentialID: req.CredentialID, WebhookSecret: req.WebhookSecret,
			SkipTLS: req.SkipTLS,
		}
		if err := dao.Create(cfg); err != nil {
			return nil, handler.ErrInternal("Failed to create provider config: " + err.Error())
		}
		c.Set("audit_details", map[string]string{"name": req.Name, "platform": req.Platform})
		return toProtoProviderConfig(cfg), nil
	})
}

// Update .
// @router /api/v1/providers/:id [PUT]
func Update(ctx context.Context, c *app.RequestContext) {
	id, ok := pkgresponse.ParseIDParam(c, "id")
	if !ok {
		return
	}
	handler.BindAndDo(c, func(req *api.UpdateProviderConfigReq) (*providerModel.ProviderConfig, error) {
		dao := db.NewProviderConfigDAO()
		cfg, err := dao.FindByID(id)
		if err != nil {
			return nil, handler.ErrNotFound("Provider config not found")
		}
		if req.Name != "" {
			cfg.Name = req.Name
		}
		if req.BaseURL != "" {
			cfg.BaseURL = req.BaseURL
		}
		if req.CredentialID > 0 {
			credDAO := db.NewCredentialDAO()
			if _, err := credDAO.FindByID(req.CredentialID); err != nil {
				return nil, handler.ErrBadRequest("credential not found")
			}
			cfg.CredentialID = req.CredentialID
		}
		if req.WebhookSecret != "" {
			cfg.WebhookSecret = req.WebhookSecret
		}
		if req.SkipTLS != nil {
			cfg.SkipTLS = *req.SkipTLS
		}
		if err := dao.Save(cfg); err != nil {
			return nil, handler.ErrInternal("Failed to update provider config: " + err.Error())
		}
		provider_manager.GetManager().Invalidate(id)
		return toProtoProviderConfig(cfg), nil
	})
}

// Delete .
// @router /api/v1/providers/:id [DELETE]
func Delete(ctx context.Context, c *app.RequestContext) {
	id, ok := pkgresponse.ParseIDParam(c, "id")
	if !ok {
		return
	}
	dao := db.NewProviderConfigDAO()
	if _, err := dao.FindByID(id); err != nil {
		pkgresponse.NotFound(c, "Provider config not found")
		return
	}
	bindingDAO := db.NewRepoProviderBindingDAO()
	bindings, _ := bindingDAO.FindByProviderConfigID(id)
	if len(bindings) > 0 {
		pkgresponse.BadRequest(c, fmt.Sprintf("Provider config is referenced by %d binding(s), please remove bindings first", len(bindings)))
		return
	}
	if err := dao.Delete(id); err != nil {
		pkgresponse.InternalServerError(c, "Failed to delete provider config: "+err.Error())
		return
	}
	provider_manager.GetManager().Invalidate(id)
	pkgresponse.Success(c, map[string]string{"message": "Provider config deleted"})
}

// Test .
// @router /api/v1/providers/:id/test [POST]
func Test(ctx context.Context, c *app.RequestContext) {
	id, ok := pkgresponse.ParseIDParam(c, "id")
	if !ok {
		return
	}
	p, err := provider_manager.GetManager().GetProvider(id)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to get provider: "+err.Error())
		return
	}
	result, err := p.TestConnection(ctx)
	if err != nil {
		pkgresponse.InternalServerError(c, "Test connection failed: "+err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

// ListRemoteRepos .
// @router /api/v1/providers/:id/repos [GET]
func ListRemoteRepos(ctx context.Context, c *app.RequestContext) {
	id, ok := pkgresponse.ParseIDParam(c, "id")
	if !ok {
		return
	}
	p, err := provider_manager.GetManager().GetProvider(id)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to get provider: "+err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "50"))
	owner := c.Query("owner")

	opts := provider.ListRepoOptions{
		Owner:   owner,
		Page:    page,
		PerPage: perPage,
	}

	repos, err := p.ListRepos(ctx, opts)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to list repos: "+err.Error())
		return
	}

	type repoDTO struct {
		ID            int64  `json:"id"`
		Name          string `json:"name"`
		FullName      string `json:"full_name"`
		Owner         string `json:"owner"`
		Description   string `json:"description"`
		CloneURL      string `json:"clone_url"`
		SSHURL        string `json:"ssh_url"`
		DefaultBranch string `json:"default_branch"`
		Private       bool   `json:"private"`
		Platform      string `json:"platform"`
	}

	result := make([]repoDTO, 0, len(repos))
	for _, r := range repos {
		result = append(result, repoDTO{
			ID: r.ID, Name: r.Name, FullName: r.FullName, Owner: r.Owner,
			Description: r.Description, CloneURL: r.CloneURL, SSHURL: r.SSHURL,
			DefaultBranch: r.DefaultBranch, Private: r.Private, Platform: string(r.Platform),
		})
	}

	pkgresponse.Success(c, result)
}

// ListRemoteBranches .
// @router /api/v1/providers/branches [GET]
func ListRemoteBranches(ctx context.Context, c *app.RequestContext) {
	providerIDStr := c.Query("provider_id")
	owner := c.Query("owner")
	repo := c.Query("repo")
	if providerIDStr == "" || owner == "" || repo == "" {
		pkgresponse.BadRequest(c, "provider_id, owner and repo are required")
		return
	}
	pid, _ := strconv.ParseUint(providerIDStr, 10, 64)
	p, err := provider_manager.GetManager().GetProvider(uint(pid))
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to get provider: "+err.Error())
		return
	}
	branches, err := p.ListBranches(ctx, owner, repo)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to list branches: "+err.Error())
		return
	}
	type branchDTO struct {
		Name string `json:"name"`
	}
	result := make([]branchDTO, 0, len(branches))
	for _, b := range branches {
		result = append(result, branchDTO{Name: b.Name})
	}
	pkgresponse.Success(c, result)
}

// CreateRemoteBranch .
// @router /api/v1/providers/branches/create [POST]
func CreateRemoteBranch(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *struct {
		ProviderID uint   `json:"provider_id"`
		Owner      string `json:"owner"`
		Repo       string `json:"repo"`
		Branch     string `json:"branch"`
		Ref        string `json:"ref"`
	}) (any, error) {
		if req.ProviderID == 0 || req.Owner == "" || req.Repo == "" || req.Branch == "" {
			return nil, handler.ErrBadRequest("provider_id, owner, repo and branch are required")
		}
		if req.Ref == "" {
			req.Ref = "main"
		}
		p, err := provider_manager.GetManager().GetProvider(req.ProviderID)
		if err != nil {
			return nil, handler.ErrInternal("Failed to get provider: " + err.Error())
		}
		br, err := p.CreateBranch(ctx, req.Owner, req.Repo, req.Branch, req.Ref)
		if err != nil {
			return nil, handler.ErrInternal("Failed to create branch: " + err.Error())
		}
		c.Set("audit_target", fmt.Sprintf("provider:%d:%s/%s", req.ProviderID, req.Owner, req.Repo))
		c.Set("audit_details", map[string]string{"branch": req.Branch, "ref": req.Ref})
		return br, nil
	})
}

// DeleteRemoteBranch .
// @router /api/v1/providers/branches/delete [POST]
func DeleteRemoteBranch(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *struct {
		ProviderID uint   `json:"provider_id"`
		Owner      string `json:"owner"`
		Repo       string `json:"repo"`
		Branch     string `json:"branch"`
	}) (map[string]string, error) {
		if req.ProviderID == 0 || req.Owner == "" || req.Repo == "" || req.Branch == "" {
			return nil, handler.ErrBadRequest("provider_id, owner, repo and branch are required")
		}
		p, err := provider_manager.GetManager().GetProvider(req.ProviderID)
		if err != nil {
			return nil, handler.ErrInternal("Failed to get provider: " + err.Error())
		}
		if err := p.DeleteBranch(ctx, req.Owner, req.Repo, req.Branch); err != nil {
			return nil, handler.ErrInternal("Failed to delete branch: " + err.Error())
		}
		c.Set("audit_target", fmt.Sprintf("provider:%d:%s/%s", req.ProviderID, req.Owner, req.Repo))
		c.Set("audit_details", map[string]string{"branch": req.Branch})
		return map[string]string{"message": "deleted"}, nil
	})
}
