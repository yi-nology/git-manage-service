package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type ReviewFindingDAO struct{}

func NewReviewFindingDAO() *ReviewFindingDAO { return &ReviewFindingDAO{} }

func (d *ReviewFindingDAO) Create(f *po.ReviewFinding) error {
	return DB.Create(f).Error
}

func (d *ReviewFindingDAO) BatchCreate(findings []po.ReviewFinding) error {
	return DB.Create(&findings).Error
}

func (d *ReviewFindingDAO) FindByTaskID(taskID uint, severity, source string) ([]po.ReviewFinding, error) {
	var findings []po.ReviewFinding
	q := DB.Where("task_id = ?", taskID)
	if severity != "" {
		q = q.Where("severity = ?", severity)
	}
	if source != "" {
		q = q.Where("source = ?", source)
	}
	err := q.Order("severity DESC, id ASC").Find(&findings).Error
	return findings, err
}

func (d *ReviewFindingDAO) CountByTaskID(taskID uint) (int64, error) {
	var count int64
	err := DB.Model(&po.ReviewFinding{}).Where("task_id = ?", taskID).Count(&count).Error
	return count, err
}

func (d *ReviewFindingDAO) ExistsByFingerprint(taskID uint, fingerprint string) (bool, error) {
	var count int64
	err := DB.Model(&po.ReviewFinding{}).Where("task_id = ? AND fingerprint = ?", taskID, fingerprint).Count(&count).Error
	return count > 0, err
}
