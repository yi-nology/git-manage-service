package po

import (
	"encoding/json"
	"log"

	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/utils"
	"gorm.io/gorm"
)

type Repo struct {
	gorm.Model
	Key        string `gorm:"uniqueIndex" json:"key"`
	Name       string `gorm:"uniqueIndex" json:"name"`
	Path       string `json:"path"`
	RemoteURL  string `json:"remote_url"`
	AuthType   string `json:"auth_type"`   // ssh, http, none (deprecated, kept for compatibility)
	AuthKey    string `json:"auth_key"`    // SSH Key Path or Username (deprecated)
	AuthSecret string `json:"auth_secret"` // Passphrase or Password (Encrypted in DB) (deprecated)

	RemoteAuthsJSON string                     `json:"-"`                     // Stored in DB (deprecated)
	RemoteAuths     map[string]domain.AuthInfo `gorm:"-" json:"remote_auths"` // Memory & API (deprecated)

	DefaultCredentialID   uint            `gorm:"index" json:"default_credential_id"`
	RemoteCredentialsJSON string          `json:"-"`
	RemoteCredentials     map[string]uint `gorm:"-" json:"remote_credentials"`

	ProviderConfigID uint   `gorm:"index" json:"provider_config_id"`
	PlatformRepoID   string `gorm:"size:100" json:"platform_repo_id"`
	PlatformOwner    string `gorm:"size:200;index:idx_repo_platform_owner_repo,priority:1" json:"platform_owner"`
	PlatformRepo     string `gorm:"size:200;index:idx_repo_platform_owner_repo,priority:2" json:"platform_repo"`
	AuthorIdentityID *uint  `gorm:"index" json:"author_identity_id"`
}

func (Repo) TableName() string {
	return "repos"
}

func (r *Repo) BeforeSave(tx *gorm.DB) (err error) {
	// Encrypt main secret
	if r.AuthSecret != "" {
		enc, err := utils.Encrypt(r.AuthSecret)
		if err != nil {
			return err
		}
		r.AuthSecret = enc
	}

	// Handle RemoteAuths (deprecated, kept for compatibility)
	if r.RemoteAuths != nil {
		// Encrypt secrets in map
		encryptedMap := make(map[string]domain.AuthInfo)
		for k, v := range r.RemoteAuths {
			if v.Secret != "" {
				enc, err := utils.Encrypt(v.Secret)
				if err != nil {
					return err
				}
				v.Secret = enc
			}
			encryptedMap[k] = v
		}
		bytes, err := json.Marshal(encryptedMap)
		if err != nil {
			return err
		}
		r.RemoteAuthsJSON = string(bytes)
	}

	// Handle RemoteCredentials (new)
	if r.RemoteCredentials != nil {
		bytes, err := json.Marshal(r.RemoteCredentials)
		if err != nil {
			return err
		}
		r.RemoteCredentialsJSON = string(bytes)
	}

	return nil
}

func (r *Repo) AfterFind(tx *gorm.DB) (err error) {
	// Decrypt main secret
	if r.AuthSecret != "" {
		dec, err := utils.Decrypt(r.AuthSecret)
		if err == nil {
			r.AuthSecret = dec
		} else {
			// 密钥与加密时不一致：置空并记录，绝不能把密文当明文外发
			log.Printf("[Repo] decrypt auth secret failed (id=%d key=%s): %v", r.ID, r.Key, err)
			r.AuthSecret = ""
		}
	}

	// Handle RemoteAuths (deprecated, kept for compatibility)
	if r.RemoteAuthsJSON != "" {
		var encryptedMap map[string]domain.AuthInfo
		if err := json.Unmarshal([]byte(r.RemoteAuthsJSON), &encryptedMap); err == nil {
			decryptedMap := make(map[string]domain.AuthInfo)
			for k, v := range encryptedMap {
				if v.Secret != "" {
					dec, err := utils.Decrypt(v.Secret)
					if err == nil {
						v.Secret = dec
					}
				}
				decryptedMap[k] = v
			}
			r.RemoteAuths = decryptedMap
		}
	}

	// Handle RemoteCredentials (new)
	if r.RemoteCredentialsJSON != "" {
		var remoteCreds map[string]uint
		if err := json.Unmarshal([]byte(r.RemoteCredentialsJSON), &remoteCreds); err == nil {
			r.RemoteCredentials = remoteCreds
		}
	}

	return nil
}
