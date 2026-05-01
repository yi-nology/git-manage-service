package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type ReviewTaskDAO struct{}

func NewReviewTaskDAO() *ReviewTaskDAO { return &ReviewTaskDAO{} }

func (d *ReviewTaskDAO) Create(t *po.ReviewTask) error {
	return DB.Create(t).Error
}

func (d *ReviewTaskDAO) FindByID(id uint) (*po.ReviewTask, error) {
	var t po.ReviewTask
	err := DB.First(&t, id).Error
	return &t, err
}

func (d *ReviewTaskDAO) FindByRepoID(repoID uint, status string, page, pageSize int) ([]po.ReviewTask, int64, error) {
	var tasks []po.ReviewTask
	var total int64
	q := DB.Model(&po.ReviewTask{}).Where("repo_id = ?", repoID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	q.Count(&total)
	err := q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tasks).Error
	return tasks, total, err
}

func (d *ReviewTaskDAO) FindByMRIID(repoID uint, mrIID string) ([]po.ReviewTask, error) {
	var tasks []po.ReviewTask
	err := DB.Where("repo_id = ? AND mr_iid = ?", repoID, mrIID).Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

func (d *ReviewTaskDAO) Save(t *po.ReviewTask) error {
	return DB.Save(t).Error
}

func (d *ReviewTaskDAO) UpdateStatus(id uint, status, riskLevel, errMsg string) error {
	updates := map[string]interface{}{"status": status, "risk_level": riskLevel, "error_message": errMsg}
	return DB.Model(&po.ReviewTask{}).Where("id = ?", id).Updates(updates).Error
}
