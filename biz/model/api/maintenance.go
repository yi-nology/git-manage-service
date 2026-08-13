package api

import "time"

type MaintenanceSnapshotDTO struct {
	GitDirSize      string `json:"git_dir_size"`
	GitDirSizeBytes int64  `json:"git_dir_size_bytes"`
	LooseObjects    int64  `json:"loose_objects"`
	PackFiles       int    `json:"pack_files"`
	InPackObjects   int64  `json:"in_pack_objects"`
	CommitCount     int64  `json:"commit_count"`
	BranchCount     int    `json:"branch_count"`
	TagCount        int    `json:"tag_count"`
}

type GitDirBreakdown struct {
	PackDirSize       string `json:"pack_dir_size"`
	PackDirSizeBytes  int64  `json:"pack_dir_size_bytes"`
	LooseObjSize      string `json:"loose_obj_size"`
	LooseObjSizeBytes int64  `json:"loose_obj_size_bytes"`
	ReflogSize        string `json:"reflog_size"`
	ReflogSizeBytes   int64  `json:"reflog_size_bytes"`
	StashCount        int    `json:"stash_count"`
	OtherSize         string `json:"other_size"`
	OtherSizeBytes    int64  `json:"other_size_bytes"`
}

type StashEntry struct {
	Index     int    `json:"index"`
	Message   string `json:"message"`
	Size      string `json:"size"`
	SizeBytes int64  `json:"size_bytes"`
}

type RepoHealthReport struct {
	GitDirSize      string           `json:"git_dir_size"`
	GitDirSizeBytes int64            `json:"git_dir_size_bytes"`
	LooseObjects    int64            `json:"loose_objects"`
	PackFiles       int              `json:"pack_files"`
	InPackObjects   int64            `json:"in_pack_objects"`
	CommitCount     int64            `json:"commit_count"`
	BranchCount     int              `json:"branch_count"`
	TagCount        int              `json:"tag_count"`
	GitDirBreakdown *GitDirBreakdown `json:"git_dir_breakdown"`
	StashEntries    []StashEntry     `json:"stash_entries"`
	LargeFiles      []LargeFileEntry `json:"large_files"`
	Threshold       int64            `json:"threshold"`
	ThresholdHuman  string           `json:"threshold_human"`
	Excludes        []string         `json:"excludes"`
}

type LargeFileEntry struct {
	Path        string `json:"path"`
	Size        string `json:"size"`
	SizeBytes   int64  `json:"size_bytes"`
	Exists      bool   `json:"exists"`
	CommitCount int    `json:"commit_count"`
	Source      string `json:"source"`
}

type MaintenanceRecordDTO struct {
	ID             uint                    `json:"id"`
	Type           string                  `json:"type"`
	Status         string                  `json:"status"`
	TriggerBy      string                  `json:"trigger_by"`
	ParamsJSON     string                  `json:"params_json"`
	SnapshotBefore *MaintenanceSnapshotDTO `json:"snapshot_before"`
	SnapshotAfter  *MaintenanceSnapshotDTO `json:"snapshot_after"`
	ErrorMessage   string                  `json:"error_message"`
	StartedAt      *time.Time              `json:"started_at"`
	FinishedAt     *time.Time              `json:"finished_at"`
	CreatedAt      time.Time               `json:"created_at"`
	Duration       string                  `json:"duration"`
	SavedBytes     int64                   `json:"saved_bytes"`
	SavedPercent   float64                 `json:"saved_percent"`
}

type MaintenanceRecordListResponse struct {
	Records  []MaintenanceRecordDTO `json:"records"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

type MaintenanceTaskResponse struct {
	TaskID string `json:"task_id"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type FileAIRecommendation struct {
	Path           string `json:"path"`
	Size           string `json:"size"`
	SizeBytes      int64  `json:"size_bytes"`
	Recommendation string `json:"recommendation"`
	Category       string `json:"category"`
	Reason         string `json:"reason"`
	Confidence     string `json:"confidence"`
}

type MaintenanceAIAnalysisResponse struct {
	Summary         string                 `json:"summary"`
	TotalSavings    string                 `json:"total_savings"`
	TotalSaveBytes  int64                  `json:"total_save_bytes"`
	Recommendations []FileAIRecommendation `json:"recommendations"`
}

type PrefixFileEntry struct {
	Path        string `json:"path"`
	Size        string `json:"size"`
	SizeBytes   int64  `json:"size_bytes"`
	Exists      bool   `json:"exists"`
	CommitCount int    `json:"commit_count"`
}

type PrefixSlimPreview struct {
	Files      []PrefixFileEntry `json:"files"`
	TotalCount int               `json:"total_count"`
	TotalSize  string            `json:"total_size"`
	TotalBytes int64             `json:"total_bytes"`
}

type ForcePushResult struct {
	RemoteName string `json:"remote_name"`
	Platform   string `json:"platform"`
	Branches   int    `json:"branches"`
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
}
