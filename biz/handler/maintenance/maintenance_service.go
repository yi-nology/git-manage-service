package maintenance

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	maintenance "github.com/yi-nology/git-manage-service/biz/model/maintenance"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/auth"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/response"
	"github.com/yi-nology/git-platform-sdk/gitbackend"
)

func Health(ctx context.Context, c *app.RequestContext) {
	var req maintenance.HealthRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	var threshold int64
	if t := c.Query("threshold"); t != "" {
		if v, err := strconv.ParseInt(t, 10, 64); err == nil && v > 0 {
			threshold = v
		}
	}

	var excludes []string
	if e := c.Query("exclude"); e != "" {
		for _, p := range strings.Split(e, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				excludes = append(excludes, p)
			}
		}
	}

	svc := git.NewMaintenanceService()
	report, err := svc.AnalyzeHealth(repo.Path, threshold, excludes)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	if report.LargeFiles == nil {
		report.LargeFiles = []api.LargeFileEntry{}
	}
	if report.StashEntries == nil {
		report.StashEntries = []api.StashEntry{}
	}

	response.Success(c, report)
}

// markMaintenanceFailed records a failed maintenance task: updates the in-memory
// task status and persists the failure to the maintenance record. paramsJSON is
// only re-written when non-empty (Slim/SlimByPrefix set it; GC does not), which
// preserves each handler's prior behavior exactly.
func markMaintenanceFailed(dao *db.MaintenanceDAO, taskID, paramsJSON, errMsg string) {
	git.GlobalTaskManager.UpdateStatus(taskID, "failed", errMsg)
	now := time.Now()
	rec, _ := dao.FindByTaskID(taskID)
	if rec == nil {
		return
	}
	rec.Status = "failed"
	rec.ErrorMessage = errMsg
	rec.TaskID = taskID
	if paramsJSON != "" {
		rec.ParamsJSON = paramsJSON
	}
	rec.FinishedAt = &now
	dao.Update(rec)
}

func Slim(ctx context.Context, c *app.RequestContext) {
	var req maintenance.SlimRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	paths := req.GetPaths()
	if len(paths) == 0 {
		response.BadRequest(c, "paths is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	taskID := uuid.New().String()
	git.GlobalTaskManager.AddTask(taskID)

	paramsJSON, _ := json.Marshal(map[string]interface{}{
		"paths":        paths,
		"addGitignore": req.GetAddGitignore(),
	})

	record, err := git.CreateMaintenanceRecord(repo.ID, "slim", repo.Path)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	dao := db.NewMaintenanceDAO()
	record.TaskID = taskID
	record.ParamsJSON = string(paramsJSON)
	dao.Update(record)
	dao.UpdateStatus(taskID, "running", "", "")

	go func() {
		svc := git.NewMaintenanceService()
		if err := svc.SlimHistory(repo.Path, paths, req.GetAddGitignore(), taskID); err != nil {
			markMaintenanceFailed(dao, taskID, string(paramsJSON), err.Error())
			return
		}
		git.GlobalTaskManager.UpdateStatus(taskID, "success", "")
	}()

	response.Success(c, &maintenance.MaintenanceTaskResponse{TaskId: &taskID})
}

func GC(ctx context.Context, c *app.RequestContext) {
	var req maintenance.HealthRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	taskID := uuid.New().String()
	git.GlobalTaskManager.AddTask(taskID)

	record, err := git.CreateMaintenanceRecord(repo.ID, "gc", repo.Path)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	db.NewMaintenanceDAO().UpdateStatus(taskID, "running", "", "")

	go func() {
		svc := git.NewMaintenanceService()
		if err := svc.GarbageCollect(repo.Path, taskID); err != nil {
			markMaintenanceFailed(db.NewMaintenanceDAO(), taskID, "", err.Error())
			return
		}
		git.GlobalTaskManager.UpdateStatus(taskID, "success", "")
	}()

	dao := db.NewMaintenanceDAO()
	record.TaskID = taskID
	dao.Update(record)

	response.Success(c, &maintenance.MaintenanceTaskResponse{TaskId: &taskID})
}

func Gitignore(ctx context.Context, c *app.RequestContext) {
	var req maintenance.GitignoreRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	paths := req.GetPaths()
	if len(paths) == 0 {
		response.BadRequest(c, "paths is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	svc := git.NewMaintenanceService()
	if err := svc.AddToGitignore(repo.Path, paths); err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, api.MessageResponse{Message: "已添加到 .gitignore"})
}

func ListRecords(ctx context.Context, c *app.RequestContext) {
	var req maintenance.ListRecordsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	records, total, err := db.NewMaintenanceDAO().ListByRepoID(repo.ID, page, pageSize)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	dtos := make([]api.MaintenanceRecordDTO, 0, len(records))
	for _, r := range records {
		dto := convertRecordToDTO(&r)
		dtos = append(dtos, dto)
	}

	response.Success(c, api.MaintenanceRecordListResponse{
		Records:  dtos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func GetRecord(ctx context.Context, c *app.RequestContext) {
	var req maintenance.GetRecordRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	_, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	record, err := db.NewMaintenanceDAO().FindByID(uint(req.GetId()))
	if err != nil {
		response.NotFound(c, "record not found")
		return
	}

	response.Success(c, convertRecordToDTO(record))
}

func convertRecordToDTO(r *po.MaintenanceRecord) api.MaintenanceRecordDTO {
	dto := api.MaintenanceRecordDTO{
		ID:           r.ID,
		Type:         r.Type,
		Status:       r.Status,
		TriggerBy:    r.TriggerBy,
		ParamsJSON:   r.ParamsJSON,
		ErrorMessage: r.ErrorMessage,
		StartedAt:    r.StartedAt,
		FinishedAt:   r.FinishedAt,
		CreatedAt:    r.CreatedAt,
	}

	if r.SnapshotBefore != "" {
		var snap api.MaintenanceSnapshotDTO
		if json.Unmarshal([]byte(r.SnapshotBefore), &snap) == nil {
			dto.SnapshotBefore = &snap
		}
	}
	if r.SnapshotAfter != "" {
		var snap api.MaintenanceSnapshotDTO
		if json.Unmarshal([]byte(r.SnapshotAfter), &snap) == nil {
			dto.SnapshotAfter = &snap
		}
	}

	if dto.SnapshotBefore != nil && dto.SnapshotAfter != nil {
		saved := dto.SnapshotBefore.GitDirSizeBytes - dto.SnapshotAfter.GitDirSizeBytes
		dto.SavedBytes = saved
		if dto.SnapshotBefore.GitDirSizeBytes > 0 {
			dto.SavedPercent = float64(saved) / float64(dto.SnapshotBefore.GitDirSizeBytes) * 100
		}
	}

	if r.StartedAt != nil && r.FinishedAt != nil {
		d := r.FinishedAt.Sub(*r.StartedAt)
		dto.Duration = formatDuration(d)
	} else if r.StartedAt != nil {
		d := time.Since(*r.StartedAt)
		dto.Duration = formatDuration(d)
	}

	return dto
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fmin", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// AIAnalyze .
// @router /api/v1/repo/:repo_key/maintenance/ai-analyze [POST]
func AIAnalyze(ctx context.Context, c *app.RequestContext) {
	var req maintenance.AIAnalyzeRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	var threshold int64
	if req.GetThreshold() > 0 {
		threshold = req.GetThreshold()
	}

	svc := git.NewMaintenanceService()

	var healthReport *api.RepoHealthReport

	if len(req.GetFilePaths()) > 0 {
		healthReport, err = svc.AnalyzeHealthForPaths(repo.Path, threshold, req.GetFilePaths())
	} else {
		healthReport, err = svc.AnalyzeHealth(repo.Path, threshold, nil)
	}
	if err != nil {
		response.InternalError(c, err)
		return
	}

	aiSvc := git.NewMaintenanceAIService()
	result, err := aiSvc.AnalyzeSlimFiles(ctx, healthReport)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, result)
}

func PreviewPrefixSlim(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Param("repo_key")
	if repoKey == "" {
		response.BadRequest(c, "repo_key is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	var body struct {
		Prefixes []string `json:"prefixes"`
	}
	if err := c.BindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if len(body.Prefixes) == 0 {
		response.BadRequest(c, "prefixes is required")
		return
	}

	svc := git.NewMaintenanceService()
	files, err := svc.FindByPrefix(repo.Path, body.Prefixes)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	var totalBytes int64
	for _, f := range files {
		totalBytes += f.SizeBytes
	}

	response.Success(c, api.PrefixSlimPreview{
		Files:      files,
		TotalCount: len(files),
		TotalSize:  formatSize2(totalBytes),
		TotalBytes: totalBytes,
	})
}

func SlimByPrefix(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Param("repo_key")
	if repoKey == "" {
		response.BadRequest(c, "repo_key is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	var body struct {
		Prefixes     []string `json:"prefixes"`
		AddGitignore *bool    `json:"addGitignore"`
		ForcePush    *bool    `json:"forcePush"`
	}
	if err := c.BindJSON(&body); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if len(body.Prefixes) == 0 {
		response.BadRequest(c, "prefixes is required")
		return
	}

	addGitignore := true
	if body.AddGitignore != nil {
		addGitignore = *body.AddGitignore
	}
	forcePush := false
	if body.ForcePush != nil {
		forcePush = *body.ForcePush
	}

	taskID := uuid.New().String()
	git.GlobalTaskManager.AddTask(taskID)

	paramsJSON, _ := json.Marshal(map[string]interface{}{
		"prefixes":     body.Prefixes,
		"addGitignore": addGitignore,
		"forcePush":    forcePush,
	})

	record, err := git.CreateMaintenanceRecord(repo.ID, "slim_prefix", repo.Path)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	dao := db.NewMaintenanceDAO()
	record.TaskID = taskID
	record.ParamsJSON = string(paramsJSON)
	dao.Update(record)
	dao.UpdateStatus(taskID, "running", "", "")

	repoID := repo.ID
	repoPath := repo.Path

	go func() {
		svc := git.NewMaintenanceService()
		if err := svc.SlimHistoryByPrefix(repoPath, body.Prefixes, addGitignore, taskID); err != nil {
			markMaintenanceFailed(dao, taskID, string(paramsJSON), err.Error())
			return
		}

		if forcePush {
			git.GlobalTaskManager.AppendLog(taskID, "开始强制推送到远端...")
			results := doForcePushAllRemotes(repoID, repoPath, taskID)
			for _, r := range results {
				if r.Success {
					git.GlobalTaskManager.AppendLog(taskID, fmt.Sprintf("推送 %s (%s) 成功，%d 个分支", r.RemoteName, r.Platform, r.Branches))
				} else {
					git.GlobalTaskManager.AppendLog(taskID, fmt.Sprintf("推送 %s (%s) 失败: %s", r.RemoteName, r.Platform, r.Error))
				}
			}
		}

		git.GlobalTaskManager.UpdateStatus(taskID, "success", "")
	}()

	response.Success(c, &maintenance.MaintenanceTaskResponse{TaskId: &taskID})
}

func ForcePushRemotes(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Param("repo_key")
	if repoKey == "" {
		response.BadRequest(c, "repo_key is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	taskID := uuid.New().String()
	git.GlobalTaskManager.AddTask(taskID)

	repoID := repo.ID
	repoPath := repo.Path

	paramsJSON, _ := json.Marshal(map[string]interface{}{
		"operation": "force_push_all",
	})

	record, err := git.CreateMaintenanceRecord(repoID, "force_push", repoPath)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	dao := db.NewMaintenanceDAO()
	record.TaskID = taskID
	record.ParamsJSON = string(paramsJSON)
	dao.Update(record)
	dao.UpdateStatus(taskID, "running", "", "")

	go func() {
		results := doForcePushAllRemotes(repoID, repoPath, taskID)
		allSuccess := true
		for _, r := range results {
			if !r.Success {
				allSuccess = false
			}
		}
		if allSuccess {
			afterSnap := git.NewMaintenanceService().TakeSnapshot(repoPath)
			afterJSON, _ := json.Marshal(afterSnap)
			dao.UpdateStatus(taskID, "success", "", string(afterJSON))
			git.GlobalTaskManager.UpdateStatus(taskID, "success", "")
		} else {
			errMsg := "部分远端推送失败"
			dao.UpdateStatus(taskID, "failed", errMsg, "")
			git.GlobalTaskManager.UpdateStatus(taskID, "failed", errMsg)
		}
	}()

	response.Success(c, &maintenance.ForcePushResponse{Results: []string{taskID}})
}

func doForcePushAllRemotes(repoID uint, repoPath string, taskID string) []api.ForcePushResult {
	tm := git.GlobalTaskManager

	bindings, err := db.NewRepoProviderBindingDAO().FindByRepoIDWithProvider(repoID)
	if err != nil || len(bindings) == 0 {
		tm.AppendLog(taskID, "未找到远端绑定")
		return nil
	}

	gitSvc := git.NewGitService()
	branches, err := gitSvc.GetBranches(repoPath)
	if err != nil {
		tm.AppendLog(taskID, "获取本地分支失败: "+err.Error())
		return nil
	}

	var localBranches []string
	for _, b := range branches {
		if !strings.Contains(b, "/") {
			localBranches = append(localBranches, b)
		}
	}

	if len(localBranches) == 0 {
		tm.AppendLog(taskID, "没有本地分支可推送")
		return nil
	}

	tm.AppendLog(taskID, fmt.Sprintf("找到 %d 个本地分支: %s", len(localBranches), strings.Join(localBranches, ", ")))

	var results []api.ForcePushResult

	for _, binding := range bindings {
		remoteName := binding.RemoteName
		if remoteName == "" {
			remoteName = "origin"
		}

		platform := binding.ProviderConfig.Platform
		if platform == "" {
			platform = "unknown"
		}

		remoteURL, _ := gitSvc.GetRemoteURL(repoPath, remoteName)
		if remoteURL == "" {
			tm.AppendLog(taskID, fmt.Sprintf("跳过 %s: 无法获取远端 URL", remoteName))
			results = append(results, api.ForcePushResult{
				RemoteName: remoteName,
				Platform:   platform,
				Success:    false,
				Error:      "无法获取远端 URL",
			})
			continue
		}

		tm.AppendLog(taskID, fmt.Sprintf("推送 %s (%s) -> %s ...", remoteName, platform, remoteURL))

		skipTLS := binding.ProviderConfig.SkipTLS

		successCount := 0
		var pushErr string

		for _, branch := range localBranches {
			hash, err := gitSvc.ResolveRevision(repoPath, branch)
			if err != nil {
				pushErr = fmt.Sprintf("分支 %s 解析失败: %v", branch, err)
				tm.AppendLog(taskID, pushErr)
				continue
			}

			pushErrDetail := pushBranchToRemote(repoID, repoPath, remoteName, remoteURL, branch, hash, gitSvc, skipTLS)
			if pushErrDetail != "" {
				pushErr = pushErrDetail
				tm.AppendLog(taskID, fmt.Sprintf("  分支 %s 推送失败: %s", branch, pushErrDetail))
			} else {
				successCount++
				tm.AppendLog(taskID, fmt.Sprintf("  分支 %s 推送成功", branch))
			}
		}

		results = append(results, api.ForcePushResult{
			RemoteName: remoteName,
			Platform:   platform,
			Branches:   successCount,
			Success:    pushErr == "" || successCount > 0,
			Error:      pushErr,
		})
	}

	return results
}

func pushBranchToRemote(repoID uint, repoPath, remoteName, remoteURL, branch, hash string, gitSvc *git.GitService, skipTLS bool) string {
	repo, err := db.NewRepoDAO().FindByID(repoID)
	if err != nil {
		return fmt.Sprintf("获取仓库信息失败: %v", err)
	}

	authSvc := auth.NewAuthService()
	authMethod, isDBKey, err := authSvc.ResolveCredentialForRemote(
		repo.RemoteCredentials,
		repo.DefaultCredentialID,
		repo.RemoteAuths,
		remoteName,
		repo.AuthType, repo.AuthKey, repo.AuthSecret,
	)
	if err != nil {
	}

	hasAuth := authMethod.Type != gitbackend.AuthNone || isDBKey
	pushOpts := []string{"--force"}

	if remoteURL != "" && hasAuth {
		if isDBKey {
			credID := auth.GetCredentialIDForRemote(repo.RemoteCredentials, repo.DefaultCredentialID, remoteName)
			if credID > 0 {
				privateKey, passphrase, keyErr := authSvc.GetCredentialKeyContent(credID)
				if keyErr == nil && privateKey != "" {
					err := gitSvc.PushWithDBKey(repoPath, remoteURL, hash, branch, privateKey, passphrase, pushOpts, nil)
					if err != nil {
						return err.Error()
					}
					return ""
				}
			}
		}
		if authMethod.Type != gitbackend.AuthNone {
			err := gitSvc.PushWithSDKAuth(repoPath, remoteURL, hash, branch, authMethod, pushOpts, nil, skipTLS)
			if err != nil {
				return err.Error()
			}
			return ""
		}
	}

	err = gitSvc.Push(repoPath, remoteName, hash, branch, pushOpts, nil, skipTLS)
	if err != nil {
		return err.Error()
	}
	return ""
}

func formatSize2(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
