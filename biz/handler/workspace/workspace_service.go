package workspace

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/model/workspace"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/biz/service/llm"
	"github.com/yi-nology/git-manage-service/pkg/handler"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

type PushCurrentReq struct {
	RepoKey string `json:"repo_key" vd:"len($)>0"`
}

type RemoveTrackingReq struct {
	RepoKey string   `json:"repo_key" vd:"len($)>0"`
	Files   []string `json:"files"`
}

type AddToGitignoreReq struct {
	RepoKey  string   `json:"repo_key" vd:"len($)>0"`
	Patterns []string `json:"patterns"`
}

type GenerateCommitMessageReq struct {
	RepoKey string `json:"repo_key"`
}

func GetWorkspaceStatus(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepo(c,
		func(req *workspace.GetWorkspaceStatusReq) string { return req.RepoKey },
		func(repo *po.Repo, req *workspace.GetWorkspaceStatusReq) (any, error) {
			return git.NewGitService().GetWorkspaceStatus(repo.Path)
		},
	)
}

func PushCurrent(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepoVoid(c,
		func(req *PushCurrentReq) string { return req.RepoKey },
		func(repo *po.Repo, req *PushCurrentReq) error {
			return git.NewGitService().PushCurrent(repo.Path)
		},
	)
}

func RemoveTracking(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepoVoid(c,
		func(req *RemoveTrackingReq) string { return req.RepoKey },
		func(repo *po.Repo, req *RemoveTrackingReq) error {
			return git.NewGitService().RemoveTracking(repo.Path, req.Files)
		},
	)
}

func AddToGitignore(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepoVoid(c,
		func(req *AddToGitignoreReq) string { return req.RepoKey },
		func(repo *po.Repo, req *AddToGitignoreReq) error {
			return git.NewGitService().AddToGitignore(repo.Path, req.Patterns)
		},
	)
}

func GetWorkspaceDiff(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepo(c,
		func(req *workspace.GetWorkspaceDiffReq) string { return req.RepoKey },
		func(repo *po.Repo, req *workspace.GetWorkspaceDiffReq) (any, error) {
			return git.NewGitService().GetWorkspaceDiff(repo.Path, req.File, req.StagedOnly)
		},
	)
}

func StageFiles(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepoVoid(c,
		func(req *workspace.StageFilesReq) string { return req.RepoKey },
		func(repo *po.Repo, req *workspace.StageFilesReq) error {
			return git.NewGitService().StageFiles(repo.Path, req.Files, req.StageAll)
		},
	)
}

func UnstageFiles(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepoVoid(c,
		func(req *workspace.UnstageFilesReq) string { return req.RepoKey },
		func(repo *po.Repo, req *workspace.UnstageFilesReq) error {
			return git.NewGitService().UnstageFiles(repo.Path, req.Files, req.UnstageAll)
		},
	)
}

func CommitChanges(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepo(c,
		func(req *workspace.CommitChangesReq) string { return req.RepoKey },
		func(repo *po.Repo, req *workspace.CommitChangesReq) (any, error) {
			if req.Message == "" {
				return nil, handler.ErrBadRequest("commit message is required")
			}
			return git.NewGitService().CommitChanges(repo.Path, req.Files, req.StageAll, req.Message, req.AuthorName, req.AuthorEmail, req.Push, req.PushRemote)
		},
	)
}

func PullWithResolve(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepo(c,
		func(req *workspace.PullWithResolveReq) string { return req.RepoKey },
		func(repo *po.Repo, req *workspace.PullWithResolveReq) (any, error) {
			return git.NewGitService().PullWithResolve(repo.Path, req.Remote, req.Branch, req.FetchOnly)
		},
	)
}

func GetConflictDetail(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepo(c,
		func(req *workspace.GetConflictDetailReq) string { return req.RepoKey },
		func(repo *po.Repo, req *workspace.GetConflictDetailReq) (any, error) {
			return git.NewGitService().GetConflictDetail(repo.Path, req.File)
		},
	)
}

func MarkConflictResolved(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepoVoid(c,
		func(req *workspace.MarkConflictResolvedReq) string { return req.RepoKey },
		func(repo *po.Repo, req *workspace.MarkConflictResolvedReq) error {
			return git.NewGitService().MarkConflictResolved(repo.Path, req.File, req.ResolvedContent, req.Stage)
		},
	)
}

func AIResolveConflict(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c,
		func(req *workspace.AIResolveConflictReq) (any, error) {
			aiSvc := git.NewWorkspaceAIService()
			return aiSvc.AIResolveConflict(req.OursContent, req.TheirsContent, req.BaseContent, req.File, req.Hint)
		},
	)
}

func GenerateCommitMessage(ctx context.Context, c *app.RequestContext) {
	var req GenerateCommitMessageReq
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
