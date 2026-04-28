package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type RepoProviderBindingDAO struct{}

func NewRepoProviderBindingDAO() *RepoProviderBindingDAO {
	return &RepoProviderBindingDAO{}
}

func (d *RepoProviderBindingDAO) Create(b *po.RepoProviderBinding) error {
	return DB.Create(b).Error
}

func (d *RepoProviderBindingDAO) FindByID(id uint) (*po.RepoProviderBinding, error) {
	var b po.RepoProviderBinding
	err := DB.Preload("ProviderConfig").Preload("Repo").First(&b, id).Error
	return &b, err
}

func (d *RepoProviderBindingDAO) FindByRepoID(repoID uint) ([]po.RepoProviderBinding, error) {
	var bindings []po.RepoProviderBinding
	err := DB.Where("repo_id = ? AND status = ?", repoID, "active").Find(&bindings).Error
	return bindings, err
}

func (d *RepoProviderBindingDAO) FindByRepoIDWithProvider(repoID uint) ([]po.RepoProviderBinding, error) {
	var bindings []po.RepoProviderBinding
	err := DB.Where("repo_id = ? AND status = ?", repoID, "active").
		Preload("ProviderConfig").Preload("Repo").
		Find(&bindings).Error
	return bindings, err
}

func (d *RepoProviderBindingDAO) FindPrimaryByRepoID(repoID uint) (*po.RepoProviderBinding, error) {
	var b po.RepoProviderBinding
	err := DB.Where("repo_id = ? AND is_primary = ? AND status = ?", repoID, true, "active").
		Preload("ProviderConfig").Preload("Repo").
		First(&b).Error
	return &b, err
}

func (d *RepoProviderBindingDAO) FindByRepoAndProvider(repoID, providerConfigID uint) (*po.RepoProviderBinding, error) {
	var b po.RepoProviderBinding
	err := DB.Where("repo_id = ? AND provider_config_id = ? AND status = ?", repoID, providerConfigID, "active").First(&b).Error
	return &b, err
}

func (d *RepoProviderBindingDAO) FindByPlatformRepo(providerConfigID uint, platformOwner, platformRepo string) (*po.RepoProviderBinding, error) {
	var b po.RepoProviderBinding
	err := DB.Where("provider_config_id = ? AND platform_owner = ? AND platform_repo = ? AND status = ?",
		providerConfigID, platformOwner, platformRepo, "active").First(&b).Error
	return &b, err
}

func (d *RepoProviderBindingDAO) FindAll() ([]po.RepoProviderBinding, error) {
	var bindings []po.RepoProviderBinding
	err := DB.Where("status = ?", "active").Preload("ProviderConfig").Preload("Repo").Find(&bindings).Error
	return bindings, err
}

func (d *RepoProviderBindingDAO) FindByProviderConfigID(providerConfigID uint) ([]po.RepoProviderBinding, error) {
	var bindings []po.RepoProviderBinding
	err := DB.Where("provider_config_id = ? AND status = ?", providerConfigID, "active").
		Preload("ProviderConfig").Preload("Repo").
		Find(&bindings).Error
	return bindings, err
}

func (d *RepoProviderBindingDAO) Save(b *po.RepoProviderBinding) error {
	return DB.Save(b).Error
}

func (d *RepoProviderBindingDAO) Delete(id uint) error {
	return DB.Delete(&po.RepoProviderBinding{}, id).Error
}

func (d *RepoProviderBindingDAO) SoftDelete(id uint) error {
	return DB.Model(&po.RepoProviderBinding{}).Where("id = ?", id).Update("status", "deleted").Error
}

func (d *RepoProviderBindingDAO) DeleteByRepoID(repoID uint) error {
	return DB.Model(&po.RepoProviderBinding{}).
		Where("repo_id = ? AND status = ?", repoID, "active").
		Update("status", "deleted").Error
}

func (d *RepoProviderBindingDAO) ClearPrimaryByRepoID(repoID uint) error {
	return DB.Model(&po.RepoProviderBinding{}).
		Where("repo_id = ? AND is_primary = ? AND status = ?", repoID, true, "active").
		Update("is_primary", false).Error
}

func (d *RepoProviderBindingDAO) ExistsByRepoAndProvider(repoID, providerConfigID uint) (bool, error) {
	var count int64
	err := DB.Model(&po.RepoProviderBinding{}).
		Where("repo_id = ? AND provider_config_id = ? AND status = ?", repoID, providerConfigID, "active").
		Count(&count).Error
	return count > 0, err
}

func (d *RepoProviderBindingDAO) ExistsByPlatformRepo(providerConfigID uint, platformOwner, platformRepo string) (bool, error) {
	var count int64
	err := DB.Model(&po.RepoProviderBinding{}).
		Where("provider_config_id = ? AND platform_owner = ? AND platform_repo = ? AND status = ?",
			providerConfigID, platformOwner, platformRepo, "active").
		Count(&count).Error
	return count > 0, err
}

func (d *RepoProviderBindingDAO) FindByRepoKey(repoKey string) ([]po.RepoProviderBinding, error) {
	var repo po.Repo
	if err := DB.Where("key = ?", repoKey).First(&repo).Error; err != nil {
		return nil, err
	}
	return d.FindByRepoIDWithProvider(repo.ID)
}
