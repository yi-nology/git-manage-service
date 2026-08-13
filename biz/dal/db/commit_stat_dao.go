package db

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
	"gorm.io/gorm/clause"
)

type CommitStatDAO struct{}

func NewCommitStatDAO() *CommitStatDAO {
	return &CommitStatDAO{}
}

// FindLatestCommitTime returns the latest commit time for a repo
func (dao *CommitStatDAO) FindLatestCommitTime(repoID uint) (time.Time, error) {
	var stat po.CommitStat
	err := DB.Where("repo_id = ?", repoID).Order("commit_time desc").First(&stat).Error
	if err != nil {
		return time.Time{}, nil // Return zero time if not found (start from beginning)
	}
	return stat.CommitTime, nil
}

// BatchSave inserts or updates commit stats
func (dao *CommitStatDAO) BatchSave(stats []*po.CommitStat) error {
	if len(stats) == 0 {
		return nil
	}
	// Upsert on conflict
	return DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "repo_id"}, {Name: "commit_hash"}},
		DoUpdates: clause.AssignmentColumns([]string{"additions", "deletions", "author_name", "author_email", "commit_time"}),
	}).CreateInBatches(stats, 100).Error
}
