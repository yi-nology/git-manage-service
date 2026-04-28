package api

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type RepoProviderBindingDTO struct {
	ID               uint      `json:"id"`
	RepoID           uint      `json:"repo_id"`
	RepoKey          string    `json:"repo_key"`
	RepoName         string    `json:"repo_name"`
	ProviderConfigID uint      `json:"provider_config_id"`
	ProviderName     string    `json:"provider_name"`
	Platform         string    `json:"platform"`
	PlatformOwner    string    `json:"platform_owner"`
	PlatformRepo     string    `json:"platform_repo"`
	PlatformRepoID   string    `json:"platform_repo_id"`
	RemoteName       string    `json:"remote_name"`
	IsPrimary        bool      `json:"is_primary"`
	WebhookID        string    `json:"webhook_id"`
	WebhookURL       string    `json:"webhook_url"`
	HasWebhook       bool      `json:"has_webhook"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateBindingReq struct {
	RepoKey          string `json:"repo_key" binding:"required"`
	ProviderConfigID uint   `json:"provider_config_id" binding:"required"`
	PlatformOwner    string `json:"platform_owner" binding:"required"`
	PlatformRepo     string `json:"platform_repo" binding:"required"`
	RemoteName       string `json:"remote_name"`
	IsPrimary        bool   `json:"is_primary"`
	RegisterWebhook  bool   `json:"register_webhook"`
}

type UpdateBindingReq struct {
	RemoteName     *string `json:"remote_name"`
	IsPrimary      *bool   `json:"is_primary"`
	PlatformRepoID *string `json:"platform_repo_id"`
}

type ListBindingsReq struct {
	RepoKey          string `query:"repo_key"`
	ProviderConfigID uint   `query:"provider_config_id"`
}

type AutoDetectReq struct {
	RepoKey string `json:"repo_key"`
}

type BindingSuggestion struct {
	ProviderConfigID uint   `json:"provider_config_id"`
	Platform         string `json:"platform"`
	PlatformOwner    string `json:"platform_owner"`
	PlatformRepo     string `json:"platform_repo"`
	RemoteName       string `json:"remote_name"`
	RemoteURL        string `json:"remote_url"`
	Confidence       string `json:"confidence"`
	MatchSource      string `json:"match_source"`
}

type AutoDetectResp struct {
	Suggestions []BindingSuggestion `json:"suggestions"`
}

func NewBindingDTO(b po.RepoProviderBinding) RepoProviderBindingDTO {
	dto := RepoProviderBindingDTO{
		ID:               b.ID,
		RepoID:           b.RepoID,
		ProviderConfigID: b.ProviderConfigID,
		PlatformOwner:    b.PlatformOwner,
		PlatformRepo:     b.PlatformRepo,
		PlatformRepoID:   b.PlatformRepoID,
		RemoteName:       b.RemoteName,
		IsPrimary:        b.IsPrimary,
		WebhookID:        b.WebhookID,
		WebhookURL:       b.WebhookURL,
		HasWebhook:       b.WebhookID != "",
		Status:           b.Status,
		CreatedAt:        b.CreatedAt,
		UpdatedAt:        b.UpdatedAt,
	}
	if b.Repo.ID > 0 {
		dto.RepoKey = b.Repo.Key
		dto.RepoName = b.Repo.Name
	}
	if b.ProviderConfig.ID > 0 {
		dto.ProviderName = b.ProviderConfig.Name
		dto.Platform = b.ProviderConfig.Platform
	}
	return dto
}
