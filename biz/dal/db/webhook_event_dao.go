package db

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type WebhookEventDAO struct{ BaseDAO[po.WebhookEvent] }

func NewWebhookEventDAO() *WebhookEventDAO { return &WebhookEventDAO{} }

// FindByEventID 根据事件 ID 查询
func (d *WebhookEventDAO) FindByEventID(eventID string) (*po.WebhookEvent, error) {
	var event po.WebhookEvent
	return &event, DB.Where("event_id = ?", eventID).First(&event).Error
}

// List 分页带筛选查询
func (d *WebhookEventDAO) List(eventType, source, status string, page, pageSize int) ([]po.WebhookEvent, int64, error) {
	q := DB.Model(new(po.WebhookEvent))
	if eventType != "" {
		q = q.Where("event_type = ?", eventType)
	}
	if source != "" {
		q = q.Where("source = ?", source)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var events []po.WebhookEvent
	return events, total, q.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&events).Error
}

// WebhookRuleDAO webhook 规则
type WebhookRuleDAO struct{ BaseDAO[po.WebhookRule] }

func NewWebhookRuleDAO() *WebhookRuleDAO { return &WebhookRuleDAO{} }

// FindAll 覆盖基类：仅返回启用的规则
func (d *WebhookRuleDAO) FindAll() ([]po.WebhookRule, error) {
	var rules []po.WebhookRule
	return rules, DB.Where("enabled = ?", true).Find(&rules).Error
}

// FindByProviderConfigID 查询 provider 关联的启用规则（含全局规则）
func (d *WebhookRuleDAO) FindByProviderConfigID(providerConfigID uint) ([]po.WebhookRule, error) {
	var rules []po.WebhookRule
	return rules, DB.Where("provider_config_id = ? OR provider_config_id = 0", providerConfigID).
		Where("enabled = ?", true).Find(&rules).Error
}
