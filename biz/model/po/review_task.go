package po

import (
	"time"

	"gorm.io/gorm"
)

type ReviewTask struct {
	gorm.Model
	RepoID           uint   `gorm:"index"`
	ProviderConfigID uint   `gorm:"index"`
	Platform         string `gorm:"size:20"`
	PlatformOwner    string `gorm:"size:255"`
	PlatformRepo     string `gorm:"size:255"`
	EventType        string `gorm:"size:64"`
	MRIID            string `gorm:"size:64;index"`
	SourceBranch     string `gorm:"size:255"`
	TargetBranch     string `gorm:"size:255"`
	CommitSHA        string `gorm:"size:128;index"`
	TriggerType      string `gorm:"size:32"`
	TriggerUser      string `gorm:"size:255"`
	Status           string `gorm:"size:32;index"`
	RiskLevel        string `gorm:"size:32"`
	ReviewMode       string `gorm:"size:32;default:'llm'"`
	Summary          string `gorm:"type:text"`
	ErrorMessage     string `gorm:"type:text"`
	ConfigSnapshot   string `gorm:"type:text"`
	RawDiff          string `gorm:"type:text"`
	ProcessLog       string `gorm:"type:text"`
	StartedAt        *time.Time
	FinishedAt       *time.Time
}

func (ReviewTask) TableName() string { return "review_tasks" }
