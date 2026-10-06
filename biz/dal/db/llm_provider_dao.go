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

// UpsertWithDefault 在单个事务内保存 provider 并维护「存在 provider 时
// 有且仅有一个默认」：p 为默认则清掉其余；保存后全库没有默认（用户取消
// 默认、或这是首个 provider）则提升 p。p.IsDefault 回写为最终值。
func (d *LLMProviderDAO) UpsertWithDefault(p *po.LLMProvider) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(p).Error; err != nil {
			return err
		}
		if p.IsDefault {
			return tx.Model(new(po.LLMProvider)).
				Where("is_default = ? AND id <> ?", true, p.ID).
				Update("is_default", false).Error
		}
		var defaults int64
		if err := tx.Model(new(po.LLMProvider)).Where("is_default = ?", true).Count(&defaults).Error; err != nil {
			return err
		}
		if defaults == 0 {
			if err := tx.Model(new(po.LLMProvider)).Where("id = ?", p.ID).
				Update("is_default", true).Error; err != nil {
				return err
			}
			p.IsDefault = true
		}
		return nil
	})
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
