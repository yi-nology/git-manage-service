package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"gorm.io/gorm"
)

type RepoDAO struct{}

func NewRepoDAO() *RepoDAO {
	return &RepoDAO{}
}

func (d *RepoDAO) Create(repo *po.Repo) error {
	return DB.Create(repo).Error
}

func (d *RepoDAO) FindAll() ([]po.Repo, error) {
	var repos []po.Repo
	err := DB.Find(&repos).Error
	return repos, err
}

func (d *RepoDAO) FindByKey(key string) (*po.Repo, error) {
	var repo po.Repo
	err := DB.Where("key = ?", key).First(&repo).Error
	return &repo, err
}

func (d *RepoDAO) FindByPath(path string) (*po.Repo, error) {
	var repo po.Repo
	err := DB.Where("path = ?", path).First(&repo).Error
	return &repo, err
}

func (d *RepoDAO) Save(repo *po.Repo) error {
	return DB.Save(repo).Error
}

func (d *RepoDAO) Delete(repo *po.Repo) error {
	return DB.Delete(repo).Error
}

func (d *RepoDAO) DeleteWithBindings(repo *po.Repo) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&po.RepoProviderBinding{}).Where("repo_id = ? AND status = ?", repo.ID, "active").
			Update("status", "deleted").Error; err != nil {
			return err
		}
		return tx.Delete(repo).Error
	})
}

// FindByID 根据ID查询仓库
func (d *RepoDAO) FindByID(id uint) (*po.Repo, error) {
	var repo po.Repo
	err := DB.First(&repo, id).Error
	return &repo, err
}

// FindByPlatformOwnerRepo looks up a repo by its platform owner/repo slug via
// the (platform_owner, platform_repo) index. Use this instead of FindAll() +
// in-memory matching on hot paths like incoming webhooks.
func (d *RepoDAO) FindByPlatformOwnerRepo(owner, repo string) (*po.Repo, error) {
	var r po.Repo
	err := DB.Where("platform_owner = ? AND platform_repo = ?", owner, repo).First(&r).Error
	return &r, err
}
