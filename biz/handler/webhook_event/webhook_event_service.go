package webhook_event

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/webhookevent"
	pkgresponse "github.com/yi-nology/git-manage-service/pkg/response"
)

func List(ctx context.Context, c *app.RequestContext) {
	var req api.ListWebhookEventsReq
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}
	events, total, err := webhookevent.List(req.EventType, req.Source, req.Status, req.Page, req.PageSize)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to list webhook events: "+err.Error())
		return
	}
	pkgresponse.Success(c, map[string]interface{}{
		"items": events,
		"total": total,
	})
}

func Retry(ctx context.Context, c *app.RequestContext) {
	var req struct {
		EventID uint `json:"event_id"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if req.EventID == 0 {
		pkgresponse.BadRequest(c, "event_id is required")
		return
	}
	if err := webhookevent.Retry(req.EventID); err != nil {
		pkgresponse.InternalServerError(c, "Failed to retry event: "+err.Error())
		return
	}
	c.Set("audit_target", fmt.Sprintf("webhook_event:%d", req.EventID))
	pkgresponse.Success(c, map[string]string{"message": "Event retried"})
}

// ListRules .
// @router /api/v1/webhook/event-rules [GET]
func ListRules(ctx context.Context, c *app.RequestContext) {
	rules, err := db.NewWebhookEventRuleDAO().FindAll()
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}

	pkgresponse.Success(c, map[string]interface{}{
		"rules": rules,
		"total": len(rules),
	})
}

// GetRule .
// @router /api/v1/webhook/event-rules/:id [GET]
func GetRule(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkgresponse.BadRequest(c, "invalid id")
		return
	}

	rule, err := db.NewWebhookEventRuleDAO().FindByID(uint(id))
	if err != nil {
		pkgresponse.NotFound(c, "event rule not found")
		return
	}

	pkgresponse.Success(c, rule)
}

// CreateRule .
// @router /api/v1/webhook/event-rules [POST]
func CreateRule(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Name        string `json:"name"`
		EventType   string `json:"event_type"`
		Description string `json:"description"`
		MatchRules  string `json:"match_rules"`
		IsActive    *bool  `json:"is_active"`
		Priority    int    `json:"priority"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if req.Name == "" || req.EventType == "" || req.MatchRules == "" {
		pkgresponse.BadRequest(c, "name, event_type and match_rules are required")
		return
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	rule := &po.WebhookEventRule{
		Name:        req.Name,
		EventType:   req.EventType,
		Description: req.Description,
		MatchRules:  req.MatchRules,
		IsActive:    isActive,
		Priority:    req.Priority,
	}
	if err := db.NewWebhookEventRuleDAO().Create(rule); err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}

	c.Set("audit_target", "event_rule:"+req.Name)
	pkgresponse.Success(c, rule)
}

// UpdateRule .
// @router /api/v1/webhook/event-rules/:id [PUT]
func UpdateRule(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkgresponse.BadRequest(c, "invalid id")
		return
	}

	rule, err := db.NewWebhookEventRuleDAO().FindByID(uint(id))
	if err != nil {
		pkgresponse.NotFound(c, "event rule not found")
		return
	}

	var req struct {
		Name        string `json:"name"`
		EventType   string `json:"event_type"`
		Description string `json:"description"`
		MatchRules  string `json:"match_rules"`
		IsActive    *bool  `json:"is_active"`
		Priority    int    `json:"priority"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}

	if req.Name != "" {
		rule.Name = req.Name
	}
	if req.EventType != "" {
		rule.EventType = req.EventType
	}
	if req.Description != "" {
		rule.Description = req.Description
	}
	if req.MatchRules != "" {
		rule.MatchRules = req.MatchRules
	}
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}
	rule.Priority = req.Priority

	if err := db.NewWebhookEventRuleDAO().Save(rule); err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}

	pkgresponse.Success(c, rule)
}

// DeleteRule .
// @router /api/v1/webhook/event-rules/:id [DELETE]
func DeleteRule(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		pkgresponse.BadRequest(c, "invalid id")
		return
	}

	rule, err := db.NewWebhookEventRuleDAO().FindByID(uint(id))
	if err != nil {
		pkgresponse.NotFound(c, "event rule not found")
		return
	}

	if err := db.NewWebhookEventRuleDAO().Delete(rule); err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}

	pkgresponse.Success(c, map[string]string{"status": "deleted"})
}
