package query

import (
	"github.com/yi-nology/git-manage-service/biz/dal"
	"github.com/yi-nology/git-manage-service/biz/model"
)

type RepoDAO struct{}

func NewRepoDAO() *RepoDAO {
	return &RepoDAO{}
}

func (d *RepoDAO) Create(repo *model.Repo) error {
	return dal.DB.Create(repo).Error
}

func (d *RepoDAO) FindAll() ([]model.Repo, error) {
	var repos []model.Repo
	err := dal.DB.Find(&repos).Error
	return repos, err
}

func (d *RepoDAO) FindByKey(key string) (*model.Repo, error) {
	var repo model.Repo
	err := dal.DB.Where("key = ?", key).First(&repo).Error
	return &repo, err
}

func (d *RepoDAO) Save(repo *model.Repo) error {
	return dal.DB.Save(repo).Error
}

func (d *RepoDAO) Delete(repo *model.Repo) error {
	return dal.DB.Delete(repo).Error
}
