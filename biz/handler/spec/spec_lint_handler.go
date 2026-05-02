package spec

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	spec "github.com/yi-nology/git-manage-service/biz/model/spec"
	lintSvc "github.com/yi-nology/git-manage-service/biz/service/lint"
	specService "github.com/yi-nology/git-manage-service/biz/service/spec"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func LintSpec(ctx context.Context, c *app.RequestContext) {
	var req spec.LintRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	content := req.GetContent()
	if content == "" {
		response.BadRequest(c, "content is required")
		return
	}

	lintService := lintSvc.NewLintService()
	result, err := lintService.LintWithAI(ctx, content, req.GetRules(), req.GetMode())
	if err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, result)
}

func GetLintRules(ctx context.Context, c *app.RequestContext) {
	rules, err := db.NewLintRuleDAO().FindAll()
	if err != nil {
		response.InternalError(c, err)
		return
	}

	var dtos []api.LintRule
	for _, r := range rules {
		dtos = append(dtos, api.LintRule{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Category:    r.Category,
			Severity:    r.Severity,
			Pattern:     r.Pattern,
			Enabled:     r.Enabled,
			Priority:    r.Priority,
			CreatedAt:   r.CreatedAt,
			UpdatedAt:   r.UpdatedAt,
		})
	}

	if dtos == nil {
		dtos = []api.LintRule{}
	}

	response.Success(c, dtos)
}

func UpdateLintRule(ctx context.Context, c *app.RequestContext) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "rule id is required")
		return
	}

	var req api.UpdateLintRuleReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	dao := db.NewLintRuleDAO()
	rule, err := dao.FindByID(id)
	if err != nil {
		response.NotFound(c, "rule not found")
		return
	}

	if req.Name != "" {
		rule.Name = req.Name
	}
	if req.Description != "" {
		rule.Description = req.Description
	}
	if req.Category != "" {
		rule.Category = req.Category
	}
	if req.Severity != "" {
		rule.Severity = req.Severity
	}
	if req.Pattern != "" {
		rule.Pattern = req.Pattern
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if req.Priority != nil {
		rule.Priority = *req.Priority
	}

	if err := dao.Save(rule); err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, api.LintRule{
		ID:          rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		Category:    rule.Category,
		Severity:    rule.Severity,
		Pattern:     rule.Pattern,
		Enabled:     rule.Enabled,
		Priority:    rule.Priority,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	})
}

func CreateLintRule(ctx context.Context, c *app.RequestContext) {
	var req api.CreateLintRuleReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.ID == "" {
		response.BadRequest(c, "id is required")
		return
	}
	if req.Name == "" {
		response.BadRequest(c, "name is required")
		return
	}

	dao := db.NewLintRuleDAO()
	exists, _ := dao.ExistsByID(req.ID)
	if exists {
		response.BadRequest(c, "rule with this id already exists")
		return
	}

	rule := &po.LintRule{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		Category:    req.Category,
		Severity:    req.Severity,
		Pattern:     req.Pattern,
		Enabled:     req.Enabled,
		Priority:    req.Priority,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if rule.Category == "" {
		rule.Category = "custom"
	}
	if rule.Severity == "" {
		rule.Severity = "warning"
	}

	if err := dao.Create(rule); err != nil {
		response.InternalError(c, err)
		return
	}

	response.Success(c, api.LintRule{
		ID:          rule.ID,
		Name:        rule.Name,
		Description: rule.Description,
		Category:    rule.Category,
		Severity:    rule.Severity,
		Pattern:     rule.Pattern,
		Enabled:     rule.Enabled,
		Priority:    rule.Priority,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	})
}

func ValidateSpec(ctx context.Context, c *app.RequestContext) {
	var req spec.ValidateSpecRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	svc := specService.NewSpecService()
	result := svc.ValidateSpec(req.GetContent())

	response.Success(c, result)
}
