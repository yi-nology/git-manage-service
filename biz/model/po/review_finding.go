package po

import (
	"gorm.io/gorm"
)

type ReviewFinding struct {
	gorm.Model
	TaskID      uint   `gorm:"index"`
	Source      string `gorm:"size:32;index"`
	RuleID      string `gorm:"size:128"`
	Severity    string `gorm:"size:32;index"`
	FilePath    string `gorm:"type:text"`
	OldLine     int
	NewLine     int
	Title       string `gorm:"size:512"`
	Message     string `gorm:"type:text"`
	Suggestion  string `gorm:"type:text"`
	Fingerprint string `gorm:"size:128;index"`
	RawPayload  string `gorm:"type:text"`
}

func (ReviewFinding) TableName() string { return "review_findings" }
