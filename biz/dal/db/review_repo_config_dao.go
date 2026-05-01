package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type ReviewRepoConfigDAO struct{}

func NewReviewRepoConfigDAO() *ReviewRepoConfigDAO { return &ReviewRepoConfigDAO{} }

func (d *ReviewRepoConfigDAO) Create(c *po.ReviewRepoConfig) error {
	return DB.Create(c).Error
}

func (d *ReviewRepoConfigDAO) FindByRemoteRepo(providerConfigID uint, platformOwner, platformRepo string) (*po.ReviewRepoConfig, error) {
	var c po.ReviewRepoConfig
	err := DB.Where("provider_config_id = ? AND platform_owner = ? AND platform_repo = ?", providerConfigID, platformOwner, platformRepo).First(&c).Error
	return &c, err
}

func (d *ReviewRepoConfigDAO) FindByID(id uint) (*po.ReviewRepoConfig, error) {
	var c po.ReviewRepoConfig
	err := DB.First(&c, id).Error
	return &c, err
}

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

func (d *ReviewRepoConfigDAO) Delete(id uint) error {
	return DB.Delete(&po.ReviewRepoConfig{}, id).Error
}

func (d *ReviewRepoConfigDAO) FindByProviderConfigID(providerConfigID uint) ([]po.ReviewRepoConfig, error) {
	var configs []po.ReviewRepoConfig
	err := DB.Where("provider_config_id = ?", providerConfigID).Find(&configs).Error
	return configs, err
}
