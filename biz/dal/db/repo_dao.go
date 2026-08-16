package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"gorm.io/gorm"
)

type RepoDAO struct{ BaseDAO[po.Repo] }

func NewRepoDAO() *RepoDAO { return &RepoDAO{} }

// FindByKey 根据唯一 key 查询
func (d *RepoDAO) FindByKey(key string) (*po.Repo, error) {
	var repo po.Repo
	return &repo, DB.Where("key = ?", key).First(&repo).Error
}

// FindByPath 根据本地路径查询
func (d *RepoDAO) FindByPath(path string) (*po.Repo, error) {
	var repo po.Repo
	return &repo, DB.Where("path = ?", path).First(&repo).Error
}

// Delete 覆盖基类：参数为对象指针（非 ID）
func (d *RepoDAO) Delete(repo *po.Repo) error {
	return DB.Delete(repo).Error
}

// DeleteWithBindings 事务删除仓库并标记关联绑定为 deleted
func (d *RepoDAO) DeleteWithBindings(repo *po.Repo) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&po.RepoProviderBinding{}).Where("repo_id = ? AND status = ?", repo.ID, "active").
			Update("status", "deleted").Error; err != nil {
			return err
		}
		return tx.Delete(repo).Error
	})
}

// FindByPlatformOwnerRepo 通过平台 owner/repo slug 查找（用于 webhook 热路径）
func (d *RepoDAO) FindByPlatformOwnerRepo(owner, repo string) (*po.Repo, error) {
	var r po.Repo
	return &r, DB.Where("platform_owner = ? AND platform_repo = ?", owner, repo).First(&r).Error
}
