package ai

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	aiSvc "github.com/yi-nology/git-manage-service/biz/service/ai"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

var aiService = aiSvc.NewService()

// DiagnoseSyncFailure .
// @router /api/v1/ai/sync/failure [POST]
func DiagnoseSyncFailure(ctx context.Context, c *app.RequestContext) {
	var req api.SyncFailureRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := aiService.DiagnoseSyncFailure(ctx, req)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, resp)
}

// GenerateRepoSummary .
// @router /api/v1/ai/repo/summary [POST]
func GenerateRepoSummary(ctx context.Context, c *app.RequestContext) {
	var req api.RepoSummaryRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := aiService.GenerateRepoSummary(ctx, req)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, resp)
}

// GenerateCommitMessage .
// @router /api/v1/ai/commit/message [POST]
func GenerateCommitMessage(ctx context.Context, c *app.RequestContext) {
	var req api.CommitMessageRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := aiService.GenerateCommitMessage(ctx, req)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, resp)
}

// CodeReview .
// @router /api/v1/ai/review [POST]
func CodeReview(ctx context.Context, c *app.RequestContext) {
	var req api.CodeReviewRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := aiService.CodeReview(ctx, req)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, resp)
}

// ReviewReplyDraft .
// @router /api/v1/ai/review/reply [POST]
func ReviewReplyDraft(ctx context.Context, c *app.RequestContext) {
	var req api.ReviewReplyRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := aiService.ReviewReplyDraft(ctx, req)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, resp)
}

// ReviewSummary .
// @router /api/v1/ai/review/summary [POST]
func ReviewSummary(ctx context.Context, c *app.RequestContext) {
	var req api.ReviewSummaryRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := aiService.ReviewSummary(ctx, req)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, resp)
}

// ResolveConflict .
// @router /api/v1/ai/conflict/resolve [POST]
func ResolveConflict(ctx context.Context, c *app.RequestContext) {
	var req api.ConflictResolveRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := aiService.ResolveConflict(ctx, req)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, resp)
}

// ExplainConflict .
// @router /api/v1/ai/conflict/explain [POST]
func ExplainConflict(ctx context.Context, c *app.RequestContext) {
	var req api.ConflictResolveRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := aiService.ExplainConflict(ctx, req)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, resp)
}

// GenerateBranchRule .
// @router /api/v1/ai/branch/rule [POST]
func GenerateBranchRule(ctx context.Context, c *app.RequestContext) {
	var req api.BranchRuleRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := aiService.GenerateBranchRule(ctx, req)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, resp)
}

// GenerateSpecTemplate .
// @router /api/v1/ai/spec/template [POST]
func GenerateSpecTemplate(ctx context.Context, c *app.RequestContext) {
	var req api.SpecTemplateRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := aiService.GenerateSpecTemplate(ctx, req)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, resp)
}

// RewriteSpecSection .
// @router /api/v1/ai/spec/rewrite [POST]
func RewriteSpecSection(ctx context.Context, c *app.RequestContext) {
	var req api.SpecRewriteRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := aiService.RewriteSpecSection(ctx, req)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, resp)
}

// RecommendProviderBinding .
// @router /api/v1/ai/provider/binding [POST]
func RecommendProviderBinding(ctx context.Context, c *app.RequestContext) {
	var req api.ProviderBindingRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := aiService.RecommendProviderBinding(ctx, req)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, resp)
}

// AnalyzePatchRisk .
// @router /api/v1/ai/patch/analyze [POST]
func AnalyzePatchRisk(ctx context.Context, c *app.RequestContext) {
	var req api.PatchAnalysisRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := aiService.AnalyzePatchRisk(ctx, req)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, resp)
}

// SummarizeAuditLogs .
// @router /api/v1/ai/audit/summary [POST]
func SummarizeAuditLogs(ctx context.Context, c *app.RequestContext) {
	var req api.AuditSummaryRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := aiService.SummarizeAuditLogs(ctx, req)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, resp)
}

// AnalyzeStatsInsight .
// @router /api/v1/ai/stats/insight [POST]
func AnalyzeStatsInsight(ctx context.Context, c *app.RequestContext) {
	var req api.StatsInsightRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := aiService.AnalyzeStatsInsight(ctx, req)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, resp)
}

// AnalyzeWebhookFailure .
// @router /api/v1/ai/webhook/failure [POST]
func AnalyzeWebhookFailure(ctx context.Context, c *app.RequestContext) {
	var req api.WebhookFailureRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := aiService.AnalyzeWebhookFailure(ctx, req)
	if err != nil {
		response.InternalError(c, err)
		return
	}
	response.Success(c, resp)
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
