package ai

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	aiSvc "github.com/yi-nology/git-manage-service/biz/service/ai"
	"github.com/yi-nology/git-manage-service/pkg/handler"
)

var aiService = aiSvc.NewService()

// DiagnoseSyncFailure .
// @router /api/v1/ai/sync/failure [POST]
func DiagnoseSyncFailure(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.SyncFailureRequest) (*api.AIDiagnosisResponse, error) {
		return aiService.DiagnoseSyncFailure(ctx, *req)
	})
}

// GenerateRepoSummary .
// @router /api/v1/ai/repo/summary [POST]
func GenerateRepoSummary(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.RepoSummaryRequest) (*api.AIAdviceResponse, error) {
		return aiService.GenerateRepoSummary(ctx, *req)
	})
}

// GenerateCommitMessage .
// @router /api/v1/ai/commit/message [POST]
func GenerateCommitMessage(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.CommitMessageRequest) (*api.AIDraftResponse, error) {
		return aiService.GenerateCommitMessage(ctx, *req)
	})
}

// CodeReview .
// @router /api/v1/ai/review [POST]
func CodeReview(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.CodeReviewRequest) (*api.AIReviewResponse, error) {
		return aiService.CodeReview(ctx, *req)
	})
}

// ReviewReplyDraft .
// @router /api/v1/ai/review/reply [POST]
func ReviewReplyDraft(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.ReviewReplyRequest) (*api.AIDraftResponse, error) {
		return aiService.ReviewReplyDraft(ctx, *req)
	})
}

// ReviewSummary .
// @router /api/v1/ai/review/summary [POST]
func ReviewSummary(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.ReviewSummaryRequest) (*api.AIReviewResponse, error) {
		return aiService.ReviewSummary(ctx, *req)
	})
}

// ResolveConflict .
// @router /api/v1/ai/conflict/resolve [POST]
func ResolveConflict(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.ConflictResolveRequest) (*api.AIDraftResponse, error) {
		return aiService.ResolveConflict(ctx, *req)
	})
}

// ExplainConflict .
// @router /api/v1/ai/conflict/explain [POST]
func ExplainConflict(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.ConflictResolveRequest) (*api.AIAdviceResponse, error) {
		return aiService.ExplainConflict(ctx, *req)
	})
}

// GenerateBranchRule .
// @router /api/v1/ai/branch/rule [POST]
func GenerateBranchRule(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.BranchRuleRequest) (*api.AIDraftResponse, error) {
		return aiService.GenerateBranchRule(ctx, *req)
	})
}

// GenerateSpecTemplate .
// @router /api/v1/ai/spec/template [POST]
func GenerateSpecTemplate(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.SpecTemplateRequest) (*api.AIDraftResponse, error) {
		return aiService.GenerateSpecTemplate(ctx, *req)
	})
}

// RewriteSpecSection .
// @router /api/v1/ai/spec/rewrite [POST]
func RewriteSpecSection(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.SpecRewriteRequest) (*api.AIDraftResponse, error) {
		return aiService.RewriteSpecSection(ctx, *req)
	})
}

// RecommendProviderBinding .
// @router /api/v1/ai/provider/binding [POST]
func RecommendProviderBinding(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.ProviderBindingRequest) (*api.AIAdviceResponse, error) {
		return aiService.RecommendProviderBinding(ctx, *req)
	})
}

// AnalyzePatchRisk .
// @router /api/v1/ai/patch/analyze [POST]
func AnalyzePatchRisk(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.PatchAnalysisRequest) (*api.AIDiagnosisResponse, error) {
		return aiService.AnalyzePatchRisk(ctx, *req)
	})
}

// SummarizeAuditLogs .
// @router /api/v1/ai/audit/summary [POST]
func SummarizeAuditLogs(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.AuditSummaryRequest) (*api.AIAdviceResponse, error) {
		return aiService.SummarizeAuditLogs(ctx, *req)
	})
}

// AnalyzeStatsInsight .
// @router /api/v1/ai/stats/insight [POST]
func AnalyzeStatsInsight(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.StatsInsightRequest) (*api.AIAdviceResponse, error) {
		return aiService.AnalyzeStatsInsight(ctx, *req)
	})
}

// AnalyzeWebhookFailure .
// @router /api/v1/ai/webhook/failure [POST]
func AnalyzeWebhookFailure(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.WebhookFailureRequest) (*api.AIDiagnosisResponse, error) {
		return aiService.AnalyzeWebhookFailure(ctx, *req)
	})
}

// SubmitUserFeedback .
// @router /api/v1/ai/feedback [POST]
func SubmitUserFeedback(ctx context.Context, c *app.RequestContext) {
	handler.Do(c, func(req *api.AIFeedbackRequest) error {
		return aiService.SubmitUserFeedback(*req)
	})
}
