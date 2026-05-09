package workspace

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	common "github.com/yi-nology/git-manage-service/biz/model/common"
	workspace "github.com/yi-nology/git-manage-service/biz/model/workspace"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/biz/service/llm"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func GetWorkspaceStatus(ctx context.Context, c *app.RequestContext) {
	var req workspace.GetWorkspaceStatusReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	result, err := gitSvc.GetWorkspaceStatus(repo.Path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}

func PushCurrent(ctx context.Context, c *app.RequestContext) {
	var req struct {
		RepoKey string `json:"repo_key" vd:"len($)>0"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	if err := gitSvc.PushCurrent(repo.Path); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func RemoveTracking(ctx context.Context, c *app.RequestContext) {
	var req struct {
		RepoKey string   `json:"repo_key" vd:"len($)>0"`
		Files   []string `json:"files"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	if err := gitSvc.RemoveTracking(repo.Path, req.Files); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func AddToGitignore(ctx context.Context, c *app.RequestContext) {
	var req struct {
		RepoKey  string   `json:"repo_key" vd:"len($)>0"`
		Patterns []string `json:"patterns"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	if err := gitSvc.AddToGitignore(repo.Path, req.Patterns); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func GetWorkspaceDiff(ctx context.Context, c *app.RequestContext) {
	var req workspace.GetWorkspaceDiffReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	result, err := gitSvc.GetWorkspaceDiff(repo.Path, req.File, req.StagedOnly)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}

func StageFiles(ctx context.Context, c *app.RequestContext) {
	var req workspace.StageFilesReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	err = gitSvc.StageFiles(repo.Path, req.Files, req.StageAll)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func UnstageFiles(ctx context.Context, c *app.RequestContext) {
	var req workspace.UnstageFilesReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	err = gitSvc.UnstageFiles(repo.Path, req.Files, req.UnstageAll)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func CommitChanges(ctx context.Context, c *app.RequestContext) {
	var req workspace.CommitChangesReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Message == "" {
		response.BadRequest(c, "commit message is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	result, err := gitSvc.CommitChanges(repo.Path, req.Files, req.StageAll, req.Message, req.AuthorName, req.AuthorEmail, req.Push, req.PushRemote)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}

func PullWithResolve(ctx context.Context, c *app.RequestContext) {
	var req workspace.PullWithResolveReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	result, err := gitSvc.PullWithResolve(repo.Path, req.Remote, req.Branch, req.FetchOnly)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}

func GetConflictDetail(ctx context.Context, c *app.RequestContext) {
	var req workspace.GetConflictDetailReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	result, err := gitSvc.GetConflictDetail(repo.Path, req.File)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}

func MarkConflictResolved(ctx context.Context, c *app.RequestContext) {
	var req workspace.MarkConflictResolvedReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	err = gitSvc.MarkConflictResolved(repo.Path, req.File, req.ResolvedContent, req.Stage)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, nil)
}

func AIResolveConflict(ctx context.Context, c *app.RequestContext) {
	var req workspace.AIResolveConflictReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	aiSvc := git.NewWorkspaceAIService()
	result, err := aiSvc.AIResolveConflict(req.OursContent, req.TheirsContent, req.BaseContent, req.File, req.Hint)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, result)
}

func GenerateCommitMessage(ctx context.Context, c *app.RequestContext) {
	var req struct {
		RepoKey string `json:"repo_key"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	diffOutput, err := gitSvc.GetWorkspaceDiffRaw(repo.Path, "")
	if err != nil {
		response.InternalServerError(c, "获取 diff 失败: "+err.Error())
		return
	}

	if diffOutput == "" {
		response.BadRequest(c, "没有变更可以生成提交信息")
		return
	}

	provider, err := llm.GetDefaultProvider()
	if err != nil {
		response.InternalServerError(c, "未配置 LLM 提供商: "+err.Error())
		return
	}

	truncated := diffOutput
	if len(truncated) > 8000 {
		truncated = truncated[:8000] + "\n... (truncated)"
	}

	prompt := fmt.Sprintf(`根据以下 git diff 变更内容，生成一条简洁的 commit message。
要求：
1. 使用中文
2. 不超过一行，不超过72个字符
3. 使用 "类型: 简短描述" 格式，类型包括: feat/fix/refactor/docs/style/test/chore
4. 只输出 commit message 本身，不要任何解释或额外文字

Diff 内容:
%s`, truncated)

	resp, err := provider.Chat(ctx, &llm.ChatRequest{
		Messages: []llm.ChatMessage{{Role: "user", Content: prompt}},
	})
	if err != nil {
		response.InternalServerError(c, "AI 生成失败: "+err.Error())
		return
	}

	msg := resp.Content
	if len(msg) > 200 {
		msg = msg[:200]
	}

	response.Success(c, map[string]string{"message": msg})
}
