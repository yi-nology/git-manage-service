package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"gorm.io/gorm"
)

type AuthorIdentityDAO struct{ BaseDAO[po.AuthorIdentity] }

func NewAuthorIdentityDAO() *AuthorIdentityDAO { return &AuthorIdentityDAO{} }

// Update 更新身份（与 Save 相同，保持向后兼容）
func (d *AuthorIdentityDAO) Update(identity *po.AuthorIdentity) error {
	return DB.Save(identity).Error
}

// ListAll 按默认优先、创建时间正序列出
func (d *AuthorIdentityDAO) ListAll() ([]po.AuthorIdentity, error) {
	var identities []po.AuthorIdentity
	return identities, DB.Order("is_default DESC, created_at ASC").Find(&identities).Error
}

// GetDefault 查询默认身份
func (d *AuthorIdentityDAO) GetDefault() (*po.AuthorIdentity, error) {
	var identity po.AuthorIdentity
	return &identity, DB.Where("is_default = ?", true).First(&identity).Error
}

// ClearAllDefaults 取消所有默认
func (d *AuthorIdentityDAO) ClearAllDefaults() error {
	return DB.Model(new(po.AuthorIdentity)).Where("is_default = ?", true).Update("is_default", false).Error
}

// SetDefault 设为默认（事务）
func (d *AuthorIdentityDAO) SetDefault(id uint) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(new(po.AuthorIdentity)).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(new(po.AuthorIdentity)).Where("id = ?", id).Update("is_default", true).Error
	})
}
