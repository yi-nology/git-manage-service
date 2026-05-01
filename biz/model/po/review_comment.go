package po

import (
	"gorm.io/gorm"
)

type ReviewComment struct {
	gorm.Model
	TaskID            uint   `gorm:"index"`
	FindingID         uint
	ProviderCommentID string `gorm:"size:128"`
	CommentType       string `gorm:"size:32"`
	FilePath          string
	LineNumber        int
	Body              string `gorm:"type:text"`
	Status            string `gorm:"size:32"`
}

func (ReviewComment) TableName() string { return "review_comments" }

type MergeCheckResult struct {
	gorm.Model
	RepoID    uint   `gorm:"index"`
	MRIID     string `gorm:"size:64;index"`
	CommitSHA string `gorm:"size:128;index"`
	CheckType string `gorm:"size:64"`
	Status    string `gorm:"size:32"`
	RiskLevel string `gorm:"size:32"`
	Message   string `gorm:"type:text"`
}

func (MergeCheckResult) TableName() string { return "merge_check_results" }
