package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type RepoProviderBindingDAO struct {
	BaseDAO[po.RepoProviderBinding]
}

func NewRepoProviderBindingDAO() *RepoProviderBindingDAO { return &RepoProviderBindingDAO{} }

// FindByID 覆盖基类：带 Preload
func (d *RepoProviderBindingDAO) FindByID(id uint) (*po.RepoProviderBinding, error) {
	var b po.RepoProviderBinding
	return &b, DB.Preload("ProviderConfig").Preload("Repo").First(&b, id).Error
}

// FindAll 覆盖基类：仅活跃绑定 + Preload
func (d *RepoProviderBindingDAO) FindAll() ([]po.RepoProviderBinding, error) {
	var bindings []po.RepoProviderBinding
	return bindings, DB.Where("status = ?", "active").Preload("ProviderConfig").Preload("Repo").Find(&bindings).Error
}

// FindByRepoID 查询仓库的活跃绑定
func (d *RepoProviderBindingDAO) FindByRepoID(repoID uint) ([]po.RepoProviderBinding, error) {
	var bindings []po.RepoProviderBinding
	return bindings, DB.Where("repo_id = ? AND status = ?", repoID, "active").Find(&bindings).Error
}

// FindByRepoIDWithProvider 查询仓库的活跃绑定（带 Preload）
func (d *RepoProviderBindingDAO) FindByRepoIDWithProvider(repoID uint) ([]po.RepoProviderBinding, error) {
	var bindings []po.RepoProviderBinding
	return bindings, DB.Where("repo_id = ? AND status = ?", repoID, "active").
		Preload("ProviderConfig").Preload("Repo").Find(&bindings).Error
}

// FindPrimaryByRepoID 查询仓库的主绑定
func (d *RepoProviderBindingDAO) FindPrimaryByRepoID(repoID uint) (*po.RepoProviderBinding, error) {
	var b po.RepoProviderBinding
	return &b, DB.Where("repo_id = ? AND is_primary = ? AND status = ?", repoID, true, "active").
		Preload("ProviderConfig").Preload("Repo").First(&b).Error
}

// FindByRepoAndProvider 按仓库+provider 查询
func (d *RepoProviderBindingDAO) FindByRepoAndProvider(repoID, providerConfigID uint) (*po.RepoProviderBinding, error) {
	var b po.RepoProviderBinding
	return &b, DB.Where("repo_id = ? AND provider_config_id = ? AND status = ?", repoID, providerConfigID, "active").First(&b).Error
}

// FindByPlatformRepo 按平台 owner/repo 查询
func (d *RepoProviderBindingDAO) FindByPlatformRepo(providerConfigID uint, platformOwner, platformRepo string) (*po.RepoProviderBinding, error) {
	var b po.RepoProviderBinding
	return &b, DB.Where("provider_config_id = ? AND platform_owner = ? AND platform_repo = ? AND status = ?",
		providerConfigID, platformOwner, platformRepo, "active").First(&b).Error
}

// FindByProviderConfigID 查询 provider 的活跃绑定
func (d *RepoProviderBindingDAO) FindByProviderConfigID(providerConfigID uint) ([]po.RepoProviderBinding, error) {
	var bindings []po.RepoProviderBinding
	return bindings, DB.Where("provider_config_id = ? AND status = ?", providerConfigID, "active").
		Preload("ProviderConfig").Preload("Repo").Find(&bindings).Error
}

// SoftDelete 软删除（标记 status=deleted）
func (d *RepoProviderBindingDAO) SoftDelete(id uint) error {
	return DB.Model(new(po.RepoProviderBinding)).Where("id = ?", id).Update("status", "deleted").Error
}

// ClearPrimaryByRepoID 清除仓库的主绑定标记
func (d *RepoProviderBindingDAO) ClearPrimaryByRepoID(repoID uint) error {
	return DB.Model(new(po.RepoProviderBinding)).
		Where("repo_id = ? AND is_primary = ? AND status = ?", repoID, true, "active").
		Update("is_primary", false).Error
}

// ExistsByRepoAndProvider 检查绑定是否存在（复合条件，不能用基类简化）
func (d *RepoProviderBindingDAO) ExistsByRepoAndProvider(repoID, providerConfigID uint) (bool, error) {
	var count int64
	err := DB.Model(new(po.RepoProviderBinding)).
		Where("repo_id = ? AND provider_config_id = ? AND status = ?", repoID, providerConfigID, "active").
		Count(&count).Error
	return count > 0, err
}

// ExistsByPlatformRepo 按平台 owner/repo 检查是否存在
func (d *RepoProviderBindingDAO) ExistsByPlatformRepo(providerConfigID uint, platformOwner, platformRepo string) (bool, error) {
	var count int64
	err := DB.Model(new(po.RepoProviderBinding)).
		Where("provider_config_id = ? AND platform_owner = ? AND platform_repo = ? AND status = ?",
			providerConfigID, platformOwner, platformRepo, "active").Count(&count).Error
	return count > 0, err
}

// FindByRepoKey 通过仓库 key 查询绑定
func (d *RepoProviderBindingDAO) FindByRepoKey(repoKey string) ([]po.RepoProviderBinding, error) {
	var repo po.Repo
	if err := DB.Where("key = ?", repoKey).First(&repo).Error; err != nil {
		return nil, err
	}
	return d.FindByRepoIDWithProvider(repo.ID)
}
