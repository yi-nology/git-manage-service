package query

import (
	"github.com/yi-nology/git-manage-service/biz/dal"
	"github.com/yi-nology/git-manage-service/biz/model"
)

type SyncRunDAO struct{}

func NewSyncRunDAO() *SyncRunDAO {
	return &SyncRunDAO{}
}

func (d *SyncRunDAO) Create(run *model.SyncRun) error {
	return dal.DB.Create(run).Error
}

func (d *SyncRunDAO) Save(run *model.SyncRun) error {
	return dal.DB.Save(run).Error
}

func (d *SyncRunDAO) FindLatest(limit int) ([]model.SyncRun, error) {
	var runs []model.SyncRun
	err := dal.DB.Order("start_time desc").Limit(limit).Preload("Task").Find(&runs).Error
	return runs, err
}

func (d *SyncRunDAO) FindByTaskKeys(taskKeys []string, limit int) ([]model.SyncRun, error) {
	var runs []model.SyncRun
	if len(taskKeys) == 0 {
		return []model.SyncRun{}, nil
	}
	err := dal.DB.Where("task_key IN ?", taskKeys).
		Order("start_time desc").Limit(limit).Preload("Task").Find(&runs).Error
	return runs, err
}

func (d *SyncRunDAO) Delete(id uint) error {
	return dal.DB.Delete(&model.SyncRun{}, id).Error
}
