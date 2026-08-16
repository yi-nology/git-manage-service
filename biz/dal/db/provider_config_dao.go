package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type ProviderConfigDAO struct{ BaseDAO[po.ProviderConfig] }

func NewProviderConfigDAO() *ProviderConfigDAO { return &ProviderConfigDAO{} }

// FindByPlatform 按平台类型查询
func (d *ProviderConfigDAO) FindByPlatform(platform string) ([]po.ProviderConfig, error) {
	var configs []po.ProviderConfig
	err := DB.Where("platform = ?", platform).Find(&configs).Error
	return configs, err
}

// FindAll 覆盖基类，按更新时间倒序
func (d *ProviderConfigDAO) FindAll() ([]po.ProviderConfig, error) {
	var configs []po.ProviderConfig
	return configs, DB.Order("updated_at DESC").Find(&configs).Error
}
