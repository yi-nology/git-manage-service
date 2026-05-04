package db

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type MaintenanceDAO struct{}

func NewMaintenanceDAO() *MaintenanceDAO {
	return &MaintenanceDAO{}
}

func (d *MaintenanceDAO) Create(record *po.MaintenanceRecord) error {
	return DB.Create(record).Error
}

func (d *MaintenanceDAO) Update(record *po.MaintenanceRecord) error {
	return DB.Save(record).Error
}

func (d *MaintenanceDAO) FindByID(id uint) (*po.MaintenanceRecord, error) {
	var record po.MaintenanceRecord
	err := DB.First(&record, id).Error
	return &record, err
}

func (d *MaintenanceDAO) FindByTaskID(taskID string) (*po.MaintenanceRecord, error) {
	var record po.MaintenanceRecord
	err := DB.Where("task_id = ?", taskID).First(&record).Error
	return &record, err
}

func (d *MaintenanceDAO) ListByRepoID(repoID uint, page, pageSize int) ([]po.MaintenanceRecord, int64, error) {
	var records []po.MaintenanceRecord
	var total int64
	db := DB.Where("repo_id = ?", repoID)
	db.Model(&po.MaintenanceRecord{}).Count(&total)
	err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error
	return records, total, err
}

func (d *MaintenanceDAO) UpdateProgress(taskID string, logs string) error {
	return DB.Model(&po.MaintenanceRecord{}).Where("task_id = ?", taskID).Update("progress_logs", logs).Error
}

func (d *MaintenanceDAO) UpdateStatus(taskID string, status string, errMsg string, snapshotAfter string) error {
	updates := map[string]interface{}{
		"status":         status,
		"error_message":  errMsg,
		"snapshot_after": snapshotAfter,
	}
	if status == "success" || status == "failed" {
		now := time.Now()
		updates["finished_at"] = &now
	}
	return DB.Model(&po.MaintenanceRecord{}).Where("task_id = ?", taskID).Updates(updates).Error
}
