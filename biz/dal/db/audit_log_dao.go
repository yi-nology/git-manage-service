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
