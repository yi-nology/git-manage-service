package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type MirrorSyncLogDAO struct{}

func NewMirrorSyncLogDAO() *MirrorSyncLogDAO {
	return &MirrorSyncLogDAO{}
}

func (d *MirrorSyncLogDAO) Create(log *po.MirrorSyncLog) error {
	return DB.Create(log).Error
}

func (d *MirrorSyncLogDAO) Save(log *po.MirrorSyncLog) error {
	return DB.Save(log).Error
}

func (d *MirrorSyncLogDAO) FindByID(id uint) (*po.MirrorSyncLog, error) {
	var log po.MirrorSyncLog
	if err := DB.Preload("Mirror").First(&log, id).Error; err != nil {
		return nil, err
	}
	return &log, nil
}

func (d *MirrorSyncLogDAO) FindByMirrorID(mirrorID uint, limit int) ([]po.MirrorSyncLog, error) {
	var logs []po.MirrorSyncLog
	q := DB.Where("mirror_id = ?", mirrorID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&logs).Error
	return logs, err
}

func (d *MirrorSyncLogDAO) FindLatest(limit int) ([]po.MirrorSyncLog, error) {
	var logs []po.MirrorSyncLog
	err := DB.Preload("Mirror").Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (d *MirrorSyncLogDAO) FindByMirrorIDs(mirrorIDs []uint, limit int) ([]po.MirrorSyncLog, error) {
	var logs []po.MirrorSyncLog
	q := DB.Where("mirror_id IN ?", mirrorIDs).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Preload("Mirror").Find(&logs).Error
	return logs, err
}

func (d *MirrorSyncLogDAO) Delete(id uint) error {
	return DB.Delete(&po.MirrorSyncLog{}, id).Error
}

func (d *MirrorSyncLogDAO) CountByMirrorID(mirrorID uint) (int64, error) {
	var count int64
	err := DB.Model(&po.MirrorSyncLog{}).Where("mirror_id = ?", mirrorID).Count(&count).Error
	return count, err
}
