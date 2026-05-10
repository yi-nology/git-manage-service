package po

import (
	"gorm.io/gorm"
)

type ReviewRepoConfig struct {
	gorm.Model
	ProviderConfigID     uint   `gorm:"index;not null"`
	PlatformOwner        string `gorm:"size:200;not null"`
	PlatformRepo         string `gorm:"size:200;not null"`
	Enabled              bool   `gorm:"default:true"`
	BlockOnHigh          bool   `gorm:"default:true"`
	AutoReviewOnMR       bool   `gorm:"default:true"`
	LLMProvider          string `gorm:"size:64"`
	MaxFiles             int    `gorm:"default:50"`
	MaxDiffLines         int    `gorm:"default:3000"`
	RuleOverridesJSON    string `gorm:"type:text"`
	ScopeNote            string `gorm:"size:500"`
	PromptPrefixOverride string `gorm:"type:text"`
	PromptIntentOverride string `gorm:"type:text"`
	ReviewMode           string `gorm:"size:32;default:'llm'"`
	CLIConfigJSON        string `gorm:"type:text"`
	CustomPrompt         string `gorm:"type:text"`
	UseCustomPrompt      bool   `gorm:"default:false"`
	ExcludeFileTypes     string `gorm:"type:text"`
	IgnorePatterns       string `gorm:"type:text"`
}

func (ReviewRepoConfig) TableName() string { return "review_repo_configs" }
