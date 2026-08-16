package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type BranchRuleSetDAO struct{ BaseDAO[po.BranchRuleSet] }

func NewBranchRuleSetDAO() *BranchRuleSetDAO { return &BranchRuleSetDAO{} }

// FindGlobal 查询全局规则集
func (d *BranchRuleSetDAO) FindGlobal() (*po.BranchRuleSet, error) {
	var s po.BranchRuleSet
	return &s, DB.Where("scope_type = ? AND scope_id = ?", "global", "default").First(&s).Error
}

// Upsert 存在则更新，否则创建
func (d *BranchRuleSetDAO) Upsert(s *po.BranchRuleSet) error {
	var existing po.BranchRuleSet
	err := DB.Where("scope_type = ? AND scope_id = ?", s.ScopeType, s.ScopeID).First(&existing).Error
	if err != nil {
		return DB.Create(s).Error
	}
	s.ID = existing.ID
	s.CreatedAt = existing.CreatedAt
	return DB.Save(s).Error
}

// BranchRuleOverrideDAO 分支规则覆盖
type BranchRuleOverrideDAO struct{ BaseDAO[po.BranchRuleOverride] }

func NewBranchRuleOverrideDAO() *BranchRuleOverrideDAO { return &BranchRuleOverrideDAO{} }

// FindByRemoteRepo 按平台配置和 owner/repo 查询
func (d *BranchRuleOverrideDAO) FindByRemoteRepo(providerConfigID uint, platformOwner, platformRepo string) (*po.BranchRuleOverride, error) {
	var o po.BranchRuleOverride
	return &o, DB.Where("provider_config_id = ? AND platform_owner = ? AND platform_repo = ?",
		providerConfigID, platformOwner, platformRepo).First(&o).Error
}

// Upsert 存在则更新，否则创建
func (d *BranchRuleOverrideDAO) Upsert(o *po.BranchRuleOverride) error {
	var existing po.BranchRuleOverride
	err := DB.Where("provider_config_id = ? AND platform_owner = ? AND platform_repo = ?",
		o.ProviderConfigID, o.PlatformOwner, o.PlatformRepo).First(&existing).Error
	if err != nil {
		return DB.Create(o).Error
	}
	o.ID = existing.ID
	o.CreatedAt = existing.CreatedAt
	return DB.Save(o).Error
}
