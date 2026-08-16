package webhook_event

import (
	"context"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/webhookevent"
	"github.com/yi-nology/git-manage-service/pkg/handler"
	pkgresponse "github.com/yi-nology/git-manage-service/pkg/response"
)

func List(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.ListWebhookEventsReq) (map[string]interface{}, error) {
		if req.Page == 0 {
			req.Page = 1
		}
		if req.PageSize == 0 {
			req.PageSize = 20
		}
		events, total, err := webhookevent.List(req.EventType, req.Source, req.Status, req.Page, req.PageSize)
		if err != nil {
			return nil, handler.ErrInternal("Failed to list webhook events: " + err.Error())
		}
		return map[string]interface{}{
			"items": events,
			"total": total,
		}, nil
	})
}

func Retry(ctx context.Context, c *app.RequestContext) {
	type retryReq struct {
		EventID uint `json:"event_id"`
	}
	handler.BindAndDo(c, func(req *retryReq) (map[string]string, error) {
		if req.EventID == 0 {
			return nil, handler.ErrBadRequest("event_id is required")
		}
		if err := webhookevent.Retry(req.EventID); err != nil {
			return nil, handler.ErrInternal("Failed to retry event: " + err.Error())
		}
		c.Set("audit_target", fmt.Sprintf("webhook_event:%d", req.EventID))
		return map[string]string{"message": "Event retried"}, nil
	})
}

// webhookRuleRequest is the create/update payload for a webhook rule.
type webhookRuleRequest struct {
	Name             string                 `json:"name"`
	ProviderConfigID uint                   `json:"provider_config_id"`
	EventTypePattern string                 `json:"event_type_pattern"`
	RepoPattern      string                 `json:"repo_pattern"`
	Action           string                 `json:"action"`
	ActionConfig     map[string]interface{} `json:"action_config"`
	Enabled          bool                   `json:"enabled"`
}

func (r *webhookRuleRequest) toPO() *po.WebhookRule {
	return &po.WebhookRule{
		Name:             r.Name,
		ProviderConfigID: r.ProviderConfigID,
		EventTypePattern: r.EventTypePattern,
		RepoPattern:      r.RepoPattern,
		Action:           r.Action,
		ActionConfig:     r.ActionConfig,
		Enabled:          r.Enabled,
	}
}

// webhookRuleDTO is the JSON response shape (snake_case, including the gorm
// primary key) so the rule API matches the rest of the snake_case contract.
type webhookRuleDTO struct {
	ID               uint                   `json:"id"`
	Name             string                 `json:"name"`
	ProviderConfigID uint                   `json:"provider_config_id"`
	EventTypePattern string                 `json:"event_type_pattern"`
	RepoPattern      string                 `json:"repo_pattern"`
	Action           string                 `json:"action"`
	ActionConfig     map[string]interface{} `json:"action_config"`
	Enabled          bool                   `json:"enabled"`
	CreatedAt        string                 `json:"created_at"`
	UpdatedAt        string                 `json:"updated_at"`
}

func toRuleDTO(r *po.WebhookRule) webhookRuleDTO {
	return webhookRuleDTO{
		ID:               r.ID,
		Name:             r.Name,
		ProviderConfigID: r.ProviderConfigID,
		EventTypePattern: r.EventTypePattern,
		RepoPattern:      r.RepoPattern,
		Action:           r.Action,
		ActionConfig:     r.ActionConfig,
		Enabled:          r.Enabled,
		CreatedAt:        r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        r.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ListRules lists all webhook rules.
func ListRules(ctx context.Context, c *app.RequestContext) {
	rules, err := db.NewWebhookRuleDAO().FindAll()
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to list webhook rules: "+err.Error())
		return
	}
	dtos := make([]webhookRuleDTO, 0, len(rules))
	for i := range rules {
		dtos = append(dtos, toRuleDTO(&rules[i]))
	}
	pkgresponse.Success(c, map[string]interface{}{"items": dtos})
}

// CreateRule creates a webhook rule.
func CreateRule(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *webhookRuleRequest) (webhookRuleDTO, error) {
		if req.Name == "" || req.Action == "" {
			return webhookRuleDTO{}, handler.ErrBadRequest("name and action are required")
		}
		rule := req.toPO()
		if err := db.NewWebhookRuleDAO().Create(rule); err != nil {
			return webhookRuleDTO{}, handler.ErrInternal("Failed to create webhook rule: " + err.Error())
		}
		c.Set("audit_target", "webhook_rule:"+rule.Name)
		return toRuleDTO(rule), nil
	})
}

// UpdateRule updates a webhook rule by id.
func UpdateRule(ctx context.Context, c *app.RequestContext) {
	id, ok := pkgresponse.ParseIDParam(c, "id")
	if !ok {
		return
	}
	handler.BindAndDo(c, func(req *webhookRuleRequest) (webhookRuleDTO, error) {
		rule := req.toPO()
		rule.ID = uint(id)
		if err := db.NewWebhookRuleDAO().Save(rule); err != nil {
			return webhookRuleDTO{}, handler.ErrInternal("Failed to update webhook rule: " + err.Error())
		}
		c.Set("audit_target", fmt.Sprintf("webhook_rule:%d", id))
		return toRuleDTO(rule), nil
	})
}

// DeleteRule deletes a webhook rule by id.
func DeleteRule(ctx context.Context, c *app.RequestContext) {
	id, ok := pkgresponse.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := db.NewWebhookRuleDAO().Delete(uint(id)); err != nil {
		pkgresponse.InternalServerError(c, "Failed to delete webhook rule: "+err.Error())
		return
	}
	c.Set("audit_target", fmt.Sprintf("webhook_rule:%d", id))
	pkgresponse.Success(c, map[string]string{"message": "Rule deleted"})
}
