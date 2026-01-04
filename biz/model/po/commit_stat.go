package po

import (
	"time"

	"gorm.io/gorm"
)

type CommitStat struct {
	gorm.Model
	RepoID      uint      `gorm:"index:idx_repo_hash,unique;not null" json:"repo_id"`
	CommitHash  string    `gorm:"type:varchar(64);index:idx_repo_hash,unique;not null" json:"commit_hash"`
	AuthorName  string    `gorm:"type:varchar(255)" json:"author_name"`
	AuthorEmail string    `gorm:"type:varchar(255);index" json:"author_email"`
	CommitTime  time.Time `gorm:"index" json:"commit_time"`
	Additions   int       `json:"additions"`
	Deletions   int       `json:"deletions"`
}

func (c *CommitStat) TableName() string {
	return "commit_stats"
}
