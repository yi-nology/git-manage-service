package po

import "gorm.io/gorm"

type RepoProviderBinding struct {
	gorm.Model
	RepoID           uint `gorm:"uniqueIndex:idx_repo_provider;index"`
	ProviderConfigID uint `gorm:"uniqueIndex:idx_repo_provider;index"`
	PlatformOwner    string
	PlatformRepo     string
	PlatformRepoID   string `gorm:"size:100"`
	RemoteName       string `gorm:"size:100"`
	IsPrimary        bool   `gorm:"default:false;index"`
	WebhookID        string `gorm:"size:200"`
	WebhookURL       string `gorm:"size:500"`
	Status           string `gorm:"size:20;default:active"`

	Repo           Repo           `gorm:"foreignKey:RepoID"`
	ProviderConfig ProviderConfig `gorm:"foreignKey:ProviderConfigID"`
}

func (RepoProviderBinding) TableName() string { return "repo_provider_bindings" }
