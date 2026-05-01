package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type BranchRuleSetDAO struct{}

func NewBranchRuleSetDAO() *BranchRuleSetDAO { return &BranchRuleSetDAO{} }

func (d *BranchRuleSetDAO) FindGlobal() (*po.BranchRuleSet, error) {
	var s po.BranchRuleSet
	err := DB.Where("scope_type = ? AND scope_id = ?", "global", "default").First(&s).Error
	return &s, err
}

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

func (d *BranchRuleSetDAO) Save(s *po.BranchRuleSet) error {
	return DB.Save(s).Error
}

type BranchRuleOverrideDAO struct{}

func NewBranchRuleOverrideDAO() *BranchRuleOverrideDAO {
	return &BranchRuleOverrideDAO{}
}

func (d *BranchRuleOverrideDAO) FindByRemoteRepo(providerConfigID uint, platformOwner, platformRepo string) (*po.BranchRuleOverride, error) {
	var o po.BranchRuleOverride
	err := DB.Where("provider_config_id = ? AND platform_owner = ? AND platform_repo = ?",
		providerConfigID, platformOwner, platformRepo).First(&o).Error
	return &o, err
}

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

func (d *BranchRuleOverrideDAO) Delete(id uint) error {
	return DB.Delete(&po.BranchRuleOverride{}, id).Error
}
