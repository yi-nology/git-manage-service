package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type ReviewRepoConfigDAO struct{ BaseDAO[po.ReviewRepoConfig] }

func NewReviewRepoConfigDAO() *ReviewRepoConfigDAO { return &ReviewRepoConfigDAO{} }

// FindByRemoteRepo 按平台配置和 owner/repo 查询
func (d *ReviewRepoConfigDAO) FindByRemoteRepo(providerConfigID uint, platformOwner, platformRepo string) (*po.ReviewRepoConfig, error) {
	var c po.ReviewRepoConfig
	return &c, DB.Where("provider_config_id = ? AND platform_owner = ? AND platform_repo = ?",
		providerConfigID, platformOwner, platformRepo).First(&c).Error
}

// Upsert 存在则更新，否则创建
func (d *ReviewRepoConfigDAO) Upsert(c *po.ReviewRepoConfig) error {
	var existing po.ReviewRepoConfig
	err := DB.Where("provider_config_id = ? AND platform_owner = ? AND platform_repo = ?",
		c.ProviderConfigID, c.PlatformOwner, c.PlatformRepo).First(&existing).Error
	if err != nil {
		return DB.Create(c).Error
	}
	c.ID = existing.ID
	c.CreatedAt = existing.CreatedAt
	return DB.Save(c).Error
}

// FindByProviderConfigID 按 provider 配置查询
func (d *ReviewRepoConfigDAO) FindByProviderConfigID(providerConfigID uint) ([]po.ReviewRepoConfig, error) {
	var configs []po.ReviewRepoConfig
	return configs, DB.Where("provider_config_id = ?", providerConfigID).Find(&configs).Error
}
