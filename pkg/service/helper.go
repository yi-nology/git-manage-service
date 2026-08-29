// Package service 提供 service 层通用辅助函数，消除重复的 DB 查找 + 错误包装模式。
package service

import (
	"fmt"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

// lookup 通过给定的 DAO 查找函数定位实体；err 非空时统一包装为
// "<label> not found" 返回。K 为查找键类型（uint 主键或 string key），
// 类型与错误语义由 find 函数推断，无需在调用处显式指定。
func lookup[T any, K any](key K, find func(K) (*T, error), label string) (*T, error) {
	v, err := find(key)
	if err != nil {
		return nil, fmt.Errorf("%s not found: %w", label, err)
	}
	return v, nil
}

// GetRepoByKey 通过唯一 key 查找仓库，找不到时返回包装后的错误。
func GetRepoByKey(key string) (*po.Repo, error) {
	return lookup(key, db.NewRepoDAO().FindByKey, "repo")
}

// GetProviderByID 通过 ID 查找 provider 配置，找不到时返回包装后的错误。
func GetProviderByID(id uint) (*po.ProviderConfig, error) {
	return lookup(id, db.NewProviderConfigDAO().FindByID, "provider config")
}

// GetCredentialByID 通过 ID 查找凭证，找不到时返回包装后的错误。
func GetCredentialByID(id uint) (*po.Credential, error) {
	return lookup(id, db.NewCredentialDAO().FindByID, "credential")
}

// GetSSHKeyByID 通过 ID 查找 SSH 密钥，找不到时返回包装后的错误。
func GetSSHKeyByID(id uint) (*po.SSHKey, error) {
	return lookup(id, db.NewSSHKeyDAO().FindByID, "SSH key")
}

// GetLLMProviderByID 通过 ID 查找 LLM Provider，找不到时返回包装后的错误。
func GetLLMProviderByID(id uint) (*po.LLMProvider, error) {
	return lookup(id, db.NewLLMProviderDAO().FindByID, "LLM provider")
}

// GetMirrorByID 通过 ID 查找镜像，找不到时返回包装后的错误。
func GetMirrorByID(id uint) (*po.Mirror, error) {
	return lookup(id, db.NewMirrorDAO().FindByID, "mirror")
}

// GetNotificationChannelByID 通过 ID 查找通知渠道，找不到时返回包装后的错误。
func GetNotificationChannelByID(id uint) (*po.NotificationChannel, error) {
	return lookup(id, db.NewNotificationChannelDAO().FindByID, "notification channel")
}
