package ai

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	aiSvc "github.com/yi-nology/git-manage-service/biz/service/ai"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

var aiService = aiSvc.NewService()

// handleAIRequest is the shared bind→invoke→respond shape for every AI
// endpoint handler below (they were 17 byte-identical copies).
func handleAIRequest[Req any, Resp any](ctx context.Context, c *app.RequestContext, fn func(ctx context.Context, req Req) (Resp, error)) {
	var req Req
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := fn(ctx, req)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, resp)
}

// DiagnoseSyncFailure .
// @router /api/v1/ai/sync/failure [POST]
func DiagnoseSyncFailure(ctx context.Context, c *app.RequestContext) {
	handleAIRequest(ctx, c, aiService.DiagnoseSyncFailure)
}

// GenerateRepoSummary .
// @router /api/v1/ai/repo/summary [POST]
func GenerateRepoSummary(ctx context.Context, c *app.RequestContext) {
	handleAIRequest(ctx, c, aiService.GenerateRepoSummary)
}

// GenerateCommitMessage .
// @router /api/v1/ai/commit/message [POST]
func GenerateCommitMessage(ctx context.Context, c *app.RequestContext) {
	handleAIRequest(ctx, c, aiService.GenerateCommitMessage)
}

// CodeReview .
// @router /api/v1/ai/review [POST]
func CodeReview(ctx context.Context, c *app.RequestContext) {
	handleAIRequest(ctx, c, aiService.CodeReview)
}

// ReviewReplyDraft .
// @router /api/v1/ai/review/reply [POST]
func ReviewReplyDraft(ctx context.Context, c *app.RequestContext) {
	handleAIRequest(ctx, c, aiService.ReviewReplyDraft)
}

// ReviewSummary .
// @router /api/v1/ai/review/summary [POST]
func ReviewSummary(ctx context.Context, c *app.RequestContext) {
	handleAIRequest(ctx, c, aiService.ReviewSummary)
}

// ResolveConflict .
// @router /api/v1/ai/conflict/resolve [POST]
func ResolveConflict(ctx context.Context, c *app.RequestContext) {
	handleAIRequest(ctx, c, aiService.ResolveConflict)
}

// ExplainConflict .
// @router /api/v1/ai/conflict/explain [POST]
func ExplainConflict(ctx context.Context, c *app.RequestContext) {
	handleAIRequest(ctx, c, aiService.ExplainConflict)
}

// GenerateBranchRule .
// @router /api/v1/ai/branch/rule [POST]
func GenerateBranchRule(ctx context.Context, c *app.RequestContext) {
	handleAIRequest(ctx, c, aiService.GenerateBranchRule)
}

// GenerateSpecTemplate .
// @router /api/v1/ai/spec/template [POST]
func GenerateSpecTemplate(ctx context.Context, c *app.RequestContext) {
	handleAIRequest(ctx, c, aiService.GenerateSpecTemplate)
}

// RewriteSpecSection .
// @router /api/v1/ai/spec/rewrite [POST]
func RewriteSpecSection(ctx context.Context, c *app.RequestContext) {
	handleAIRequest(ctx, c, aiService.RewriteSpecSection)
}

// RecommendProviderBinding .
// @router /api/v1/ai/provider/binding [POST]
func RecommendProviderBinding(ctx context.Context, c *app.RequestContext) {
	handleAIRequest(ctx, c, aiService.RecommendProviderBinding)
}

// AnalyzePatchRisk .
// @router /api/v1/ai/patch/analyze [POST]
func AnalyzePatchRisk(ctx context.Context, c *app.RequestContext) {
	handleAIRequest(ctx, c, aiService.AnalyzePatchRisk)
}

// SummarizeAuditLogs .
// @router /api/v1/ai/audit/summary [POST]
func SummarizeAuditLogs(ctx context.Context, c *app.RequestContext) {
	handleAIRequest(ctx, c, aiService.SummarizeAuditLogs)
}

// AnalyzeStatsInsight .
// @router /api/v1/ai/stats/insight [POST]
func AnalyzeStatsInsight(ctx context.Context, c *app.RequestContext) {
	handleAIRequest(ctx, c, aiService.AnalyzeStatsInsight)
}

// AnalyzeWebhookFailure .
// @router /api/v1/ai/webhook/failure [POST]
func AnalyzeWebhookFailure(ctx context.Context, c *app.RequestContext) {
	handleAIRequest(ctx, c, aiService.AnalyzeWebhookFailure)
}

// SubmitUserFeedback .
// @router /api/v1/ai/feedback [POST]
func SubmitUserFeedback(ctx context.Context, c *app.RequestContext) {
	var req api.AIFeedbackRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := aiService.SubmitUserFeedback(req); err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, map[string]bool{"success": true})
}
