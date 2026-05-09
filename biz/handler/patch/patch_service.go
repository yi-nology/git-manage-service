package patch

import (
	"context"
	"path/filepath"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	common "github.com/yi-nology/git-manage-service/biz/model/common"
	patchModel "github.com/yi-nology/git-manage-service/biz/model/patch"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func toProtoPatchInfo(name, path, modTime string, size int64, sequence int32, isApplied, canApply, hasConflict bool) *patchModel.PatchInfo {
	return &patchModel.PatchInfo{
		Name:        &name,
		Path:        &path,
		Size:        &size,
		ModTime:     &modTime,
		Sequence:    &sequence,
		IsApplied:   &isApplied,
		CanApply:    &canApply,
		HasConflict: &hasConflict,
	}
}

// GeneratePatch 生成 patch
// @router /api/v1/patch/generate [POST]
func GeneratePatch(ctx context.Context, c *app.RequestContext) {
	var req api.GeneratePatchReq
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

	var patchContent string
	if len(req.Commits) > 0 {
		// 为指定的 commit 列表生成 patch
		patchContent, err = gitSvc.GeneratePatchForCommits(repo.Path, req.Commits)
	} else {
		// 为 base..target 生成 patch
		if req.Base == "" || req.Target == "" {
			response.BadRequest(c, "base and target are required when commits is empty")
			return
		}
		patchContent, err = gitSvc.GeneratePatch(repo.Path, req.Base, req.Target)
	}

	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]string{
		"content": patchContent,
	})
}

// SavePatch 保存 patch 到仓库
// @router /api/v1/patch/save [POST]
func SavePatch(ctx context.Context, c *app.RequestContext) {
	var req api.SavePatchReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.PatchName == "" {
		response.BadRequest(c, "patch_name is required")
		return
	}

	if req.PatchContent == "" {
		response.BadRequest(c, "patch_content is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	savedPath, err := gitSvc.SavePatch(repo.Path, req.PatchContent, req.PatchName, req.CustomPath, req.CommitMessage)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]string{
		"path": savedPath,
		"name": filepath.Base(savedPath),
	})
}

// ListPatches 列出仓库中的所有 patch
// @router /api/v1/patch/list [GET]
func ListPatches(ctx context.Context, c *app.RequestContext) {
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

	gitSvc := git.NewGitService()
	patches, err := gitSvc.ListPatches(repo.Path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	var dtos []*patchModel.PatchInfo
	for _, p := range patches {
		dtos = append(dtos, toProtoPatchInfo(
			p.Name, p.Path, p.ModTime, p.Size,
			int32(p.Sequence), p.IsApplied, p.CanApply, p.HasConflict,
		))
	}

	response.Success(c, dtos)
}

// GetPatchContent 获取 patch 内容
// @router /api/v1/patch/content [GET]
func GetPatchContent(ctx context.Context, c *app.RequestContext) {
	patchPath := c.Query("path")
	if patchPath == "" {
		response.BadRequest(c, "path is required")
		return
	}

	gitSvc := git.NewGitService()
	content, err := gitSvc.GetPatchContent(patchPath)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]string{
		"content": content,
	})
}

// DownloadPatch 下载 patch 文件
// @router /api/v1/patch/download [GET]
func DownloadPatch(ctx context.Context, c *app.RequestContext) {
	patchPath := c.Query("path")
	if patchPath == "" {
		response.BadRequest(c, "path is required")
		return
	}

	gitSvc := git.NewGitService()
	content, err := gitSvc.GetPatchContent(patchPath)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	// 设置下载头
	fileName := filepath.Base(patchPath)
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename=\""+fileName+"\"")
	c.String(200, content)
}

// ApplyPatch 应用 patch
// @router /api/v1/patch/apply [POST]
func ApplyPatch(ctx context.Context, c *app.RequestContext) {
	var req api.ApplyPatchReq
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

	var applyErr error
	if req.PatchContent != "" {
		applyErr = gitSvc.ApplyPatchFromContent(repo.Path, req.PatchContent, req.SignOff, req.CommitMessage)
	} else if req.PatchPath != "" {
		applyErr = gitSvc.ApplyPatch(repo.Path, req.PatchPath, req.SignOff, req.CommitMessage)
	} else {
		response.BadRequest(c, "patch_path or patch_content is required")
		return
	}

	if applyErr != nil {
		response.InternalServerError(c, applyErr.Error())
		return
	}

	response.Success(c, map[string]string{
		"message": "patch applied successfully",
	})
}

// CheckPatch 检查 patch 是否可以应用
// @router /api/v1/patch/check [POST]
func CheckPatch(ctx context.Context, c *app.RequestContext) {
	var req struct {
		RepoKey   string `json:"repo_key" form:"repo_key"`
		PatchPath string `json:"patch_path" form:"patch_path"`
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
	stats, err := gitSvc.GetPatchStats(repo.Path, req.PatchPath)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	statStr, _ := stats["stat"].(string)
	canApply, _ := stats["can_apply"].(bool)
	var errStr string
	if e, ok := stats["error"].(string); ok {
		errStr = e
	}

	dto := &patchModel.PatchStats{
		Stat:     &statStr,
		CanApply: &canApply,
		Error:    &errStr,
	}

	response.Success(c, dto)
}

// DeletePatch 删除 patch
// @router /api/v1/patch/delete [POST]
func DeletePatch(ctx context.Context, c *app.RequestContext) {
	var req api.DeletePatchReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	gitSvc := git.NewGitService()
	if err := gitSvc.DeletePatch(req.PatchPath); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]string{
		"message": "patch deleted",
	})
}
