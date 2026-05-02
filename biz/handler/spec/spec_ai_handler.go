package spec

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	spec "github.com/yi-nology/git-manage-service/biz/model/spec"
	lintSvc "github.com/yi-nology/git-manage-service/biz/service/lint"
	"github.com/yi-nology/git-manage-service/biz/service/llm"
	specService "github.com/yi-nology/git-manage-service/biz/service/spec"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func AIAssistSpec(ctx context.Context, c *app.RequestContext) {
	var req spec.AIAssistRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	content := req.GetContent()
	prompt := req.GetPrompt()
	if content == "" || prompt == "" {
		response.BadRequest(c, "content and prompt are required")
		return
	}

	var history []llm.ChatMessage
	for _, h := range req.GetHistory() {
		history = append(history, llm.ChatMessage{Role: h.GetRole(), Content: h.GetContent()})
	}

	svc := specService.NewSpecService()
	result, applyContent, err := svc.AIAssist(ctx, content, prompt, req.GetAction(), history)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	resp := api.AIAssistResponse{Result: result, ApplyContent: applyContent}
	c.Set("audit_target", "repo:spec")
	c.Set("audit_details", map[string]string{"action": req.GetAction()})
	response.Success(c, resp)
}

func AIFixSpec(ctx context.Context, c *app.RequestContext) {
	var req spec.AIFixRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	content := req.GetContent()
	issue := req.GetIssue()
	if content == "" || issue == "" {
		response.BadRequest(c, "content and issue are required")
		return
	}

	result, err := lintSvc.AIFix(ctx, content, issue, int(req.GetLine()), req.GetSeverity())
	if err != nil {
		response.InternalError(c, err)
		return
	}
	c.Set("audit_target", "repo:spec")
	c.Set("audit_details", map[string]string{"severity": req.GetSeverity()})
	response.Success(c, api.AIFixResponse{Content: result})
}

func FormatSpec(ctx context.Context, c *app.RequestContext) {
	var err error
	var req spec.FormatSpecRequest
	err = c.BindAndValidate(&req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	content := req.GetContent()
	if content == "" {
		response.BadRequest(c, "content is required")
		return
	}

	opts := specService.FormatOptions{
		Curlify:         req.GetCurlify(),
		RemoveClean:     req.GetRemoveClean(),
		RemoveBuildRoot: req.GetRemoveBuildRoot(),
		RemoveGroup:     req.GetRemoveGroup(),
		LicenseSPDX:     req.GetLicenseSpdx(),
		SortDeps:        req.GetSortDeps(),
		TabToSpaces:     req.GetTabToSpaces(),
		IndentSize:      int(req.GetIndentSize()),
		PreambleOrder:   req.GetPreambleOrder(),
		AlignValues:     req.GetAlignValues(),
		PathMacros:      req.GetPathMacros(),
		UtilMacros:      req.GetUtilMacros(),
		CommonCleanup:   req.GetCommonCleanup(),
		ConditionalTrim: req.GetConditionalTrim(),
	}

	formatter := specService.NewSpecFormatter()
	formatted, changes, err := formatter.Format(content, opts)
	if err != nil {
		response.InternalError(c, err)
		return
	}

	var dtos []api.FormatChangeDTO
	for _, ch := range changes {
		dtos = append(dtos, api.FormatChangeDTO{
			Line:   ch.Line,
			Type:   ch.Type,
			Before: ch.Before,
			After:  ch.After,
			Reason: ch.Reason,
		})
	}
	if dtos == nil {
		dtos = []api.FormatChangeDTO{}
	}

	response.Success(c, api.FormatResponse{
		Content: formatted,
		Changes: dtos,
	})
}
