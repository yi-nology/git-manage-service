package patch

import (
	"context"
	"path/filepath"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	patchModel "github.com/yi-nology/git-manage-service/biz/model/patch"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/handler"
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
	handler.DoWithRepo(c,
		func(req *api.GeneratePatchReq) string { return req.RepoKey },
		func(repo *po.Repo, req *api.GeneratePatchReq) (map[string]string, error) {
			gitSvc := git.NewGitService()
			var patchContent string
			var err error
			if len(req.Commits) > 0 {
				patchContent, err = gitSvc.GeneratePatchForCommits(repo.Path, req.Commits)
			} else {
				if req.Base == "" || req.Target == "" {
					return nil, handler.ErrBadRequest("base and target are required when commits is empty")
				}
				patchContent, err = gitSvc.GeneratePatch(repo.Path, req.Base, req.Target)
			}
			if err != nil {
				return nil, handler.ErrInternal(err.Error())
			}
			return map[string]string{"content": patchContent}, nil
		})
}

// SavePatch 保存 patch 到仓库
// @router /api/v1/patch/save [POST]
func SavePatch(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepo(c,
		func(req *api.SavePatchReq) string { return req.RepoKey },
		func(repo *po.Repo, req *api.SavePatchReq) (map[string]string, error) {
			if req.PatchName == "" {
				return nil, handler.ErrBadRequest("patch_name is required")
			}
			if req.PatchContent == "" {
				return nil, handler.ErrBadRequest("patch_content is required")
			}
			savedPath, err := git.NewGitService().SavePatch(repo.Path, req.PatchContent, req.PatchName, req.CustomPath, req.CommitMessage)
			if err != nil {
				return nil, handler.ErrInternal(err.Error())
			}
			return map[string]string{
				"path": savedPath,
				"name": filepath.Base(savedPath),
			}, nil
		})
}

// ListPatches 列出仓库中的所有 patch
// @router /api/v1/patch/list [GET]
func ListPatches(ctx context.Context, c *app.RequestContext) {
	handler.DoWithQueryRepo(c, func(repo *po.Repo) ([]*patchModel.PatchInfo, error) {
		patches, err := git.NewGitService().ListPatches(repo.Path)
		if err != nil {
			return nil, handler.ErrInternal(err.Error())
		}
		var dtos []*patchModel.PatchInfo
		for _, p := range patches {
			dtos = append(dtos, toProtoPatchInfo(
				p.Name, p.Path, p.ModTime, p.Size,
				int32(p.Sequence), p.IsApplied, p.CanApply, p.HasConflict,
			))
		}
		return dtos, nil
	})
}

// GetPatchContent 获取 patch 内容
// @router /api/v1/patch/content [GET]
func GetPatchContent(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *struct {
		Path string `json:"path" query:"path" form:"path"`
	}) (map[string]string, error) {
		if req.Path == "" {
			return nil, handler.ErrBadRequest("path is required")
		}
		content, err := git.NewGitService().GetPatchContent(req.Path)
		if err != nil {
			return nil, handler.ErrInternal(err.Error())
		}
		return map[string]string{"content": content}, nil
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
	handler.DoWithRepo(c,
		func(req *api.ApplyPatchReq) string { return req.RepoKey },
		func(repo *po.Repo, req *api.ApplyPatchReq) (map[string]string, error) {
			gitSvc := git.NewGitService()
			if req.PatchContent != "" {
				if err := gitSvc.ApplyPatchFromContent(repo.Path, req.PatchContent, req.SignOff, req.CommitMessage); err != nil {
					return nil, handler.ErrInternal(err.Error())
				}
			} else if req.PatchPath != "" {
				if err := gitSvc.ApplyPatch(repo.Path, req.PatchPath, req.SignOff, req.CommitMessage); err != nil {
					return nil, handler.ErrInternal(err.Error())
				}
			} else {
				return nil, handler.ErrBadRequest("patch_path or patch_content is required")
			}
			return map[string]string{"message": "patch applied successfully"}, nil
		})
}

// CheckPatch 检查 patch 是否可以应用
// @router /api/v1/patch/check [POST]
func CheckPatch(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepo(c,
		func(req *struct {
			RepoKey   string `json:"repo_key" form:"repo_key"`
			PatchPath string `json:"patch_path" form:"patch_path"`
		}) string {
			return req.RepoKey
		},
		func(repo *po.Repo, req *struct {
			RepoKey   string `json:"repo_key" form:"repo_key"`
			PatchPath string `json:"patch_path" form:"patch_path"`
		}) (*patchModel.PatchStats, error) {
			stats, err := git.NewGitService().GetPatchStats(repo.Path, req.PatchPath)
			if err != nil {
				return nil, handler.ErrInternal(err.Error())
			}
			statStr, _ := stats["stat"].(string)
			canApply, _ := stats["can_apply"].(bool)
			var errStr string
			if e, ok := stats["error"].(string); ok {
				errStr = e
			}
			return &patchModel.PatchStats{
				Stat:     &statStr,
				CanApply: &canApply,
				Error:    &errStr,
			}, nil
		})
}

// DeletePatch 删除 patch
// @router /api/v1/patch/delete [POST]
func DeletePatch(ctx context.Context, c *app.RequestContext) {
	handler.Do(c, func(req *api.DeletePatchReq) error {
		return git.NewGitService().DeletePatch(req.PatchPath)
	})
}
