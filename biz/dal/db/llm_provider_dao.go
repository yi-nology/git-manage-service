package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"gorm.io/gorm"
)

type LLMProviderDAO struct{}

func NewLLMProviderDAO() *LLMProviderDAO { return &LLMProviderDAO{} }

func (d *LLMProviderDAO) Create(p *po.LLMProvider) error {
	return DB.Create(p).Error
}

func (d *LLMProviderDAO) FindByID(id uint) (*po.LLMProvider, error) {
	var p po.LLMProvider
	err := DB.First(&p, id).Error
	return &p, err
}

func (d *LLMProviderDAO) FindAll() ([]po.LLMProvider, error) {
	var providers []po.LLMProvider
	err := DB.Order("created_at ASC").Find(&providers).Error
	return providers, err
}

func (d *LLMProviderDAO) FindByName(name string) (*po.LLMProvider, error) {
	var p po.LLMProvider
	err := DB.Where("name = ?", name).First(&p).Error
	return &p, err
}

func (d *LLMProviderDAO) FindDefault() (*po.LLMProvider, error) {
	var p po.LLMProvider
	err := DB.Where("is_default = ?", true).First(&p).Error
	return &p, err
}

func (d *LLMProviderDAO) Save(p *po.LLMProvider) error {
	return DB.Save(p).Error
}

func (d *LLMProviderDAO) Delete(id uint) error {
	return DB.Delete(&po.LLMProvider{}, id).Error
}

func (d *LLMProviderDAO) ClearAllDefault() error {
	return DB.Model(&po.LLMProvider{}).Where("is_default = ?", true).Update("is_default", false).Error
}

func (d *LLMProviderDAO) SetDefault(id uint) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&po.LLMProvider{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(&po.LLMProvider{}).Where("id = ?", id).Update("is_default", true).Error
	})
}

func (d *LLMProviderDAO) ExistsByName(name string) (bool, error) {
	var count int64
	err := DB.Model(&po.LLMProvider{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

func (d *LLMProviderDAO) ExistsByNameExcludeID(name string, excludeID uint) (bool, error) {
	var count int64
	err := DB.Model(&po.LLMProvider{}).Where("name = ? AND id != ?", name, excludeID).Count(&count).Error
	return count > 0, err
}
