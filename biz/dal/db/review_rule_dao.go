package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"gorm.io/gorm"
)

type ReviewRuleDAO struct{}

func NewReviewRuleDAO() *ReviewRuleDAO { return &ReviewRuleDAO{} }

func (d *ReviewRuleDAO) FindAll() ([]po.ReviewRule, error) {
	var rules []po.ReviewRule
	err := DB.Order("sort_order ASC, created_at ASC").Find(&rules).Error
	return rules, err
}

func (d *ReviewRuleDAO) FindEnabledPromptRules() ([]po.ReviewRule, error) {
	var rules []po.ReviewRule
	err := DB.Where("enabled = ? AND rule_type = ? AND prompt_text != ''", true, "prompt", true).
		Order("sort_order ASC, created_at ASC").
		Find(&rules).Error
	return rules, err
}

func (d *ReviewRuleDAO) FindEnabledIDs() (map[string]bool, error) {
	var rules []po.ReviewRule
	err := DB.Where("enabled = ?", true).Select("id").Find(&rules).Error
	if err != nil {
		return nil, err
	}
	m := make(map[string]bool, len(rules))
	for _, r := range rules {
		m[r.ID] = true
	}
	return m, nil
}

func (d *ReviewRuleDAO) FindByID(id string) (*po.ReviewRule, error) {
	var rule po.ReviewRule
	err := DB.Where("id = ?", id).First(&rule).Error
	return &rule, err
}

func (d *ReviewRuleDAO) Create(r *po.ReviewRule) error {
	return DB.Create(r).Error
}

func (d *ReviewRuleDAO) Save(r *po.ReviewRule) error {
	return DB.Save(r).Error
}

func (d *ReviewRuleDAO) Delete(id string) error {
	return DB.Where("id = ?", id).Delete(&po.ReviewRule{}).Error
}

func (d *ReviewRuleDAO) Count() (int64, error) {
	var count int64
	err := DB.Model(&po.ReviewRule{}).Count(&count).Error
	return count, err
}

func (d *ReviewRuleDAO) BatchSave(rules []po.ReviewRule) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		for i := range rules {
			if err := tx.Save(&rules[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
