package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type ReviewAuditLogDAO struct{}

func NewReviewAuditLogDAO() *ReviewAuditLogDAO { return &ReviewAuditLogDAO{} }

func (d *ReviewAuditLogDAO) Create(log *po.ReviewAuditLog) error {
	return DB.Create(log).Error
}

func (d *ReviewAuditLogDAO) FindByTaskID(taskID uint, page, pageSize int) ([]po.ReviewAuditLog, int64, error) {
	var logs []po.ReviewAuditLog
	var total int64
	q := DB.Model(&po.ReviewAuditLog{})
	if taskID > 0 {
		q = q.Where("task_id = ?", taskID)
	}
	q.Count(&total)
	err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}
