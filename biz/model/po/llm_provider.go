package po

import (
	"github.com/yi-nology/git-manage-service/biz/utils"
	"gorm.io/gorm"
)

type LLMProvider struct {
	gorm.Model
	Name      string `gorm:"uniqueIndex;size:64;not null"`
	Type      string `gorm:"size:32;not null"`
	BaseURL   string `gorm:"size:500"`
	APIKey    string `gorm:"size:500"`
	AIModel   string `gorm:"size:128"`
	MaxTokens int    `gorm:"default:4096"`
	IsDefault bool   `gorm:"default:false;index"`
	PresetID  string `gorm:"size:64;index"`
	Protocol  string `gorm:"size:32"`
}

func (LLMProvider) TableName() string { return "llm_providers" }

func (p *LLMProvider) BeforeSave(tx *gorm.DB) error {
	if p.APIKey != "" {
		enc, err := utils.Encrypt(p.APIKey)
		if err != nil {
			return err
		}
		p.APIKey = enc
	}
	return nil
}

func (p *LLMProvider) AfterFind(tx *gorm.DB) error {
	if p.APIKey != "" {
		dec, err := utils.Decrypt(p.APIKey)
		if err == nil {
			p.APIKey = dec
		}
	}
	return nil
}
