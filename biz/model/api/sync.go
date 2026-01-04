package api

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type SyncRunDTO struct {
	ID           uint        `json:"id"`
	TaskKey      string      `json:"task_key"`
	Status       string      `json:"status"`
	CommitRange  string      `json:"commit_range"`
	ErrorMessage string      `json:"error_message"`
	Details      string      `json:"details"`
	StartTime    time.Time   `json:"start_time"`
	EndTime      time.Time   `json:"end_time"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Task         SyncTaskDTO `json:"task"`
}

func NewSyncRunDTO(r po.SyncRun) SyncRunDTO {
	dto := SyncRunDTO{
		ID:           r.ID,
		TaskKey:      r.TaskKey,
		Status:       r.Status,
		CommitRange:  r.CommitRange,
		ErrorMessage: r.ErrorMessage,
		Details:      r.Details,
		StartTime:    r.StartTime,
		EndTime:      r.EndTime,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}
	if r.Task.ID != 0 {
		dto.Task = NewSyncTaskDTO(r.Task)
	}
	return dto
}
