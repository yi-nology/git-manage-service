package adapter

import (
	"context"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

// SyncTaskDAO Bridge: 同步任务 DAO 桥接
// 用于在 git-sync-service 中访问数据库时进行模型转换
type SyncTaskDAO struct {
	nativeDAO *db.SyncTaskDAO
}

func NewSyncTaskDAO() *SyncTaskDAO {
	return &SyncTaskDAO{
		nativeDAO: db.NewSyncTaskDAO(),
	}
}

func (d *SyncTaskDAO) FindByKey(ctx context.Context, key string) (*GitSyncTask, error) {
	task, err := d.nativeDAO.FindByKey(key)
	if err != nil {
		return nil, err
	}
	return ToGitSyncTask(task), nil
}

// SyncRunDAO Bridge: 同步运行记录 DAO 桥接
type SyncRunDAO struct {
	nativeDAO *db.SyncRunDAO
}

func NewSyncRunDAO() *SyncRunDAO {
	return &SyncRunDAO{
		nativeDAO: db.NewSyncRunDAO(),
	}
}

func (d *SyncRunDAO) Create(ctx context.Context, run *GitSyncRun) error {
	nativeRun := FromGitSyncRun(run)
	return d.nativeDAO.Create(nativeRun)
}

func (d *SyncRunDAO) Save(ctx context.Context, run *GitSyncRun) error {
	nativeRun := FromGitSyncRun(run)
	return d.nativeDAO.Save(nativeRun)
}

// RepoDAO Bridge: 仓库 DAO 桥接
type RepoDAO struct {
	nativeDAO *db.RepoDAO
}

func NewRepoDAO() *RepoDAO {
	return &RepoDAO{
		nativeDAO: db.NewRepoDAO(),
	}
}

func (d *RepoDAO) FindByKey(ctx context.Context, key string) (*GitSyncRepo, error) {
	repo, err := d.nativeDAO.FindByKey(key)
	if err != nil {
		return nil, err
	}
	return ToGitSyncRepo(repo), nil
}

func (d *RepoDAO) Create(ctx context.Context, repo *GitSyncRepo) error {
	nativeRepo := &po.Repo{
		Key:        repo.Key,
		Name:       repo.Name,
		RemoteURL:  repo.CloneURL,
		Path:       repo.Path,
		AuthType:   "token",
		AuthSecret: repo.AccessToken,
	}
	return d.nativeDAO.Create(nativeRepo)
}
