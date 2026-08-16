package db

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
	"gorm.io/gorm"
)

type AuditLogDAO struct{ BaseDAO[po.AuditLog] }

func NewAuditLogDAO() *AuditLogDAO { return &AuditLogDAO{} }

// FindLatest 获取最近 N 条日志
func (d *AuditLogDAO) FindLatest(limit int) ([]po.AuditLog, error) {
	var logs []po.AuditLog
	return logs, DB.Order("created_at desc").Limit(limit).Find(&logs).Error
}

// FindPage 分页查询
func (d *AuditLogDAO) FindPage(page, pageSize int) ([]po.AuditLog, error) {
	return d.FindPageWithFilters(page, pageSize, "", "", "", "")
}

// FindPageWithFilters 带筛选的分页查询
func (d *AuditLogDAO) FindPageWithFilters(page, pageSize int, action, target, startDate, endDate string) ([]po.AuditLog, error) {
	var logs []po.AuditLog
	offset := (page - 1) * pageSize
	query := d.applyFilters(DB, action, target, startDate, endDate)
	return logs, query.Select("id", "action", "target", "operator", "ip_address", "user_agent", "created_at").
		Order("created_at desc").Offset(offset).Limit(pageSize).Find(&logs).Error
}

// CountWithFilters 带筛选的计数
func (d *AuditLogDAO) CountWithFilters(action, target, startDate, endDate string) (int64, error) {
	var count int64
	return count, d.applyFilters(DB.Model(new(po.AuditLog)), action, target, startDate, endDate).Count(&count).Error
}

func (d *AuditLogDAO) applyFilters(query *gorm.DB, action, target, startDate, endDate string) *gorm.DB {
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if target != "" {
		query = query.Where("target LIKE ?", "%"+target+"%")
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate+" 23:59:59")
	}
	return query
}

// FindByDateRange 查询日期范围内的日志
func (d *AuditLogDAO) FindByDateRange(startDate, endDate time.Time) ([]po.AuditLog, error) {
	var logs []po.AuditLog
	return logs, DB.Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Order("created_at asc").Find(&logs).Error
}

// DeleteByDateRange 删除日期范围内的日志
func (d *AuditLogDAO) DeleteByDateRange(startDate, endDate time.Time) error {
	return DB.Where("created_at >= ? AND created_at <= ?", startDate, endDate).Delete(new(po.AuditLog)).Error
}
