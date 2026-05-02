package author

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

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
