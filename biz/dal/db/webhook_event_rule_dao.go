package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type WebhookEventRuleDAO struct{}

func NewWebhookEventRuleDAO() *WebhookEventRuleDAO { return &WebhookEventRuleDAO{} }

func (d *WebhookEventRuleDAO) Create(rule *po.WebhookEventRule) error {
	return DB.Create(rule).Error
}

func (d *WebhookEventRuleDAO) FindAll() ([]po.WebhookEventRule, error) {
	var rules []po.WebhookEventRule
	err := DB.Order("priority DESC, created_at ASC").Find(&rules).Error
	return rules, err
}

func (d *WebhookEventRuleDAO) FindActive() ([]po.WebhookEventRule, error) {
	var rules []po.WebhookEventRule
	err := DB.Where("is_active = ?", true).Order("priority DESC, created_at ASC").Find(&rules).Error
	return rules, err
}

func (d *WebhookEventRuleDAO) FindByID(id uint) (*po.WebhookEventRule, error) {
	var rule po.WebhookEventRule
	err := DB.First(&rule, id).Error
	return &rule, err
}

func (d *WebhookEventRuleDAO) Save(rule *po.WebhookEventRule) error {
	return DB.Save(rule).Error
}

func (d *WebhookEventRuleDAO) Delete(rule *po.WebhookEventRule) error {
	return DB.Delete(rule).Error
}
