package po

import (
	"gorm.io/gorm"
)

type ReviewCLIConfig struct {
	gorm.Model
	Name       string `gorm:"size:100;not null"`
	CLIType    string `gorm:"size:50;not null;index"`
	ExecPath   string `gorm:"size:500;not null"`
	ConfigJSON string `gorm:"type:text"`
	IsActive   bool   `gorm:"default:true"`
}

func (ReviewCLIConfig) TableName() string { return "review_cli_configs" }
