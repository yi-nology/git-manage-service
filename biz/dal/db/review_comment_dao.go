package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type ReviewCommentDAO struct{}

func NewReviewCommentDAO() *ReviewCommentDAO { return &ReviewCommentDAO{} }

func (d *ReviewCommentDAO) Create(c *po.ReviewComment) error {
	return DB.Create(c).Error
}

func (d *ReviewCommentDAO) BatchCreate(comments []po.ReviewComment) error {
	return DB.Create(&comments).Error
}

func (d *ReviewCommentDAO) FindByTaskID(taskID uint) ([]po.ReviewComment, error) {
	var comments []po.ReviewComment
	err := DB.Where("task_id = ?", taskID).Find(&comments).Error
	return comments, err
}

type MergeCheckResultDAO struct{}

func NewMergeCheckResultDAO() *MergeCheckResultDAO { return &MergeCheckResultDAO{} }

func (d *MergeCheckResultDAO) Create(r *po.MergeCheckResult) error {
	return DB.Create(r).Error
}

func (d *MergeCheckResultDAO) FindLatest(repoID uint, mrIID, commitSHA string) (*po.MergeCheckResult, error) {
	var r po.MergeCheckResult
	err := DB.Where("repo_id = ? AND mr_iid = ? AND commit_sha = ? AND check_type = ?",
		repoID, mrIID, commitSHA, "code_review").Order("created_at DESC").First(&r).Error
	return &r, err
}
