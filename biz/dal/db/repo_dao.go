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

// DeleteWithBindings 事务删除仓库并级联清理：绑定点位 deleted、软删该仓库的
// 镜像、硬删镜像同步日志。不级联的话 scheduler/队列会对已删仓库继续同步。
func (d *RepoDAO) DeleteWithBindings(repo *po.Repo) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&po.RepoProviderBinding{}).Where("repo_id = ? AND status = ?", repo.ID, "active").
			Update("status", "deleted").Error; err != nil {
			return err
		}

		// 先收集镜像 ID：镜像软删后按 repo_id 就查不到了。
		var mirrorIDs []uint
		if err := tx.Model(new(po.Mirror)).Where("repo_id = ?", repo.ID).Pluck("id", &mirrorIDs).Error; err != nil {
			return err
		}
		if len(mirrorIDs) > 0 {
			if err := tx.Where("mirror_id IN ?", mirrorIDs).Delete(new(po.MirrorSyncLog)).Error; err != nil {
				return err
			}
		}
		if err := tx.Where("repo_id = ?", repo.ID).Delete(new(po.Mirror)).Error; err != nil {
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
