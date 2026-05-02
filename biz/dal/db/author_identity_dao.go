package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"gorm.io/gorm"
)

type AuthorIdentityDAO struct{}

func NewAuthorIdentityDAO() *AuthorIdentityDAO {
	return &AuthorIdentityDAO{}
}

func (d *AuthorIdentityDAO) Create(identity *po.AuthorIdentity) error {
	return DB.Create(identity).Error
}

func (d *AuthorIdentityDAO) Update(identity *po.AuthorIdentity) error {
	return DB.Save(identity).Error
}

func (d *AuthorIdentityDAO) Delete(id uint) error {
	return DB.Delete(&po.AuthorIdentity{}, id).Error
}

func (d *AuthorIdentityDAO) FindByID(id uint) (*po.AuthorIdentity, error) {
	var identity po.AuthorIdentity
	err := DB.First(&identity, id).Error
	return &identity, err
}

func (d *AuthorIdentityDAO) ListAll() ([]po.AuthorIdentity, error) {
	var identities []po.AuthorIdentity
	err := DB.Order("is_default DESC, created_at ASC").Find(&identities).Error
	return identities, err
}

func (d *AuthorIdentityDAO) GetDefault() (*po.AuthorIdentity, error) {
	var identity po.AuthorIdentity
	err := DB.Where("is_default = ?", true).First(&identity).Error
	return &identity, err
}

func (d *AuthorIdentityDAO) ClearAllDefaults() error {
	return DB.Model(&po.AuthorIdentity{}).Where("is_default = ?", true).Update("is_default", false).Error
}

func (d *AuthorIdentityDAO) SetDefault(id uint) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&po.AuthorIdentity{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(&po.AuthorIdentity{}).Where("id = ?", id).Update("is_default", true).Error
	})
}
