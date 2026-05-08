package db

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type MirrorDAO struct{}

func NewMirrorDAO() *MirrorDAO {
	return &MirrorDAO{}
}

func (d *MirrorDAO) Create(mirror *po.Mirror) error {
	return DB.Create(mirror).Error
}

func (d *MirrorDAO) FindByID(id uint) (*po.Mirror, error) {
	var mirror po.Mirror
	if err := DB.Preload("Repo").Preload("Credential").First(&mirror, id).Error; err != nil {
		return nil, err
	}
	return &mirror, nil
}

func (d *MirrorDAO) FindAll() ([]po.Mirror, error) {
	var mirrors []po.Mirror
	err := DB.Preload("Repo").Preload("Credential").Find(&mirrors).Error
	return mirrors, err
}

func (d *MirrorDAO) FindByRepoID(repoID uint) ([]po.Mirror, error) {
	var mirrors []po.Mirror
	err := DB.Preload("Credential").Where("repo_id = ?", repoID).Find(&mirrors).Error
	return mirrors, err
}

func (d *MirrorDAO) FindByMirrorType(mirrorType string) ([]po.Mirror, error) {
	var mirrors []po.Mirror
	err := DB.Preload("Repo").Where("mirror_type = ?", mirrorType).Find(&mirrors).Error
	return mirrors, err
}

func (d *MirrorDAO) FindEnabled() ([]po.Mirror, error) {
	var mirrors []po.Mirror
	err := DB.Preload("Repo").Preload("Credential").Where("enabled = ?", true).Find(&mirrors).Error
	return mirrors, err
}

func (d *MirrorDAO) FindDueForSync() ([]po.Mirror, error) {
	var mirrors []po.Mirror
	now := time.Now()
	err := DB.Preload("Repo").Preload("Credential").
		Where("enabled = ? AND status IN ? AND next_sync_at <= ? AND deleted_at IS NULL",
			true, []string{po.MirrorStatusActive, po.MirrorStatusFailed}, now).
		Find(&mirrors).Error
	return mirrors, err
}

func (d *MirrorDAO) FindByWebhookToken(token string) (*po.Mirror, error) {
	var mirror po.Mirror
	if err := DB.Preload("Repo").Preload("Credential").Where("webhook_token = ?", token).First(&mirror).Error; err != nil {
		return nil, err
	}
	return &mirror, nil
}

func (d *MirrorDAO) FindBySyncOnPush() ([]po.Mirror, error) {
	var mirrors []po.Mirror
	err := DB.Preload("Repo").Preload("Credential").
		Where("sync_on_push = ? AND enabled = ? AND deleted_at IS NULL", true, true).
		Find(&mirrors).Error
	return mirrors, err
}

func (d *MirrorDAO) Save(mirror *po.Mirror) error {
	return DB.Save(mirror).Error
}

func (d *MirrorDAO) Delete(id uint) error {
	return DB.Delete(&po.Mirror{}, id).Error
}

func (d *MirrorDAO) UpdateStatus(id uint, status string) error {
	return DB.Model(&po.Mirror{}).Where("id = ?", id).Update("status", status).Error
}

func (d *MirrorDAO) UpdateNextSyncAt(id uint, nextSyncAt time.Time) error {
	return DB.Model(&po.Mirror{}).Where("id = ?", id).Update("next_sync_at", nextSyncAt).Error
}

func (d *MirrorDAO) IncrementRetryCount(id uint) error {
	return DB.Model(&po.Mirror{}).Where("id = ?", id).
		UpdateColumn("retry_count", DB.Raw("retry_count + 1")).Error
}

func (d *MirrorDAO) ResetRetryCount(id uint) error {
	return DB.Model(&po.Mirror{}).Where("id = ?", id).Update("retry_count", 0).Error
}

func (d *MirrorDAO) CountByRepoID(repoID uint) (int64, error) {
	var count int64
	err := DB.Model(&po.Mirror{}).Where("repo_id = ?", repoID).Count(&count).Error
	return count, err
}

func (d *MirrorDAO) CleanupOldLogs(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	return DB.Where("created_at < ?", cutoff).Delete(&po.MirrorSyncLog{}).Error
}
