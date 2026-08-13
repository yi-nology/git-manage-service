package db

import (
	"encoding/json"
	"time"

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

func (d *ReviewFindingDAO) DeleteByTaskID(taskID uint) error {
	return DB.Where("task_id = ?", taskID).Delete(&po.ReviewFinding{}).Error
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

func (d *ReviewFindingDAO) FindByTimeRange(repoID uint, since, until time.Time) ([]po.ReviewFinding, error) {
	var findings []po.ReviewFinding
	q := DB.Model(&po.ReviewFinding{}).
		Joins("JOIN review_tasks ON review_tasks.id = review_findings.task_id").
		Where("review_findings.created_at >= ? AND review_findings.created_at <= ?", since, until)
	if repoID > 0 {
		q = q.Where("review_tasks.repo_id = ?", repoID)
	}
	err := q.Find(&findings).Error
	return findings, err
}

func (d *ReviewFindingDAO) UpdateFeedback(id uint, feedback string) error {
	var finding po.ReviewFinding
	if err := DB.First(&finding, id).Error; err != nil {
		return err
	}

	if finding.RawPayload == "" {
		return DB.Model(&finding).Update("raw_payload", feedback).Error
	}

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(finding.RawPayload), &raw); err != nil {
		return DB.Model(&finding).Update("raw_payload", feedback).Error
	}
	raw["feedback"] = feedback
	updated, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return DB.Model(&finding).Update("raw_payload", string(updated)).Error
}
