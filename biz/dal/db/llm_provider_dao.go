package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"gorm.io/gorm"
)

type LLMProviderDAO struct{ BaseDAO[po.LLMProvider] }

func NewLLMProviderDAO() *LLMProviderDAO { return &LLMProviderDAO{} }

// FindAll 覆盖基类：按创建时间正序
func (d *LLMProviderDAO) FindAll() ([]po.LLMProvider, error) {
	var providers []po.LLMProvider
	return providers, DB.Order("created_at ASC").Find(&providers).Error
}

// FindByName 根据名称查询
func (d *LLMProviderDAO) FindByName(name string) (*po.LLMProvider, error) {
	var p po.LLMProvider
	return &p, DB.Where("name = ?", name).First(&p).Error
}

// FindByNameUnscoped 包含软删除记录
func (d *LLMProviderDAO) FindByNameUnscoped(name string) (*po.LLMProvider, error) {
	var p po.LLMProvider
	return &p, DB.Unscoped().Where("name = ?", name).First(&p).Error
}

// FindDefault 查询默认 Provider
func (d *LLMProviderDAO) FindDefault() (*po.LLMProvider, error) {
	var p po.LLMProvider
	return &p, DB.Where("is_default = ?", true).First(&p).Error
}

// ClearAllDefault 取消所有默认设置
func (d *LLMProviderDAO) ClearAllDefault() error {
	return DB.Model(new(po.LLMProvider)).Where("is_default = ?", true).Update("is_default", false).Error
}

// SetDefault 设为默认（事务：先清旧再设新）
func (d *LLMProviderDAO) SetDefault(id uint) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(new(po.LLMProvider)).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(new(po.LLMProvider)).Where("id = ?", id).Update("is_default", true).Error
	})
}

// ExistsByName 覆盖基类：软删除场景需加 deleted_at IS NULL
func (d *LLMProviderDAO) ExistsByName(name string) (bool, error) {
	var count int64
	err := DB.Model(new(po.LLMProvider)).Where("name = ? AND deleted_at IS NULL", name).Count(&count).Error
	return count > 0, err
}

// FindEmbeddingProvider 查询嵌入向量 Provider
func (d *LLMProviderDAO) FindEmbeddingProvider() (*po.LLMProvider, error) {
	var p po.LLMProvider
	return &p, DB.Where("is_embedding = ?", true).Order("updated_at DESC").First(&p).Error
}
