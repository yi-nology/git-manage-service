package po

import (
	"gorm.io/gorm"
)

type AuthorIdentity struct {
	gorm.Model
	CanonicalName  string `gorm:"size:100;not null" json:"canonicalName"`
	CanonicalEmail string `gorm:"size:200;not null" json:"canonicalEmail"`
	AliasesJSON    string `gorm:"type:text" json:"aliasesJson"`
	IsDefault      bool   `gorm:"default:false;index" json:"isDefault"`
}

func (AuthorIdentity) TableName() string {
	return "author_identities"
}
