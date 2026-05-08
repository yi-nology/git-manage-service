package po

import (
	"time"

	"gorm.io/gorm"
)

const (
	TriggerTypeManual    = "manual"
	TriggerTypeCron      = "cron"
	TriggerTypeWebhook   = "webhook"
	TriggerTypePushEvent = "push_event"

	SyncLogStatusPending = "pending"
	SyncLogStatusRunning = "running"
	SyncLogStatusSuccess = "success"
	SyncLogStatusFailed  = "failed"
)

type MirrorSyncLog struct {
	gorm.Model
	MirrorID       uint       `gorm:"not null;index" json:"mirrorId"`
	TriggerType    string     `gorm:"size:20;not null" json:"triggerType"`
	Status         string     `gorm:"size:20;not null" json:"status"`
	StartedAt      *time.Time `json:"startedAt"`
	FinishedAt     *time.Time `json:"finishedAt"`
	DurationMs     int64      `gorm:"default:0" json:"durationMs"`
	BranchesSynced int        `gorm:"default:0" json:"branchesSynced"`
	CommitsPushed  int        `gorm:"default:0" json:"commitsPushed"`
	ErrorMessage   string     `json:"errorMessage"`
	DetailLog      string     `gorm:"type:text" json:"detailLog"`

	Mirror Mirror `gorm:"foreignKey:MirrorID" json:"mirror"`
}

func (MirrorSyncLog) TableName() string {
	return "mirror_sync_logs"
}
