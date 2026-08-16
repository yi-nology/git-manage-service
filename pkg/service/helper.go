// Package service 提供 service 层通用辅助函数，消除重复的 DB 查找 + 错误包装模式。
package service

import (
	"fmt"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

// GetRepoByKey 通过唯一 key 查找仓库，找不到时返回包装后的错误。
func GetRepoByKey(key string) (*po.Repo, error) {
	repo, err := db.NewRepoDAO().FindByKey(key)
	if err != nil {
		return nil, fmt.Errorf("repo not found: %w", err)
	}
	return repo, nil
}

// GetProviderByID 通过 ID 查找 provider 配置，找不到时返回包装后的错误。
func GetProviderByID(id uint) (*po.ProviderConfig, error) {
	cfg, err := db.NewProviderConfigDAO().FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("provider config not found: %w", err)
	}
	return cfg, nil
}

// GetCredentialByID 通过 ID 查找凭证，找不到时返回包装后的错误。
func GetCredentialByID(id uint) (*po.Credential, error) {
	cred, err := db.NewCredentialDAO().FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("credential not found: %w", err)
	}
	return cred, nil
}

// GetSSHKeyByID 通过 ID 查找 SSH 密钥，找不到时返回包装后的错误。
func GetSSHKeyByID(id uint) (*po.SSHKey, error) {
	key, err := db.NewSSHKeyDAO().FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("SSH key not found: %w", err)
	}
	return key, nil
}

// GetLLMProviderByID 通过 ID 查找 LLM Provider，找不到时返回包装后的错误。
func GetLLMProviderByID(id uint) (*po.LLMProvider, error) {
	p, err := db.NewLLMProviderDAO().FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("LLM provider not found: %w", err)
	}
	return p, nil
}

// GetMirrorByID 通过 ID 查找镜像，找不到时返回包装后的错误。
func GetMirrorByID(id uint) (*po.Mirror, error) {
	m, err := db.NewMirrorDAO().FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("mirror not found: %w", err)
	}
	return m, nil
}

// GetNotificationChannelByID 通过 ID 查找通知渠道，找不到时返回包装后的错误。
func GetNotificationChannelByID(id uint) (*po.NotificationChannel, error) {
	ch, err := db.NewNotificationChannelDAO().FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("notification channel not found: %w", err)
	}
	return ch, nil
}
