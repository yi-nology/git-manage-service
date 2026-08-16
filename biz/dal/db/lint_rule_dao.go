package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type LintRuleDAO struct{ BaseDAO[po.LintRule] }

func NewLintRuleDAO() *LintRuleDAO { return &LintRuleDAO{} }

// FindAll 覆盖基类：按优先级排序
func (d *LintRuleDAO) FindAll() ([]po.LintRule, error) {
	var rules []po.LintRule
	return rules, DB.Order("priority ASC, id ASC").Find(&rules).Error
}

// FindEnabled 查询启用的规则
func (d *LintRuleDAO) FindEnabled() ([]po.LintRule, error) {
	var rules []po.LintRule
	return rules, DB.Where("enabled = ?", true).Order("priority ASC, id ASC").Find(&rules).Error
}

// FindByID 覆盖基类：string 主键
func (d *LintRuleDAO) FindByID(id string) (*po.LintRule, error) {
	var rule po.LintRule
	return &rule, DB.Where("id = ?", id).First(&rule).Error
}

// FindByCategory 按分类查询
func (d *LintRuleDAO) FindByCategory(category string) ([]po.LintRule, error) {
	var rules []po.LintRule
	return rules, DB.Where("category = ?", category).Order("priority ASC, id ASC").Find(&rules).Error
}

// FindByIDs 按 ID 列表查询
func (d *LintRuleDAO) FindByIDs(ids []string) ([]po.LintRule, error) {
	var rules []po.LintRule
	return rules, DB.Where("id IN ?", ids).Order("priority ASC, id ASC").Find(&rules).Error
}

// Delete 覆盖基类：string 主键
func (d *LintRuleDAO) Delete(id string) error {
	return DB.Where("id = ?", id).Delete(new(po.LintRule)).Error
}

// ExistsByID 按 string ID 检查是否存在
func (d *LintRuleDAO) ExistsByID(id string) (bool, error) {
	return d.ExistsByField("id", id)
}
