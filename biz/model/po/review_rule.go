package po

import "time"

type ReviewRule struct {
	ID          string    `gorm:"primaryKey;size:64" json:"id"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	Severity    string    `gorm:"size:32;not null" json:"severity"`
	Category    string    `gorm:"size:64" json:"category"`
	Enabled     bool      `gorm:"default:true" json:"enabled"`
	SortOrder   int       `gorm:"default:0" json:"sort_order"`
	RuleType    string    `gorm:"size:32;not null;default:builtin" json:"rule_type"`
	PromptText  string    `gorm:"type:text" json:"prompt_text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (ReviewRule) TableName() string {
	return "review_rules"
}
