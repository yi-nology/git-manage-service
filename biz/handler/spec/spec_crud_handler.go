package spec

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	gitSvc "github.com/yi-nology/git-manage-service/biz/service/git"
	lintSvc "github.com/yi-nology/git-manage-service/biz/service/lint"
	specService "github.com/yi-nology/git-manage-service/biz/service/spec"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func GetSpecContentByPath(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	path := c.Param("path")
	if repoKey == "" {
		response.BadRequest(c, "repo_key is required")
		return
	}
	if path == "" {
		path = c.Query("path")
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := specService.NewSpecService()
	content, err := svc.GetSpecContent(repo.Path, path)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, api.FileContent{
		Path:    path,
		Content: content,
	})
}

func SaveSpecContentByPath(ctx context.Context, c *app.RequestContext) {
	path := c.Param("path")
	if path == "" {
		response.BadRequest(c, "path is required")
		return
	}

	var req api.SaveSpecContentReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Path == "" {
		req.Path = path
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := specService.NewSpecService()

	err = svc.SaveSpecContent(repo.Path, req.Path, req.Content, req.Message)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if req.AutoCommit && req.Message != "" {
		gitService := gitSvc.NewGitService()
		if err := gitService.AddAndCommit(repo.Path, req.Path, req.Message); err != nil {
			response.InternalError(c, err)
			return
		}
	}

	c.Set("audit_target", "repo:"+repo.Key)
	c.Set("audit_details", map[string]string{"path": req.Path, "message": req.Message})
	response.Success(c, api.SaveSpecResponse{
		Message: "spec saved successfully",
		Path:    req.Path,
	})
}

func CommitSpec(ctx context.Context, c *app.RequestContext) {
	path := c.Param("path")
	if path == "" {
		response.BadRequest(c, "path is required")
		return
	}

	var req api.CommitSpecReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Message == "" {
		response.BadRequest(c, "message is required")
		return
	}

	if req.Path == "" {
		req.Path = path
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	if req.Content != "" {
		svc := specService.NewSpecService()
		if err := svc.SaveSpecContent(repo.Path, req.Path, req.Content, ""); err != nil {
			response.InternalError(c, err)
			return
		}
	}

	gitService := gitSvc.NewGitService()
	if err := gitService.AddAndCommit(repo.Path, req.Path, req.Message); err != nil {
		response.InternalError(c, err)
		return
	}

	c.Set("audit_target", "repo:"+repo.Key)
	c.Set("audit_details", map[string]string{"path": req.Path, "message": req.Message})
	response.Success(c, api.CommitResponse{
		Message: "committed successfully",
	})
}

func ListSpecFiles(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	if repoKey == "" {
		response.BadRequest(c, "repo_key is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := specService.NewSpecService()
	files, err := svc.ListSpecFiles(repo.Path)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if files == nil {
		files = []specService.SpecFileInfo{}
	}

	var dtos []api.SpecFileInfo
	for _, f := range files {
		dtos = append(dtos, api.SpecFileInfo{
			Name:    f.Name,
			Path:    f.Path,
			IsDir:   f.IsDir,
			Size:    f.Size,
			ModTime: f.ModTime,
		})
	}

	response.Success(c, dtos)
}

func GetSpecContent(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	path := c.Query("path")
	if repoKey == "" || path == "" {
		response.BadRequest(c, "repo_key and path are required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := specService.NewSpecService()
	content, err := svc.GetSpecContent(repo.Path, path)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, api.SpecContentResponse{
		Content: content,
		Path:    path,
	})
}

func SaveSpecContent(ctx context.Context, c *app.RequestContext) {
	var req api.SaveSpecReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	lintService := lintSvc.NewLintService()
	validationResult, err := lintService.Lint(req.Content, nil)
	if err == nil && validationResult != nil {
		for _, issue := range validationResult.Issues {
			if issue.Severity == "error" {
				response.BadRequest(c, "Spec validation failed: "+issue.Message)
				return
			}
		}
	}

	svc := specService.NewSpecService()

	err = svc.SaveSpecContent(repo.Path, req.Path, req.Content, req.CommitMessage)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if req.CommitMessage != "" {
		gitService := gitSvc.NewGitService()
		if err := gitService.AddAndCommit(repo.Path, req.Path, req.CommitMessage); err != nil {
			response.InternalError(c, err)
			return
		}
	}

	c.Set("audit_target", "repo:"+repo.Key)
	c.Set("audit_details", map[string]string{"path": req.Path, "commit_message": req.CommitMessage})
	response.Success(c, api.SaveWithValidationResponse{
		Message:          "spec saved successfully",
		ValidationResult: validationResult,
	})
}

func CreateSpecFile(ctx context.Context, c *app.RequestContext) {
	var req api.CreateSpecFileReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := specService.NewSpecService()
	path, err := svc.CreateSpecFileWithContent(repo.Path, req.Path, req.Name, req.Content)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	c.Set("audit_target", "repo:"+repo.Key)
	c.Set("audit_details", map[string]string{"path": path})
	response.Success(c, api.CreateFileResponse{
		Path:    path,
		Message: "Spec 文件创建成功",
	})
}

func DeleteSpecFile(ctx context.Context, c *app.RequestContext) {
	var req api.DeleteSpecFileReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := specService.NewSpecService()
	err = svc.DeleteSpecFile(repo.Path, req.Path)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if req.CommitMessage != "" {
		gitService := gitSvc.NewGitService()
		if err := gitService.RemoveAndCommit(repo.Path, req.Path, req.CommitMessage); err != nil {
			response.InternalError(c, err)
			return
		}
	}

	c.Set("audit_target", "repo:"+repo.Key)
	c.Set("audit_details", map[string]string{"path": req.Path})
	response.Success(c, api.MessageResponse{
		Message: "spec deleted",
	})
}
