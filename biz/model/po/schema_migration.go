package po

import "time"

type SchemaMigration struct {
	Version   string    `gorm:"primaryKey;size:64" json:"version"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	AppliedAt time.Time `gorm:"not null" json:"applied_at"`
}

func (SchemaMigration) TableName() string {
	return "schema_migrations"
}
