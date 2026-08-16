package db

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type MirrorDAO struct{ BaseDAO[po.Mirror] }

func NewMirrorDAO() *MirrorDAO { return &MirrorDAO{} }

// FindByID 覆盖基类：带 Preload
func (d *MirrorDAO) FindByID(id uint) (*po.Mirror, error) {
	var mirror po.Mirror
	return &mirror, DB.Preload("Repo").Preload("Credential").First(&mirror, id).Error
}

// FindAll 覆盖基类：带 Preload
func (d *MirrorDAO) FindAll() ([]po.Mirror, error) {
	var mirrors []po.Mirror
	return mirrors, DB.Preload("Repo").Preload("Credential").Find(&mirrors).Error
}

// FindByRepoID 查询仓库的镜像
func (d *MirrorDAO) FindByRepoID(repoID uint) ([]po.Mirror, error) {
	var mirrors []po.Mirror
	return mirrors, DB.Preload("Credential").Where("repo_id = ?", repoID).Find(&mirrors).Error
}

// FindEnabled 查询启用的镜像
func (d *MirrorDAO) FindEnabled() ([]po.Mirror, error) {
	var mirrors []po.Mirror
	return mirrors, DB.Preload("Repo").Preload("Credential").Where("enabled = ?", true).Find(&mirrors).Error
}

// FindDueForSync 查询到期需同步的镜像
func (d *MirrorDAO) FindDueForSync() ([]po.Mirror, error) {
	var mirrors []po.Mirror
	return mirrors, DB.Preload("Repo").Preload("Credential").
		Where("enabled = ? AND status IN ? AND next_sync_at <= ? AND deleted_at IS NULL",
			true, []string{po.MirrorStatusActive, po.MirrorStatusFailed}, time.Now()).
		Find(&mirrors).Error
}

// FindByWebhookToken 根据 webhook token 查询
func (d *MirrorDAO) FindByWebhookToken(token string) (*po.Mirror, error) {
	var mirror po.Mirror
	return &mirror, DB.Preload("Repo").Preload("Credential").Where("webhook_token = ?", token).First(&mirror).Error
}

// UpdateStatus 更新状态
func (d *MirrorDAO) UpdateStatus(id uint, status string) error {
	return DB.Model(new(po.Mirror)).Where("id = ?", id).Update("status", status).Error
}

// UpdateNextSyncAt 更新下次同步时间
func (d *MirrorDAO) UpdateNextSyncAt(id uint, nextSyncAt time.Time) error {
	return DB.Model(new(po.Mirror)).Where("id = ?", id).Update("next_sync_at", nextSyncAt).Error
}

// IncrementRetryCount 重试计数 +1
func (d *MirrorDAO) IncrementRetryCount(id uint) error {
	return DB.Model(new(po.Mirror)).Where("id = ?", id).
		UpdateColumn("retry_count", DB.Raw("retry_count + 1")).Error
}

// ResetRetryCount 重置重试计数
func (d *MirrorDAO) ResetRetryCount(id uint) error {
	return DB.Model(new(po.Mirror)).Where("id = ?", id).Update("retry_count", 0).Error
}

// CleanupOldLogs 清理过期同步日志
func (d *MirrorDAO) CleanupOldLogs(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	return DB.Where("created_at < ?", cutoff).Delete(new(po.MirrorSyncLog)).Error
}
