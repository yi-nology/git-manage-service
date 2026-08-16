package db

import (
	"encoding/json"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type ReviewFindingDAO struct{ BaseDAO[po.ReviewFinding] }

func NewReviewFindingDAO() *ReviewFindingDAO { return &ReviewFindingDAO{} }

// DeleteByTaskID 删除任务的所有发现
func (d *ReviewFindingDAO) DeleteByTaskID(taskID uint) error {
	return DB.Where("task_id = ?", taskID).Delete(new(po.ReviewFinding)).Error
}

// FindByTaskID 按筛选条件查询任务发现
func (d *ReviewFindingDAO) FindByTaskID(taskID uint, severity, source string) ([]po.ReviewFinding, error) {
	var findings []po.ReviewFinding
	q := DB.Where("task_id = ?", taskID)
	if severity != "" {
		q = q.Where("severity = ?", severity)
	}
	if source != "" {
		q = q.Where("source = ?", source)
	}
	return findings, q.Order("severity DESC, id ASC").Find(&findings).Error
}

// CountByTaskID 统计任务发现数
func (d *ReviewFindingDAO) CountByTaskID(taskID uint) (int64, error) {
	var count int64
	return count, DB.Model(new(po.ReviewFinding)).Where("task_id = ?", taskID).Count(&count).Error
}

// CountByTaskIDs 批量统计（避免 N+1 查询）
func (d *ReviewFindingDAO) CountByTaskIDs(taskIDs []uint) (map[uint]int64, error) {
	result := make(map[uint]int64)
	if len(taskIDs) == 0 {
		return result, nil
	}
	type row struct {
		TaskID uint  `gorm:"column:task_id"`
		Count  int64 `gorm:"column:cnt"`
	}
	var rows []row
	if err := DB.Model(new(po.ReviewFinding)).Select("task_id, count(*) as cnt").
		Where("task_id IN ?", taskIDs).Group("task_id").Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		result[r.TaskID] = r.Count
	}
	return result, nil
}

// FindByTimeRange 按时间范围查询
func (d *ReviewFindingDAO) FindByTimeRange(repoID uint, since, until time.Time) ([]po.ReviewFinding, error) {
	var findings []po.ReviewFinding
	q := DB.Model(new(po.ReviewFinding)).
		Joins("JOIN review_tasks ON review_tasks.id = review_findings.task_id").
		Where("review_findings.created_at >= ? AND review_findings.created_at <= ?", since, until)
	if repoID > 0 {
		q = q.Where("review_tasks.repo_id = ?", repoID)
	}
	return findings, q.Find(&findings).Error
}

// UpdateFeedback 更新发现的反馈信息
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
