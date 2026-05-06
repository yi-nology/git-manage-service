package ai

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/service/llm"
)

type TaskType string
type RiskLevel string
type OutputMode string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

const (
	OutputModeText OutputMode = "text"
	OutputModeJSON OutputMode = "json"
)

const (
	TaskSpecChat               TaskType = "spec.chat"
	TaskSpecPatch              TaskType = "spec.patch"
	TaskSpecLint               TaskType = "spec.lint"
	TaskSpecFix                TaskType = "spec.fix"
	TaskSpecRewriteSection     TaskType = "spec.rewrite_section"
	TaskSpecExplainErrors      TaskType = "spec.explain_errors"
	TaskSpecGenerateTemplate   TaskType = "spec.generate_template"
	TaskCodeReview             TaskType = "code_review"
	TaskCodeReviewReply        TaskType = "code_review.reply_draft"
	TaskCodeReviewSummary      TaskType = "code_review.summary"
	TaskCommitMessage          TaskType = "commit_message"
	TaskCommitSummary          TaskType = "commit.summary"
	TaskConflictResolve        TaskType = "conflict_resolve"
	TaskConflictExplain        TaskType = "conflict.explain"
	TaskAuthorIdentity         TaskType = "author_identity"
	TaskAuthorMergePlan        TaskType = "author.merge_plan"
	TaskMaintenance            TaskType = "maintenance"
	TaskRepoSummary            TaskType = "repo.summary"
	TaskRepoRegisterAdvice     TaskType = "repo.register_advice"
	TaskRepoOperationRisk      TaskType = "repo.operation_risk"
	TaskFileExplain            TaskType = "file.explain"
	TaskBranchRisk             TaskType = "branch.risk"
	TaskBranchRule             TaskType = "branch.rule"
	TaskMergePlan              TaskType = "merge.plan"
	TaskSyncPlan               TaskType = "sync.plan"
	TaskSyncFailureAnalysis    TaskType = "sync.failure_analysis"
	TaskWebhookFailure         TaskType = "webhook.failure_analysis"
	TaskAuditSummary           TaskType = "audit.summary"
	TaskStatsInsight           TaskType = "stats.insight"
	TaskProviderRecommendation TaskType = "provider.binding_recommendation"
	TaskPatchRiskAnalysis      TaskType = "patch.risk_analysis"
	TaskSubmoduleExplain       TaskType = "submodule.explain"
	TaskNotificationTemplate   TaskType = "notification.template"
)

var taskDefaultConfig = map[TaskType]TaskConfig{
	TaskSpecChat:               {Timeout: 60 * time.Second, MaxInputChars: 60000, OutputMode: OutputModeText, AllowAutoApply: false},
	TaskSpecPatch:              {Timeout: 60 * time.Second, MaxInputChars: 60000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskSpecLint:               {Timeout: 120 * time.Second, MaxInputChars: 80000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskSpecFix:                {Timeout: 60 * time.Second, MaxInputChars: 60000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskSpecRewriteSection:     {Timeout: 60 * time.Second, MaxInputChars: 60000, OutputMode: OutputModeText, AllowAutoApply: false},
	TaskSpecExplainErrors:      {Timeout: 60 * time.Second, MaxInputChars: 40000, OutputMode: OutputModeText, AllowAutoApply: false},
	TaskSpecGenerateTemplate:   {Timeout: 90 * time.Second, MaxInputChars: 100000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskCodeReview:             {Timeout: 120 * time.Second, MaxInputChars: 80000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskCodeReviewReply:        {Timeout: 60 * time.Second, MaxInputChars: 40000, OutputMode: OutputModeText, AllowAutoApply: false},
	TaskCodeReviewSummary:      {Timeout: 60 * time.Second, MaxInputChars: 60000, OutputMode: OutputModeText, AllowAutoApply: false},
	TaskCommitMessage:          {Timeout: 30 * time.Second, MaxInputChars: 30000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskCommitSummary:          {Timeout: 60 * time.Second, MaxInputChars: 60000, OutputMode: OutputModeText, AllowAutoApply: false},
	TaskConflictResolve:        {Timeout: 60 * time.Second, MaxInputChars: 60000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskConflictExplain:        {Timeout: 30 * time.Second, MaxInputChars: 40000, OutputMode: OutputModeText, AllowAutoApply: false},
	TaskAuthorIdentity:         {Timeout: 60 * time.Second, MaxInputChars: 60000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskAuthorMergePlan:        {Timeout: 60 * time.Second, MaxInputChars: 60000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskMaintenance:            {Timeout: 60 * time.Second, MaxInputChars: 40000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskRepoSummary:            {Timeout: 60 * time.Second, MaxInputChars: 60000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskRepoRegisterAdvice:     {Timeout: 60 * time.Second, MaxInputChars: 40000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskRepoOperationRisk:      {Timeout: 30 * time.Second, MaxInputChars: 30000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskFileExplain:            {Timeout: 60 * time.Second, MaxInputChars: 60000, OutputMode: OutputModeText, AllowAutoApply: false},
	TaskBranchRisk:             {Timeout: 30 * time.Second, MaxInputChars: 30000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskBranchRule:             {Timeout: 60 * time.Second, MaxInputChars: 40000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskMergePlan:              {Timeout: 60 * time.Second, MaxInputChars: 60000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskSyncPlan:               {Timeout: 60 * time.Second, MaxInputChars: 40000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskSyncFailureAnalysis:    {Timeout: 60 * time.Second, MaxInputChars: 60000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskWebhookFailure:         {Timeout: 30 * time.Second, MaxInputChars: 30000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskAuditSummary:           {Timeout: 60 * time.Second, MaxInputChars: 80000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskStatsInsight:           {Timeout: 60 * time.Second, MaxInputChars: 60000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskProviderRecommendation: {Timeout: 30 * time.Second, MaxInputChars: 30000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskPatchRiskAnalysis:      {Timeout: 60 * time.Second, MaxInputChars: 60000, OutputMode: OutputModeJSON, AllowAutoApply: false},
	TaskSubmoduleExplain:       {Timeout: 30 * time.Second, MaxInputChars: 20000, OutputMode: OutputModeText, AllowAutoApply: false},
	TaskNotificationTemplate:   {Timeout: 30 * time.Second, MaxInputChars: 30000, OutputMode: OutputModeJSON, AllowAutoApply: false},
}

type TaskConfig struct {
	Timeout        time.Duration
	MaxInputChars  int
	MaxTokens      int
	OutputMode     OutputMode
	AllowAutoApply bool
}

func GetTaskConfig(taskType TaskType) TaskConfig {
	if cfg, ok := taskDefaultConfig[taskType]; ok {
		return cfg
	}
	return TaskConfig{
		Timeout:        60 * time.Second,
		MaxInputChars:  60000,
		OutputMode:     OutputModeText,
		AllowAutoApply: false,
	}
}

type ProviderSelection struct {
	Name string
	ID   uint
}

type TaskRequest struct {
	Type          TaskType
	PromptVersion string
	Provider      ProviderSelection
	SystemPrompt  string
	Messages      []llm.ChatMessage
	MaxTokens     int
	MaxInputChars int
	Timeout       time.Duration
	Metadata      map[string]string
	OutputMode    OutputMode
	RepoKey       string
	OperatorID    uint
	RelatedID     string
}

type TaskResponse struct {
	Content       string
	Raw           string
	ProviderName  string
	PromptVersion string
	TaskType      TaskType
	LatencyMs     int64
	InvocationID  uint
}
