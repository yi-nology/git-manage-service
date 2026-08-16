package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"gorm.io/gorm"
)

type ReviewRuleDAO struct{ BaseDAO[po.ReviewRule] }

func NewReviewRuleDAO() *ReviewRuleDAO { return &ReviewRuleDAO{} }

// FindAll 覆盖基类：按排序和创建时间正序
func (d *ReviewRuleDAO) FindAll() ([]po.ReviewRule, error) {
	var rules []po.ReviewRule
	return rules, DB.Order("sort_order ASC, created_at ASC").Find(&rules).Error
}

// FindEnabledPromptRules 查询启用的 prompt 类型规则
func (d *ReviewRuleDAO) FindEnabledPromptRules() ([]po.ReviewRule, error) {
	var rules []po.ReviewRule
	return rules, DB.Where("enabled = ? AND rule_type = ? AND prompt_text != ''", true, "prompt").
		Order("sort_order ASC, created_at ASC").Find(&rules).Error
}

// FindEnabledIDs 返回所有启用规则的 ID 集合
func (d *ReviewRuleDAO) FindEnabledIDs() (map[string]bool, error) {
	var rules []po.ReviewRule
	if err := DB.Where("enabled = ?", true).Select("id").Find(&rules).Error; err != nil {
		return nil, err
	}
	m := make(map[string]bool, len(rules))
	for _, r := range rules {
		m[r.ID] = true
	}
	return m, nil
}

// FindByID 覆盖基类：string 主键
func (d *ReviewRuleDAO) FindByID(id string) (*po.ReviewRule, error) {
	var rule po.ReviewRule
	return &rule, DB.Where("id = ?", id).First(&rule).Error
}

// Delete 覆盖基类：string 主键
func (d *ReviewRuleDAO) Delete(id string) error {
	return DB.Where("id = ?", id).Delete(new(po.ReviewRule)).Error
}

// BatchSave 事务批量保存（逐条 Save 保证 upsert 语义）
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
