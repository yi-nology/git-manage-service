package po

import (
	"time"

	"gorm.io/gorm"
)

type MaintenanceRecord struct {
	gorm.Model
	RepoID        uint       `gorm:"index" json:"repoId"`
	Type          string     `gorm:"size:20;index" json:"type"`
	Status        string     `gorm:"size:20;index" json:"status"`
	TriggerBy     string     `gorm:"size:20" json:"triggerBy"`
	ParamsJSON    string     `json:"paramsJson"`
	SnapshotBefore string    `json:"snapshotBefore"`
	SnapshotAfter  string    `json:"snapshotAfter"`
	ErrorMessage  string     `json:"errorMessage"`
	ProgressLogs  string     `json:"progressLogs"`
	TaskID        string     `gorm:"size:64;index" json:"taskId"`
	StartedAt     *time.Time `json:"startedAt"`
	FinishedAt    *time.Time `json:"finishedAt"`
}

func (MaintenanceRecord) TableName() string {
	return "maintenance_records"
}
