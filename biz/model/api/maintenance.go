package api

import "time"

type MaintenanceSnapshotDTO struct {
	GitDirSize      string `json:"gitDirSize"`
	GitDirSizeBytes int64  `json:"gitDirSizeBytes"`
	LooseObjects    int64  `json:"looseObjects"`
	PackFiles       int    `json:"packFiles"`
	InPackObjects   int64  `json:"inPackObjects"`
	CommitCount     int64  `json:"commitCount"`
	BranchCount     int    `json:"branchCount"`
	TagCount        int    `json:"tagCount"`
}

type RepoHealthReport struct {
	GitDirSize      string              `json:"gitDirSize"`
	GitDirSizeBytes int64               `json:"gitDirSizeBytes"`
	LooseObjects    int64               `json:"looseObjects"`
	PackFiles       int                 `json:"packFiles"`
	InPackObjects   int64               `json:"inPackObjects"`
	CommitCount     int64               `json:"commitCount"`
	BranchCount     int                 `json:"branchCount"`
	TagCount        int                 `json:"tagCount"`
	LargeFiles      []LargeFileEntry    `json:"largeFiles"`
}

type LargeFileEntry struct {
	Path        string `json:"path"`
	Size        string `json:"size"`
	SizeBytes   int64  `json:"sizeBytes"`
	Exists      bool   `json:"exists"`
	CommitCount int    `json:"commitCount"`
}

type MaintenanceRecordDTO struct {
	ID             uint                  `json:"id"`
	Type           string                `json:"type"`
	Status         string                `json:"status"`
	TriggerBy      string                `json:"triggerBy"`
	ParamsJSON     string                `json:"paramsJson"`
	SnapshotBefore *MaintenanceSnapshotDTO `json:"snapshotBefore"`
	SnapshotAfter  *MaintenanceSnapshotDTO `json:"snapshotAfter"`
	ErrorMessage   string                `json:"errorMessage"`
	StartedAt      *time.Time            `json:"startedAt"`
	FinishedAt     *time.Time            `json:"finishedAt"`
	CreatedAt      time.Time             `json:"createdAt"`
	Duration       string                `json:"duration"`
	SavedBytes     int64                 `json:"savedBytes"`
	SavedPercent   float64               `json:"savedPercent"`
}

type MaintenanceRecordListResponse struct {
	Records []MaintenanceRecordDTO `json:"records"`
	Total   int64                  `json:"total"`
	Page    int                    `json:"page"`
	PageSize int                   `json:"pageSize"`
}

type MaintenanceTaskResponse struct {
	TaskID string `json:"taskId"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type HealthResponse struct {
	Report  *RepoHealthReport      `json:"report"`
	RecordID uint                  `json:"recordId"`
}
