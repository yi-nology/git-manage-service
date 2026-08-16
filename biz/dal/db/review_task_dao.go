package db

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type ReviewTaskDAO struct{ BaseDAO[po.ReviewTask] }

func NewReviewTaskDAO() *ReviewTaskDAO { return &ReviewTaskDAO{} }

// FindByRepoID 分页查询仓库的审查任务
func (d *ReviewTaskDAO) FindByRepoID(repoID uint, status string, page, pageSize int) ([]po.ReviewTask, int64, error) {
	var tasks []po.ReviewTask
	var total int64
	q := DB.Model(new(po.ReviewTask)).Where("repo_id = ?", repoID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q.Count(&total)
	return tasks, total, q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error
}

// FindByProviderConfigID 查询 MR 的审查任务
func (d *ReviewTaskDAO) FindByProviderConfigID(providerConfigID uint, mrIID string, page, pageSize int) ([]po.ReviewTask, int64, error) {
	var tasks []po.ReviewTask
	var total int64
	q := DB.Model(new(po.ReviewTask)).Where("provider_config_id = ? AND mri_id = ?", providerConfigID, mrIID)
	q.Count(&total)
	return tasks, total, q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error
}

// UpdateStatus 更新任务状态
func (d *ReviewTaskDAO) UpdateStatus(id uint, status, riskLevel, errMsg string) error {
	return DB.Model(new(po.ReviewTask)).Where("id = ?", id).
		Updates(map[string]interface{}{"status": status, "risk_level": riskLevel, "error_message": errMsg}).Error
}

// FindByTimeRange 按时间范围查询
func (d *ReviewTaskDAO) FindByTimeRange(repoID uint, since, until time.Time) ([]po.ReviewTask, error) {
	var tasks []po.ReviewTask
	q := DB.Model(new(po.ReviewTask)).Where("created_at >= ? AND created_at <= ?", since, until)
	if repoID > 0 {
		q = q.Where("repo_id = ?", repoID)
	}
	return tasks, q.Order("created_at ASC").Find(&tasks).Error
}
