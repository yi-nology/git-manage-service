package db

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type MaintenanceDAO struct{ BaseDAO[po.MaintenanceRecord] }

func NewMaintenanceDAO() *MaintenanceDAO { return &MaintenanceDAO{} }

// Update 更新记录（与 Save 相同，保持向后兼容）
func (d *MaintenanceDAO) Update(record *po.MaintenanceRecord) error {
	return DB.Save(record).Error
}

// FindByTaskID 根据 task_id 查询
func (d *MaintenanceDAO) FindByTaskID(taskID string) (*po.MaintenanceRecord, error) {
	var record po.MaintenanceRecord
	return &record, DB.Where("task_id = ?", taskID).First(&record).Error
}

// ListByRepoID 分页查询指定仓库的维护记录
func (d *MaintenanceDAO) ListByRepoID(repoID uint, page, pageSize int) ([]po.MaintenanceRecord, int64, error) {
	var records []po.MaintenanceRecord
	var total int64
	db := DB.Where("repo_id = ?", repoID)
	db.Model(new(po.MaintenanceRecord)).Count(&total)
	err := db.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&records).Error
	return records, total, err
}

// UpdateProgress 更新进度日志
func (d *MaintenanceDAO) UpdateProgress(taskID string, logs string) error {
	return DB.Model(new(po.MaintenanceRecord)).Where("task_id = ?", taskID).Update("progress_logs", logs).Error
}

// UpdateStatus 更新任务状态
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
	return DB.Model(new(po.MaintenanceRecord)).Where("task_id = ?", taskID).Updates(updates).Error
}
