package po

import (
	"time"

	"gorm.io/gorm"
)

const (
	MirrorTypePull = "pull"
	MirrorTypePush = "push"

	MirrorStatusActive  = "active"
	MirrorStatusSyncing = "syncing"
	MirrorStatusFailed  = "failed"
	MirrorStatusPaused  = "paused"
)

type Mirror struct {
	gorm.Model
	RepoID       uint       `gorm:"not null;index" json:"repoId"`
	MirrorType   string     `gorm:"size:10;not null" json:"mirrorType"`
	RemoteURL    string     `gorm:"size:512;not null" json:"remoteUrl"`
	RemoteName   string     `gorm:"size:100;default:''" json:"remoteName"`
	CredentialID *uint      `json:"credentialId"`
	BranchFilter string     `gorm:"size:512;default:''" json:"branchFilter"`
	SyncInterval int        `gorm:"default:600" json:"syncInterval"`
	CronExpr     string     `gorm:"size:100;default:''" json:"cronExpr"`
	SyncOnPush   bool       `gorm:"default:false" json:"syncOnPush"`
	GitForce     bool       `gorm:"default:false" json:"gitForce"`
	GitPrune     bool       `gorm:"default:true" json:"gitPrune"`
	GitTags      bool       `gorm:"default:true" json:"gitTags"`
	Enabled      bool       `gorm:"default:true" json:"enabled"`
	Status       string     `gorm:"size:20;default:'active'" json:"status"`
	LastSyncAt   *time.Time `json:"lastSyncAt"`
	LastError    string     `json:"lastError"`
	RetryCount   int        `gorm:"default:0" json:"retryCount"`
	NextSyncAt   *time.Time `gorm:"index" json:"nextSyncAt"`
	WebhookToken string     `gorm:"size:255;index" json:"webhookToken"`

	Repo       Repo        `gorm:"foreignKey:RepoID" json:"repo"`
	Credential *Credential `gorm:"foreignKey:CredentialID" json:"credential,omitempty"`
}

func (Mirror) TableName() string {
	return "mirrors"
}
