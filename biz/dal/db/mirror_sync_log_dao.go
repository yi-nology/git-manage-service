package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type MirrorSyncLogDAO struct{ BaseDAO[po.MirrorSyncLog] }

func NewMirrorSyncLogDAO() *MirrorSyncLogDAO { return &MirrorSyncLogDAO{} }

// FindByID 覆盖基类：带 Preload
func (d *MirrorSyncLogDAO) FindByID(id uint) (*po.MirrorSyncLog, error) {
	var log po.MirrorSyncLog
	return &log, DB.Preload("Mirror").First(&log, id).Error
}

// FindByMirrorID 查询镜像的同步日志
func (d *MirrorSyncLogDAO) FindByMirrorID(mirrorID uint, limit int) ([]po.MirrorSyncLog, error) {
	var logs []po.MirrorSyncLog
	q := DB.Where("mirror_id = ?", mirrorID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	return logs, q.Find(&logs).Error
}

// FindLatest 查询最近的同步日志（带 Mirror Preload）
func (d *MirrorSyncLogDAO) FindLatest(limit int) ([]po.MirrorSyncLog, error) {
	var logs []po.MirrorSyncLog
	return logs, DB.Preload("Mirror").Order("created_at DESC").Limit(limit).Find(&logs).Error
}
