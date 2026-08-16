package cr

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/service/crservice"
	"github.com/yi-nology/git-manage-service/pkg/handler"
	pkgresponse "github.com/yi-nology/git-manage-service/pkg/response"
)

func Create(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c,
		func(req *api.CreateCRReq) (*api.CRDTO, error) {
			if req.RepoKey == "" || req.Title == "" || req.SourceBranch == "" || req.TargetBranch == "" {
				return nil, handler.ErrBadRequest("repo_key, title, source_branch and target_branch are required")
			}
			cr, err := crservice.CreateCR(ctx, req)
			if err != nil {
				return nil, handler.ErrInternal("Failed to create CR: " + err.Error())
			}
			c.Set("audit_target", "repo:"+req.RepoKey)
			c.Set("audit_details", map[string]string{"title": req.Title, "source": req.SourceBranch, "target": req.TargetBranch})
			return cr, nil
		},
	)
}

func Get(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c,
		func(req *api.GetCRReq) (*api.CRDTO, error) {
			if req.RepoKey == "" || req.CRNumber == 0 {
				return nil, handler.ErrBadRequest("repo_key and cr_number are required")
			}
			cr, err := crservice.GetCR(ctx, req.RepoKey, req.CRNumber)
			if err != nil {
				return nil, handler.ErrInternal("Failed to get CR: " + err.Error())
			}
			return cr, nil
		},
	)
}

func List(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c,
		func(req *api.ListCRsReq) (map[string]interface{}, error) {
			if req.RepoKey == "" {
				return nil, handler.ErrBadRequest("repo_key is required")
			}
			if req.Page == 0 {
				req.Page = 1
			}
			if req.PageSize == 0 {
				req.PageSize = 20
			}
			crs, total, err := crservice.ListCRs(ctx, req.RepoKey, req.State, req.SourceBranch, req.TargetBranch, req.Page, req.PageSize)
			if err != nil {
				return nil, handler.ErrInternal("Failed to list CRs: " + err.Error())
			}
			return map[string]interface{}{
				"items": crs,
				"total": total,
			}, nil
		},
	)
}

func Merge(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c,
		func(req *api.MergeCRReq) (*api.CRDTO, error) {
			if req.RepoKey == "" || req.CRNumber == 0 {
				return nil, handler.ErrBadRequest("repo_key and cr_number are required")
			}
			cr, err := crservice.MergeCR(ctx, req.RepoKey, req.CRNumber, req.MergeCommitMessage, req.Squash, req.RemoveSourceBranch)
			if err != nil {
				return nil, handler.ErrInternal("Failed to merge CR: " + err.Error())
			}
			c.Set("audit_target", "repo:"+req.RepoKey)
			c.Set("audit_details", map[string]string{"cr_number": fmt.Sprintf("%d", req.CRNumber)})
			return cr, nil
		},
	)
}

func Close(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c,
		func(req *api.CloseCRReq) (*api.CRDTO, error) {
			if req.RepoKey == "" || req.CRNumber == 0 {
				return nil, handler.ErrBadRequest("repo_key and cr_number are required")
			}
			cr, err := crservice.CloseCR(ctx, req.RepoKey, req.CRNumber)
			if err != nil {
				return nil, handler.ErrInternal("Failed to close CR: " + err.Error())
			}
			c.Set("audit_target", "repo:"+req.RepoKey)
			return cr, nil
		},
	)
}

func Sync(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c,
		func(req *api.SyncCRsReq) (map[string]interface{}, error) {
			if req.RepoKey == "" {
				return nil, handler.ErrBadRequest("repo_key is required")
			}
			count, err := crservice.SyncCRs(ctx, req.RepoKey, req.State)
			if err != nil {
				return nil, handler.ErrInternal("Failed to sync CRs: " + err.Error())
			}
			c.Set("audit_target", "repo:"+req.RepoKey)
			return map[string]interface{}{"synced_count": count}, nil
		},
	)
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
	handler.BindAndDo(c,
		func(req *api.CreateCRByProviderReq) (*api.CRDTO, error) {
			if req.ProviderID == 0 || req.Owner == "" || req.Repo == "" || req.Title == "" || req.SourceBranch == "" || req.TargetBranch == "" {
				return nil, handler.ErrBadRequest("provider_id, owner, repo, title, source_branch and target_branch are required")
			}
			cr, err := crservice.CreateCRByProvider(ctx, req)
			if err != nil {
				return nil, handler.ErrInternal("Failed to create CR: " + err.Error())
			}
			c.Set("audit_target", fmt.Sprintf("provider:%d:%s/%s", req.ProviderID, req.Owner, req.Repo))
			return cr, nil
		},
	)
}

type mergeByProviderReq struct {
	ProviderID         uint   `json:"provider_id"`
	Owner              string `json:"owner"`
	Repo               string `json:"repo"`
	CRNumber           int    `json:"cr_number"`
	MergeCommitMessage string `json:"merge_commit_message"`
	Squash             bool   `json:"squash"`
	RemoveSourceBranch bool   `json:"remove_source_branch"`
}

func MergeByProvider(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c,
		func(req *mergeByProviderReq) (*api.CRDTO, error) {
			if req.ProviderID == 0 || req.Owner == "" || req.Repo == "" || req.CRNumber == 0 {
				return nil, handler.ErrBadRequest("provider_id, owner, repo and cr_number are required")
			}
			cr, err := crservice.MergeCRByProvider(ctx, req.ProviderID, req.Owner, req.Repo, req.CRNumber, req.MergeCommitMessage, req.Squash, req.RemoveSourceBranch)
			if err != nil {
				return nil, handler.ErrInternal("Failed to merge CR: " + err.Error())
			}
			c.Set("audit_target", fmt.Sprintf("provider:%d:%s/%s", req.ProviderID, req.Owner, req.Repo))
			return cr, nil
		},
	)
}

type closeByProviderReq struct {
	ProviderID uint   `json:"provider_id"`
	Owner      string `json:"owner"`
	Repo       string `json:"repo"`
	CRNumber   int    `json:"cr_number"`
}

func CloseByProvider(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c,
		func(req *closeByProviderReq) (*api.CRDTO, error) {
			if req.ProviderID == 0 || req.Owner == "" || req.Repo == "" || req.CRNumber == 0 {
				return nil, handler.ErrBadRequest("provider_id, owner, repo and cr_number are required")
			}
			cr, err := crservice.CloseCRByProvider(ctx, req.ProviderID, req.Owner, req.Repo, req.CRNumber)
			if err != nil {
				return nil, handler.ErrInternal("Failed to close CR: " + err.Error())
			}
			c.Set("audit_target", fmt.Sprintf("provider:%d:%s/%s", req.ProviderID, req.Owner, req.Repo))
			return cr, nil
		},
	)
}
