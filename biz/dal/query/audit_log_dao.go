package query

import (
	"github.com/yi-nology/git-manage-service/biz/dal"
	"github.com/yi-nology/git-manage-service/biz/model"
)

type AuditLogDAO struct{}

func NewAuditLogDAO() *AuditLogDAO {
	return &AuditLogDAO{}
}

func (d *AuditLogDAO) Create(log *model.AuditLog) error {
	return dal.DB.Create(log).Error
}

func (d *AuditLogDAO) FindLatest(limit int) ([]model.AuditLog, error) {
	var logs []model.AuditLog
	err := dal.DB.Order("created_at desc").Limit(limit).Find(&logs).Error
	return logs, err
}
