package api

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type RegisterRepoReq struct {
	Name                string                     `json:"name"`
	Path                string                     `json:"path"`
	RemoteURL           string                     `json:"remote_url"`
	AuthType            string                     `json:"auth_type"`
	AuthKey             string                     `json:"auth_key"`
	AuthSecret          string                     `json:"auth_secret"`
	Remotes             []domain.GitRemote         `json:"remotes"`
	RemoteAuths         map[string]domain.AuthInfo `json:"remote_auths"`
	DefaultCredentialID uint                       `json:"default_credential_id"`
	RemoteCredentials   map[string]uint            `json:"remote_credentials"`
}

type ScanRepoReq struct {
	Path string `json:"path"`
}

type CloneRepoReq struct {
	RemoteURL        string `json:"remote_url"`
	LocalPath        string `json:"local_path"`
	Name             string `json:"name"`
	AuthType         string `json:"auth_type"`
	AuthKey          string `json:"auth_key"`
	AuthSecret       string `json:"auth_secret"`
	SSHKeyID         uint   `json:"ssh_key_id"`
	CredentialID     uint   `json:"credential_id"`
	ProviderConfigID uint   `json:"provider_config_id"`
	PlatformOwner    string `json:"platform_owner"`
	PlatformRepo     string `json:"platform_repo"`
}

// ScanDirectoryReq 扫描目录请求
type ScanDirectoryReq struct {
	Path      string `json:"path"`      // 要扫描的父目录
	Depth     int    `json:"depth"`     // 扫描深度，默认 2
	Recursive bool   `json:"recursive"` // 是否递归扫描
}

// ScannedRepo 扫描到的仓库信息
type ScannedRepo struct {
	Name          string             `json:"name"`
	Path          string             `json:"path"`
	Remotes       []domain.GitRemote `json:"remotes"`
	CurrentBranch string             `json:"current_branch"`
	LastCommit    string             `json:"last_commit"`
	HasChanges    bool               `json:"has_changes"`
}

// ScanDirectoryResp 扫描目录响应
type ScanDirectoryResp struct {
	Repos []ScannedRepo `json:"repos"`
	Total int           `json:"total"`
}

// BatchCreateRepoReq 批量注册仓库请求
type BatchCreateRepoReq struct {
	Repos []BatchRepoItem `json:"repos"`
}

// BatchRepoItem 批量注册的仓库项
type BatchRepoItem struct {
	Name                string `json:"name"`
	Path                string `json:"path"`
	DefaultCredentialID uint   `json:"default_credential_id,omitempty"`
}

// BatchCreateRepoResp 批量注册仓库响应
type BatchCreateRepoResp struct {
	Success []RepoDTO         `json:"success"`
	Failed  []BatchFailedItem `json:"failed"`
}

// BatchFailedItem 失败项
type BatchFailedItem struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

type RepoDTO struct {
	ID                  uint                       `json:"id"`
	Key                 string                     `json:"key"`
	Name                string                     `json:"name"`
	Path                string                     `json:"path"`
	RemoteURL           string                     `json:"remote_url"`
	AuthType            string                     `json:"auth_type"`
	AuthKey             string                     `json:"auth_key"`
	AuthSecret          string                     `json:"auth_secret"`
	RemoteAuths         map[string]domain.AuthInfo `json:"remote_auths"`
	DefaultCredentialID uint                       `json:"default_credential_id,omitempty"`
	RemoteCredentials   map[string]uint            `json:"remote_credentials,omitempty"`
	ProviderConfigID    uint                       `json:"provider_config_id,omitempty"`
	PlatformRepoID      string                     `json:"platform_repo_id,omitempty"`
	PlatformOwner       string                     `json:"platform_owner,omitempty"`
	PlatformRepo        string                     `json:"platform_repo,omitempty"`
	CreatedAt           time.Time                  `json:"created_at"`
	UpdatedAt           time.Time                  `json:"updated_at"`
}

func NewRepoDTO(r po.Repo) RepoDTO {
	return RepoDTO{
		ID:                  r.ID,
		Key:                 r.Key,
		Name:                r.Name,
		Path:                r.Path,
		RemoteURL:           r.RemoteURL,
		AuthType:            r.AuthType,
		AuthKey:             r.AuthKey,
		AuthSecret:          r.AuthSecret,
		RemoteAuths:         r.RemoteAuths,
		DefaultCredentialID: r.DefaultCredentialID,
		RemoteCredentials:   r.RemoteCredentials,
		ProviderConfigID:    r.ProviderConfigID,
		PlatformRepoID:      r.PlatformRepoID,
		PlatformOwner:       r.PlatformOwner,
		PlatformRepo:        r.PlatformRepo,
		CreatedAt:           r.CreatedAt,
		UpdatedAt:           r.UpdatedAt,
	}
}
