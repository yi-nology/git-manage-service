package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type ReviewCommentDAO struct{ BaseDAO[po.ReviewComment] }

func NewReviewCommentDAO() *ReviewCommentDAO { return &ReviewCommentDAO{} }

// FindByTaskID 查询任务的所有评论
func (d *ReviewCommentDAO) FindByTaskID(taskID uint) ([]po.ReviewComment, error) {
	var comments []po.ReviewComment
	return comments, DB.Where("task_id = ?", taskID).Find(&comments).Error
}

// FindSummaryCommentsByMRIID 查询 MR 的 summary 类型评论
func (d *ReviewCommentDAO) FindSummaryCommentsByMRIID(providerConfigID uint, mrIID string) ([]po.ReviewComment, error) {
	var comments []po.ReviewComment
	return comments, DB.Joins("JOIN review_tasks ON review_tasks.id = review_comments.task_id").
		Where("review_tasks.provider_config_id = ? AND review_tasks.mri_id = ? AND review_comments.comment_type = ?",
			providerConfigID, mrIID, "summary").
		Order("review_comments.created_at ASC").Find(&comments).Error
}

// MergeCheckResultDAO 合并检查结果
type MergeCheckResultDAO struct{ BaseDAO[po.MergeCheckResult] }

func NewMergeCheckResultDAO() *MergeCheckResultDAO { return &MergeCheckResultDAO{} }

// FindLatest 查询最新的检查结果
func (d *MergeCheckResultDAO) FindLatest(repoID uint, mrIID, commitSHA string) (*po.MergeCheckResult, error) {
	var r po.MergeCheckResult
	return &r, DB.Where("repo_id = ? AND mr_iid = ? AND commit_sha = ? AND check_type = ?",
		repoID, mrIID, commitSHA, "code_review").Order("created_at DESC").First(&r).Error
}
