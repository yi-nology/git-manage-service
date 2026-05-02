package author

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/hertz/pkg/app"
	author "github.com/yi-nology/git-manage-service/biz/model/author"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func ListIdentities(ctx context.Context, c *app.RequestContext) {
	var req author.ListIdentitiesRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	svc := git.NewAuthorService()
	identities, err := svc.ListIdentities()
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, identities)
}

func CreateIdentity(ctx context.Context, c *app.RequestContext) {
	var req author.CreateIdentityRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	var aliases []api.AliasEntry
	if req.GetAliasesJson() != "" {
		json.Unmarshal([]byte(req.GetAliasesJson()), &aliases)
	}
	svc := git.NewAuthorService()
	result, err := svc.CreateIdentity(api.CreateIdentityRequest{
		CanonicalName:  req.GetCanonicalName(),
		CanonicalEmail: req.GetCanonicalEmail(),
		Aliases:        aliases,
	})
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, result)
}

func UpdateIdentity(ctx context.Context, c *app.RequestContext) {
	var req author.UpdateIdentityRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	var aliases []api.AliasEntry
	if req.GetAliasesJson() != "" {
		json.Unmarshal([]byte(req.GetAliasesJson()), &aliases)
	}
	svc := git.NewAuthorService()
	result, err := svc.UpdateIdentity(uint(req.GetId()), api.UpdateIdentityRequest{
		CanonicalName:  req.GetCanonicalName(),
		CanonicalEmail: req.GetCanonicalEmail(),
		Aliases:        aliases,
	})
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, result)
}

func DeleteIdentity(ctx context.Context, c *app.RequestContext) {
	var req author.DeleteIdentityRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	svc := git.NewAuthorService()
	if err := svc.DeleteIdentity(uint(req.GetId())); err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, nil)
}

func ActivateIdentity(ctx context.Context, c *app.RequestContext) {
	var req author.ActivateIdentityRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	svc := git.NewAuthorService()
	if err := svc.ActivateIdentity(uint(req.GetId())); err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, nil)
}

func GetRepoAuthorConfig(ctx context.Context, c *app.RequestContext) {
	var req author.RepoAuthorConfigRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	svc := git.NewAuthorService()
	config, err := svc.GetRepoAuthorConfig(req.GetRepoKey())
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, config)
}

func SetRepoAuthorConfig(ctx context.Context, c *app.RequestContext) {
	var req author.SetRepoAuthorConfigRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	svc := git.NewAuthorService()
	var identityID *uint
	if req.IdentityId != nil && !req.GetClear() {
		id := uint(req.GetIdentityId())
		identityID = &id
	}
	if err := svc.SetRepoAuthorConfig(req.GetRepoKey(), identityID, req.GetClear()); err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, nil)
}

func ScanAuthor(ctx context.Context, c *app.RequestContext) {
	var req author.ScanAuthorRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}
	svc := git.NewAuthorService()
	result, err := svc.ScanAuthor(repo.Path)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, result)
}

func FixAuthorAll(ctx context.Context, c *app.RequestContext) {
	var req author.FixAuthorAllRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}
	taskID, err := git.StartAuthorFixTask(repo.ID, repo.Path, nil, req.GetPushRemote())
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, api.MaintenanceTaskResponse{TaskID: taskID})
}

func FixAuthor(ctx context.Context, c *app.RequestContext) {
	var req author.FixAuthorRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}
	taskID, err := git.StartAuthorFixTask(repo.ID, repo.Path, req.GetCommitHashes(), req.GetPushRemote())
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, api.MaintenanceTaskResponse{TaskID: taskID})
}
