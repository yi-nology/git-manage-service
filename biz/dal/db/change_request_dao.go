package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type ChangeRequestDAO struct{ BaseDAO[po.ChangeRequest] }

func NewChangeRequestDAO() *ChangeRequestDAO { return &ChangeRequestDAO{} }

// FindByRepoAndNumber 根据仓库 ID 和 MR/PR 编号查询
func (d *ChangeRequestDAO) FindByRepoAndNumber(repoID uint, crNumber int) (*po.ChangeRequest, error) {
	var cr po.ChangeRequest
	return &cr, DB.Where("repo_id = ? AND cr_number = ?", repoID, crNumber).First(&cr).Error
}

// FindAllByRepo 返回仓库的所有 CR（批量同步用）
func (d *ChangeRequestDAO) FindAllByRepo(repoID uint) ([]po.ChangeRequest, error) {
	var crs []po.ChangeRequest
	return crs, DB.Where("repo_id = ?", repoID).Find(&crs).Error
}

// BatchCreate 覆盖基类：空切片时跳过
func (d *ChangeRequestDAO) BatchCreate(crs []po.ChangeRequest) error {
	if len(crs) == 0 {
		return nil
	}
	return DB.Create(&crs).Error
}

// FindByRepo 分页带筛选查询
func (d *ChangeRequestDAO) FindByRepo(repoID uint, state, sourceBranch, targetBranch string, page, pageSize int) ([]po.ChangeRequest, int64, error) {
	q := DB.Model(new(po.ChangeRequest)).Where("repo_id = ?", repoID)
	if state != "" {
		q = q.Where("state = ?", state)
	}
	if sourceBranch != "" {
		q = q.Where("source_branch = ?", sourceBranch)
	}
	if targetBranch != "" {
		q = q.Where("target_branch = ?", targetBranch)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var crs []po.ChangeRequest
	err := q.Order("updated_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&crs).Error
	return crs, total, err
}

// DeleteByRepo 删除仓库的所有 CR
func (d *ChangeRequestDAO) DeleteByRepo(repoID uint) error {
	return DB.Where("repo_id = ?", repoID).Delete(new(po.ChangeRequest)).Error
}
