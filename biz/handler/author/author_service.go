package author

import (
	"context"
	"encoding/json"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	author "github.com/yi-nology/git-manage-service/biz/model/author"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/handler"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func ListIdentities(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c,
		func(req *author.ListIdentitiesRequest) ([]api.AuthorIdentityDTO, error) {
			svc := git.NewAuthorService()
			return svc.ListIdentities()
		},
	)
}

func CreateIdentity(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c,
		func(req *author.CreateIdentityRequest) (*api.AuthorIdentityDTO, error) {
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
				return nil, handler.ErrInternal(err.Error())
			}
			c.Set("audit_details", map[string]string{"name": req.GetCanonicalName(), "email": req.GetCanonicalEmail()})
			return result, nil
		},
	)
}

func UpdateIdentity(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c,
		func(req *author.UpdateIdentityRequest) (*api.AuthorIdentityDTO, error) {
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
				return nil, handler.ErrInternal(err.Error())
			}
			return result, nil
		},
	)
}

func DeleteIdentity(ctx context.Context, c *app.RequestContext) {
	handler.Do(c,
		func(req *author.DeleteIdentityRequest) error {
			svc := git.NewAuthorService()
			if err := svc.DeleteIdentity(uint(req.GetId())); err != nil {
				return handler.ErrInternal(err.Error())
			}
			return nil
		},
	)
}

func ActivateIdentity(ctx context.Context, c *app.RequestContext) {
	handler.Do(c,
		func(req *author.ActivateIdentityRequest) error {
			svc := git.NewAuthorService()
			if err := svc.ActivateIdentity(uint(req.GetId())); err != nil {
				return handler.ErrInternal(err.Error())
			}
			return nil
		},
	)
}

func GetRepoAuthorConfig(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c,
		func(req *author.RepoAuthorConfigRequest) (*api.RepoAuthorConfigDTO, error) {
			svc := git.NewAuthorService()
			config, err := svc.GetRepoAuthorConfig(req.GetRepoKey())
			if err != nil {
				return nil, handler.ErrInternal(err.Error())
			}
			return config, nil
		},
	)
}

func SetRepoAuthorConfig(ctx context.Context, c *app.RequestContext) {
	handler.Do(c,
		func(req *author.SetRepoAuthorConfigRequest) error {
			svc := git.NewAuthorService()
			var identityID *uint
			if req.IdentityId != nil && !req.GetClear() {
				id := uint(req.GetIdentityId())
				identityID = &id
			}
			if err := svc.SetRepoAuthorConfig(req.GetRepoKey(), identityID, req.GetClear()); err != nil {
				return handler.ErrInternal(err.Error())
			}
			c.Set("audit_details", map[string]string{"repo_key": req.GetRepoKey()})
			return nil
		},
	)
}

func ScanAuthor(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepo(c,
		func(req *author.ScanAuthorRequest) string { return req.GetRepoKey() },
		func(repo *po.Repo, req *author.ScanAuthorRequest) (*api.AuthorScanResult, error) {
			svc := git.NewAuthorService()
			result, err := svc.ScanAuthor(repo.Path)
			if err != nil {
				return nil, handler.ErrInternal(err.Error())
			}
			return result, nil
		},
	)
}

func FixAuthorAll(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepo(c,
		func(req *author.FixAuthorAllRequest) string { return req.GetRepoKey() },
		func(repo *po.Repo, req *author.FixAuthorAllRequest) (api.MaintenanceTaskResponse, error) {
			taskID, err := git.StartAuthorFixTask(repo.ID, repo.Path, nil, req.GetPushRemote())
			if err != nil {
				return api.MaintenanceTaskResponse{}, handler.ErrInternal(err.Error())
			}
			return api.MaintenanceTaskResponse{TaskID: taskID}, nil
		},
	)
}

func FixAuthor(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepo(c,
		func(req *author.FixAuthorRequest) string { return req.GetRepoKey() },
		func(repo *po.Repo, req *author.FixAuthorRequest) (api.MaintenanceTaskResponse, error) {
			taskID, err := git.StartAuthorFixTask(repo.ID, repo.Path, req.GetCommitHashes(), req.GetPushRemote())
			if err != nil {
				return api.MaintenanceTaskResponse{}, handler.ErrInternal(err.Error())
			}
			return api.MaintenanceTaskResponse{TaskID: taskID}, nil
		},
	)
}

// AuthorAI .
// @router /api/v1/repo/:repo_key/author/ai [POST]
func AuthorAI(ctx context.Context, c *app.RequestContext) {
	var req api.AuthorAIRequest
	if err := c.BindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}

	svc := git.NewAuthorAIService()
	resp := &api.AuthorAIResponse{Action: req.Action}

	switch req.Action {
	case "suggest":
		repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
		if err != nil {
			response.BadRequest(c, "仓库不存在")
			return
		}
		result, err := svc.SmartSuggest(repo.Path)
		if err != nil {
			response.InternalError(c, err)
			return
		}
		resp.Suggest = result

	case "analyze":
		if req.Scan == nil {
			response.BadRequest(c, "缺少扫描结果")
			return
		}
		identities, err := db.NewAuthorIdentityDAO().ListAll()
		if err != nil {
			response.InternalError(c, err)
			return
		}
		var dtos []api.AuthorIdentityDTO
		for _, id := range identities {
			dtos = append(dtos, git.IdentityToDTO(&id))
		}
		result, err := svc.AnalyzeScan(req.Scan, dtos)
		if err != nil {
			response.InternalError(c, err)
			return
		}
		resp.Result = result

	case "merge":
		identities, err := db.NewAuthorIdentityDAO().ListAll()
		if err != nil {
			response.InternalError(c, err)
			return
		}
		var dtos []api.AuthorIdentityDTO
		for _, id := range identities {
			dtos = append(dtos, git.IdentityToDTO(&id))
		}
		result, err := svc.SuggestMerges(dtos)
		if err != nil {
			response.InternalError(c, err)
			return
		}
		resp.Merge = result

	case "risk":
		if len(req.Commits) == 0 {
			response.BadRequest(c, "缺少提交列表")
			return
		}
		repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
		if err != nil {
			response.BadRequest(c, "仓库不存在")
			return
		}
		result, err := svc.AssessRisk(req.Commits, repo.Path)
		if err != nil {
			response.InternalError(c, err)
			return
		}
		resp.Risk = result

	default:
		response.BadRequest(c, "不支持的操作: "+req.Action)
		return
	}

	response.Success(c, resp)
}

// AuthorChat .
// @router /api/v1/repo/:repo_key/author/chat [POST]
func AuthorChat(ctx context.Context, c *app.RequestContext) {
	var req api.AuthorChatRequest
	if err := c.BindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误")
		return
	}
	if req.Prompt == "" {
		response.BadRequest(c, "问题不能为空")
		return
	}

	svc := git.NewAuthorAIService()
	repoPath := ""
	var scan *api.AuthorScanResult

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err == nil {
		repoPath = repo.Path
		authorSvc := git.NewAuthorService()
		scan, _ = authorSvc.ScanAuthor(repoPath)
	}

	result, err := svc.AuthorChat(repoPath, req.Prompt, req.History, scan)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, &api.AuthorChatResponse{Result: result})
}
