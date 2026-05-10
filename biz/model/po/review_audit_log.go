package po

import (
	"gorm.io/gorm"
)

type ReviewAuditLog struct {
	gorm.Model
	TaskID       uint   `gorm:"index;not null"`
	Action       string `gorm:"size:50;not null;index"`
	Status       string `gorm:"size:20;index"`
	ErrorMessage string `gorm:"type:text"`
	Duration     int    `gorm:"default:0"`
	Metadata     string `gorm:"type:text"`
}

func (ReviewAuditLog) TableName() string { return "review_audit_logs" }

type WebhookEventRule struct {
	gorm.Model
	Name        string `gorm:"size:255;not null"`
	EventType   string `gorm:"size:100;not null;index"`
	Description string `gorm:"type:text"`
	MatchRules  string `gorm:"type:text;not null"`
	IsActive    bool   `gorm:"default:true;index"`
	Priority    int    `gorm:"default:0"`
}

func (WebhookEventRule) TableName() string { return "webhook_event_rules" }
