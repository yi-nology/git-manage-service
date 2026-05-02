package maintenance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	maintenance "github.com/yi-nology/git-manage-service/biz/model/maintenance"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/response"
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

	svc := git.NewMaintenanceService()
	report, err := svc.AnalyzeHealth(repo.Path)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	largeFiles, err := svc.FindLargeFiles(repo.Path, 1*1024*1024)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	report.LargeFiles = largeFiles
	if report.LargeFiles == nil {
		report.LargeFiles = []api.LargeFileEntry{}
	}

	response.Success(c, report)
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
		"paths":         paths,
		"addGitignore":  req.GetAddGitignore(),
	})

	record, err := git.CreateMaintenanceRecord(repo.ID, "slim", repo.Path)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	db.NewMaintenanceDAO().UpdateProgress(taskID, "")
	db.NewMaintenanceDAO().UpdateStatus(taskID, "running", "", "")

	go func() {
		svc := git.NewMaintenanceService()
		if err := svc.SlimHistory(repo.Path, paths, req.GetAddGitignore(), taskID); err != nil {
			git.GlobalTaskManager.UpdateStatus(taskID, "failed", err.Error())
			now := time.Now()
			dao := db.NewMaintenanceDAO()
			rec, _ := dao.FindByTaskID(taskID)
			if rec != nil {
				rec.Status = "failed"
				rec.ErrorMessage = err.Error()
				rec.TaskID = taskID
				rec.ParamsJSON = string(paramsJSON)
				rec.FinishedAt = &now
				dao.Update(rec)
			}
			return
		}
		git.GlobalTaskManager.UpdateStatus(taskID, "success", "")
	}()

	dao := db.NewMaintenanceDAO()
	record.TaskID = taskID
	record.ParamsJSON = string(paramsJSON)
	dao.Update(record)

	response.Success(c, api.MaintenanceTaskResponse{TaskID: taskID})
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
			git.GlobalTaskManager.UpdateStatus(taskID, "failed", err.Error())
			now := time.Now()
			dao := db.NewMaintenanceDAO()
			rec, _ := dao.FindByTaskID(taskID)
			if rec != nil {
				rec.Status = "failed"
				rec.ErrorMessage = err.Error()
				rec.TaskID = taskID
				rec.FinishedAt = &now
				dao.Update(rec)
			}
			return
		}
		git.GlobalTaskManager.UpdateStatus(taskID, "success", "")
	}()

	dao := db.NewMaintenanceDAO()
	record.TaskID = taskID
	dao.Update(record)

	response.Success(c, api.MaintenanceTaskResponse{TaskID: taskID})
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
