package model

import (
	"time"

	"gorm.io/gorm"
)

type Repo struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"uniqueIndex" json:"name"`
	Path      string         `json:"path"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type SyncTask struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	SourceRepoID uint           `json:"source_repo_id"`
	SourceRemote string         `json:"source_remote"`
	SourceBranch string         `json:"source_branch"`
	TargetRepoID uint           `json:"target_repo_id"`
	TargetRemote string         `json:"target_remote"`
	TargetBranch string         `json:"target_branch"`
	PushOptions  string         `json:"push_options"` // e.g. "--force --no-verify"
	Cron         string         `json:"cron"`         // e.g. "0 2 * * *"
	Enabled      bool           `json:"enabled"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	// Associations
	SourceRepo Repo `gorm:"foreignKey:SourceRepoID" json:"source_repo"`
	TargetRepo Repo `gorm:"foreignKey:TargetRepoID" json:"target_repo"`
}

type SyncRun struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	TaskID       uint      `json:"task_id"`
	Status       string    `json:"status"` // success, failed, conflict
	CommitRange  string    `json:"commit_range"`
	ErrorMessage string    `json:"error_message"`
	Details      string    `json:"details" gorm:"type:text"` // Execution logs
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`

	// Associations
	Task SyncTask `gorm:"foreignKey:TaskID" json:"-"`
}
