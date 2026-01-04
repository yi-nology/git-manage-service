package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type AuditLogDAO struct{}

func NewAuditLogDAO() *AuditLogDAO {
	return &AuditLogDAO{}
}

func (d *AuditLogDAO) Create(log *po.AuditLog) error {
	return DB.Create(log).Error
}

func (d *AuditLogDAO) FindLatest(limit int) ([]po.AuditLog, error) {
	var logs []po.AuditLog
	err := DB.Order("created_at desc").Limit(limit).Find(&logs).Error
	return logs, err
}

func (d *AuditLogDAO) Count() (int64, error) {
	var count int64
	err := DB.Model(&po.AuditLog{}).Count(&count).Error
	return count, err
}

func (d *AuditLogDAO) FindPage(page, pageSize int) ([]po.AuditLog, error) {
	var logs []po.AuditLog
	offset := (page - 1) * pageSize
	// Exclude 'details' column for list view to improve performance
	err := DB.Select("id", "action", "target", "operator", "ip_address", "user_agent", "created_at").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&logs).Error
	return logs, err
}

func (d *AuditLogDAO) FindByID(id uint) (*po.AuditLog, error) {
	var log po.AuditLog
	err := DB.First(&log, id).Error
	return &log, err
}
