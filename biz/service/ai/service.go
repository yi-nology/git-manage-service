package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/service/llm"
)

var (
	jsonBlockRe = regexp.MustCompile("(?s)```json\\s*(.*?)\\s*```")
)

type aiResponseMeta interface {
	SetRaw(raw string)
	SetInvocationID(id uint)
}

func fillResponseMeta(resp aiResponseMeta, raw string, invocationID uint) {
	resp.SetRaw(raw)
	resp.SetInvocationID(invocationID)
}

type Service struct {
	runner *Runner
}

func NewService() *Service {
	return &Service{
		runner: NewRunner(),
	}
}

func (s *Service) DiagnoseSyncFailure(ctx context.Context, req api.SyncFailureRequest) (*api.AIDiagnosisResponse, error) {
	context := BuildSyncFailureContext(req.Logs, req.Stderr, req.CurrentBranch, req.TrackingBranch, req.RecentActions, 30000)

	systemPrompt := `你是一个 Git 同步失败诊断专家。分析同步失败的原因，并给出具体的修复建议。

 请以 JSON 格式返回，结构如下：
 {
   "rootCause": "问题根因，简洁描述",
   "evidence": ["证据1", "证据2"],
   "recommendedActions": ["操作1", "操作2"],
   "canAutoFix": true/false,
   "riskLevel": "low/medium/high/critical",
   "fixDraft": "可自动执行的脚本或命令（如果可以）"
 }`

	if req.UserInstruction != "" {
		context += "\n\n## 用户额外要求\n" + req.UserInstruction
	}

	taskReq := TaskRequest{
		Type:         TaskSyncFailureAnalysis,
		SystemPrompt: systemPrompt,
		Messages:     []llm.ChatMessage{{Role: "user", Content: context}},
		RepoKey:      req.RepoKey,
	}

	resp, err := s.runner.Chat(ctx, taskReq)
	if err != nil {
		return nil, err
	}

	result := &api.AIDiagnosisResponse{}
	if err := parseJSONResponse(resp.Content, result); err != nil {
		result.RootCause = resp.Content
		result.RecommendedActions = []string{"请查看详细输出"}

	}
	fillResponseMeta(result, resp.Content, resp.InvocationID)

	return result, nil
}

func (s *Service) GenerateRepoSummary(ctx context.Context, repoReq api.RepoSummaryRequest) (*api.AIAdviceResponse, error) {
	branchLabel := fmt.Sprintf("%v", repoReq.Status["defaultBranch"])
	if branchLabel == "" || branchLabel == "<nil>" {
		branchLabel = fmt.Sprintf("%v", repoReq.Status["currentBranch"])
	}
	context := BuildRepoContext(repoReq.RepoKey,
		branchLabel,
		toInt64(repoReq.Status["branchCount"]),
		toInt64(repoReq.Status["tagCount"]),
		toInt64(repoReq.Status["commitCount"]),
	)

	if len(repoReq.Status) > 0 {
		var details []string
		for _, key := range []string{
			"currentBranch",
			"ahead",
			"behind",
			"stagedCount",
			"unstagedCount",
			"untrackedCount",
			"conflictedCount",
			"isClean",
			"isMerging",
			"isRebasing",
			"remoteCount",
			"hasRecentSyncFailure",
			"recentFailureCount",
		} {
			if value, ok := repoReq.Status[key]; ok {
				details = append(details, fmt.Sprintf("%s: %v", key, value))
			}
		}
		if len(details) > 0 {
			context += "\n\n## Workspace Status\n"
			context += strings.Join(details, "\n")
		}
	}

	if len(repoReq.Issues) > 0 {
		context += "\n\n## Known Issues\n"
		context += strings.Join(repoReq.Issues, "\n")
	}

	context += fmt.Sprintf("\n\n## Pending Changes\n%d files with uncommitted changes", repoReq.PendingChanges)

	if repoReq.UserInstruction != "" {
		context += "\n\n## 用户分析要求\n" + repoReq.UserInstruction
	}

	systemPrompt := `你是一个仓库健康度分析专家。分析仓库状态并给出优先级最高的 3-5 条建议。

 请以 JSON 格式返回，结构如下：
 {
   "summary": "仓库状态概述，一句话总结",
   "riskLevel": "low/medium/high/critical",
   "suggestions": ["建议1", "建议2", "建议3"],
   "actions": [{"id": "action_id", "label": "按钮文字", "type": "primary/secondary", "description": "操作说明"}]
 }`

	taskReq := TaskRequest{
		Type:         TaskRepoSummary,
		SystemPrompt: systemPrompt,
		Messages:     []llm.ChatMessage{{Role: "user", Content: context}},
		RepoKey:      repoReq.RepoKey,
	}

	resp, err := s.runner.Chat(ctx, taskReq)
	if err != nil {
		return nil, err
	}

	result := &api.AIAdviceResponse{}
	if err := parseJSONResponse(resp.Content, result); err != nil {
		result.Summary = resp.Content

	}
	fillResponseMeta(result, resp.Content, resp.InvocationID)

	return result, nil
}

func (s *Service) GenerateCommitMessage(ctx context.Context, req api.CommitMessageRequest) (*api.AIDraftResponse, error) {
	diff := req.Diff
	style := req.Style
	if style == "" {
		style = "conventional"
	}

	if len(diff) > 20000 {
		diff = ClampText(diff, 20000)
	}

	systemPrompt := fmt.Sprintf(`你是一个 Git 提交信息生成专家。根据 diff 生成符合 %s 风格的提交信息。

请以 JSON 格式返回，结构如下：
{
  "summary": "提交信息概述",
  "applyContent": "完整的提交信息字符串",
  "riskLevel": "low"
}

风格说明：
- simple: 简洁的一句话提交信息
- conventional: Conventional Commits 格式 (type(scope): description)
- detailed: 包含标题和详细改动列表`, style)

	taskReq := TaskRequest{
		Type:         TaskCommitMessage,
		SystemPrompt: systemPrompt,
		Messages:     []llm.ChatMessage{{Role: "user", Content: fmt.Sprintf("Diff:\n```diff\n%s\n```", diff)}},
		RepoKey:      req.RepoKey,
	}

	resp, err := s.runner.Chat(ctx, taskReq)
	if err != nil {
		return nil, err
	}

	result := &api.AIDraftResponse{}
	if err := parseJSONResponse(resp.Content, result); err != nil {
		result.Summary = "Generated commit message"
		result.ApplyContent = strings.TrimSpace(resp.Content)

	}
	fillResponseMeta(result, resp.Content, resp.InvocationID)

	return result, nil
}

func (s *Service) CodeReview(ctx context.Context, req api.CodeReviewRequest) (*api.AIReviewResponse, error) {
	context := BuildReviewContext(req.Diff, req.ChangedFiles, req.ExistingFindings, req.Language, 40000)

	if req.UserInstruction != "" {
		context += "\n\n## 用户特别指示\n" + req.UserInstruction
	}

	systemPrompt := `你是一个代码审查专家。分析代码变更并给出分层的审查结果。
 请以 JSON 格式返回，结构如下：
 {
   "summary": "审查总结，一句话概述整体质量",
   "blocking": [{"severity": "blocking", "category": "bug/security/performance", "message": "问题描述", "filePath": "文件路径", "startLine": 1, "endLine": 2, "suggestion": "修复建议"}],
   "highRisk": [{"severity": "high", ...}],
   "optional": [{"severity": "optional", ...}],
   "riskLevel": "low/medium/high/critical",
   "shouldMerge": true/false,
   "mergeNotes": "合并前需要注意的事项"
 }

 审查规则：
 1. blocking: 必须修复的严重 bug、安全漏洞、逻辑错误
 2. high: 建议修复的性能问题、坏味道、边界情况遗漏
 3. optional: 代码风格、命名建议、可维护性改进
 4. 不要重复 existingFindings 中已经报告的问题`

	taskReq := TaskRequest{
		Type:         TaskCodeReview,
		SystemPrompt: systemPrompt,
		Messages:     []llm.ChatMessage{{Role: "user", Content: context}},
		RepoKey:      req.RepoKey,
	}

	resp, err := s.runner.Chat(ctx, taskReq)
	if err != nil {
		return nil, err
	}

	type tempReviewResponse struct {
		*api.AIReviewResponse
		High []api.AIReviewFinding `json:"high,omitempty"`
	}
	result := &api.AIReviewResponse{}
	temp := &tempReviewResponse{AIReviewResponse: result}
	if err := parseJSONResponse(resp.Content, temp); err != nil {
		result.Summary = resp.Content

	}
	if len(temp.High) > 0 && len(result.HighRisk) == 0 {
		result.HighRisk = temp.High
	}
	fillResponseMeta(result, resp.Content, resp.InvocationID)

	return result, nil
}

func (s *Service) ReviewReplyDraft(ctx context.Context, req api.ReviewReplyRequest) (*api.AIDraftResponse, error) {
	tone := req.Tone
	if tone == "" {
		tone = "professional"
	}

	context := BuildReviewReplyContext(req.ReviewSummary, req.ReviewerComments, "", 20000)

	systemPrompt := fmt.Sprintf(`你是一个代码评审回复起草专家。根据评审意见生成%s风格的回复草稿。

请以 JSON 格式返回，结构如下：
{
  "summary": "回复概述",
  "applyContent": "完整的回复文本",
  "riskLevel": "low"
}

风格：professional（专业正式）/ friendly（友好协作）/ concise（简洁直接）`, tone)

	taskReq := TaskRequest{
		Type:         TaskCodeReviewReply,
		SystemPrompt: systemPrompt,
		Messages:     []llm.ChatMessage{{Role: "user", Content: context}},
		RepoKey:      req.RepoKey,
	}

	resp, err := s.runner.Chat(ctx, taskReq)
	if err != nil {
		return nil, err
	}

	result := &api.AIDraftResponse{}
	if err := parseJSONResponse(resp.Content, result); err != nil {
		result.Summary = "Generated reply draft"
		result.ApplyContent = strings.TrimSpace(resp.Content)

	}
	fillResponseMeta(result, resp.Content, resp.InvocationID)

	return result, nil
}

func (s *Service) ReviewSummary(ctx context.Context, req api.ReviewSummaryRequest) (*api.AIReviewResponse, error) {
	var findingsText string
	if len(req.Findings) > 0 {
		lines := make([]string, 0, len(req.Findings))
		for _, f := range req.Findings {
			lines = append(lines, fmt.Sprintf("- [%s] %s: %s\n  %s", f.Severity, f.FilePath, f.Title, f.Message))
		}
		findingsText = strings.Join(lines, "\n")
	} else {
		findingsText = "(无发现)"
	}

	context := fmt.Sprintf(`## 审查任务 #%s
仓库: %s
状态: %s

## 已有审查发现
%s

## 变更文件列表
%s`, req.TaskID, req.RepoKey, req.TaskStatus, findingsText, strings.Join(req.ChangedFiles, ", "))

	if req.UserInstruction != "" {
		context = context + "\n\n## 用户额外问题\n" + req.UserInstruction
	}

	systemPrompt := `你是一个代码评审结果分析专家。对已有审查结果进行二次分析和总结。

请以 JSON 格式返回，结构如下：
{
  "summary": "整体总结分析",
  "blocking": [{"severity": "critical", "category": "安全", "message": "描述", "filePath": "路径", "title": "标题"}],
  "highRisk": [{"severity": "high", "category": "风险", "message": "描述", "filePath": "路径", "title": "标题"}],
  "optional": [{"severity": "info", "category": "建议", "message": "描述", "filePath": "路径", "title": "标题"}],
  "riskLevel": "critical/high/medium/low",
  "shouldMerge": true/false,
  "mergeNotes": "合并前需要注意的事项"
}

分析要点：
1. 对已有的发现进行风险等级汇总和归纳
2. 给出明确的合并建议
3. 重点突出需要优先处理的问题
4. 如果用户有额外问题，优先针对问题给出针对性回答`

	taskReq := TaskRequest{
		Type:         TaskCodeReviewSummary,
		SystemPrompt: systemPrompt,
		Messages:     []llm.ChatMessage{{Role: "user", Content: context}},
		RepoKey:      req.RepoKey,
	}

	resp, err := s.runner.Chat(ctx, taskReq)
	if err != nil {
		return nil, err
	}

	result := &api.AIReviewResponse{}
	if err := parseJSONResponse(resp.Content, result); err != nil {
		result.Summary = resp.Content

	}
	fillResponseMeta(result, resp.Content, resp.InvocationID)

	return result, nil
}

func (s *Service) ResolveConflict(ctx context.Context, req api.ConflictResolveRequest) (*api.AIDraftResponse, error) {
	context := BuildConflictContext(req.ConflictDiff, req.OursBranch, req.TheirsBranch, 30000)

	systemPrompt := `你是一个 Git 冲突解决专家。分析冲突并生成合并后的代码。

请以 JSON 格式返回，结构如下：
{
  "summary": "冲突解决概述",
  "changeSummary": "变更说明",
  "applyContent": "合并后的完整代码",
  "riskLevel": "low/medium/high",
  "references": [{"type": "conflict", "id": "1", "label": "冲突位置"}]
}

解决策略：
1. 如果两边修改不重叠，尝试智能合并
2. 优先保留逻辑更完整的实现
3. 如果有 API 变更，优先保留兼容性
4. 必须生成可编译的有效代码`

	taskReq := TaskRequest{
		Type:         TaskConflictResolve,
		SystemPrompt: systemPrompt,
		Messages:     []llm.ChatMessage{{Role: "user", Content: context}},
		RepoKey:      req.RepoKey,
	}

	resp, err := s.runner.Chat(ctx, taskReq)
	if err != nil {
		return nil, err
	}

	result := &api.AIDraftResponse{}
	if err := parseJSONResponse(resp.Content, result); err != nil {
		result.Summary = "Conflict resolution draft"
		result.ApplyContent = strings.TrimSpace(resp.Content)

	}
	fillResponseMeta(result, resp.Content, resp.InvocationID)

	return result, nil
}

func (s *Service) ExplainConflict(ctx context.Context, req api.ConflictResolveRequest) (*api.AIAdviceResponse, error) {
	context := BuildConflictContext(req.ConflictDiff, req.OursBranch, req.TheirsBranch, 30000)

	systemPrompt := `你是一个 Git 冲突解释专家。分析冲突的原因并给出解决方案建议。

请以 JSON 格式返回，结构如下：
{
  "summary": "冲突概述",
  "riskLevel": "low/medium/high",
  "suggestions": ["方案1：保留 ours，理由...", "方案2：保留 theirs，理由...", "方案3：手动合并，理由..."],
  "references": [{"type": "section", "id": "1", "label": "冲突区块1"}]
}`

	taskReq := TaskRequest{
		Type:         TaskConflictExplain,
		SystemPrompt: systemPrompt,
		Messages:     []llm.ChatMessage{{Role: "user", Content: context}},
		RepoKey:      req.RepoKey,
	}

	resp, err := s.runner.Chat(ctx, taskReq)
	if err != nil {
		return nil, err
	}

	result := &api.AIAdviceResponse{}
	if err := parseJSONResponse(resp.Content, result); err != nil {
		result.Summary = resp.Content

	}
	fillResponseMeta(result, resp.Content, resp.InvocationID)

	return result, nil
}

func (s *Service) GenerateBranchRule(ctx context.Context, req api.BranchRuleRequest) (*api.AIDraftResponse, error) {
	b := NewContextBuilder()
	b.AddSection("Repository Type", req.RepoType)
	b.AddListSection("Existing Branches", req.ExistingBranches)

	systemPrompt := `你是一个分支策略专家。为仓库生成分支命名规则和保护策略。

请以 JSON 格式返回，结构如下：
{
  "summary": "分支策略概述",
  "applyContent": "完整的分支规则配置说明（Markdown 格式）",
  "riskLevel": "low"
}`

	taskReq := TaskRequest{
		Type:         TaskBranchRule,
		SystemPrompt: systemPrompt,
		Messages:     []llm.ChatMessage{{Role: "user", Content: b.Build()}},
		RepoKey:      req.RepoKey,
	}

	resp, err := s.runner.Chat(ctx, taskReq)
	if err != nil {
		return nil, err
	}

	result := &api.AIDraftResponse{}
	if err := parseJSONResponse(resp.Content, result); err != nil {
		result.Summary = "Branch rule recommendation"
		result.ApplyContent = strings.TrimSpace(resp.Content)

	}
	fillResponseMeta(result, resp.Content, resp.InvocationID)

	return result, nil
}

func (s *Service) GenerateSpecTemplate(ctx context.Context, req api.SpecTemplateRequest) (*api.AIDraftResponse, error) {
	b := NewContextBuilder()
	b.AddSection("Package Name", req.PackageName)
	b.AddSection("Spec Type", req.SpecType)

	if req.ExistingSpecContent != "" {
		if len(req.ExistingSpecContent) > 30000 {
			b.AddCodeSection("Existing Spec Reference", "yaml", ClampText(req.ExistingSpecContent, 30000))
		} else {
			b.AddCodeSection("Existing Spec Reference", "yaml", req.ExistingSpecContent)
		}
	}

	systemPrompt := `你是一个 RPM Spec 文件模板生成专家。生成符合规范的 spec 骨架。
请以 JSON 格式返回，结构如下：
{
  "summary": "模板说明",
  "changeSummary": "生成了哪些核心部分",
  "applyContent": "完整的 spec 文件内容",
  "riskLevel": "low"
}`

	taskReq := TaskRequest{
		Type:         TaskSpecGenerateTemplate,
		SystemPrompt: systemPrompt,
		Messages:     []llm.ChatMessage{{Role: "user", Content: b.Build()}},
		RepoKey:      req.RepoKey,
	}

	resp, err := s.runner.Chat(ctx, taskReq)
	if err != nil {
		return nil, err
	}

	result := &api.AIDraftResponse{}
	if err := parseJSONResponse(resp.Content, result); err != nil {
		result.Summary = "Spec template generated"
		result.ApplyContent = strings.TrimSpace(resp.Content)

	}
	fillResponseMeta(result, resp.Content, resp.InvocationID)

	return result, nil
}

func (s *Service) RewriteSpecSection(ctx context.Context, req api.SpecRewriteRequest) (*api.AIDraftResponse, error) {
	context := BuildSpecContext(req.SpecContent, req.SectionName, nil, 40000)
	context += fmt.Sprintf("\n\n## Rewrite Instruction\n%s", req.Instruction)

	systemPrompt := `你是一个 Spec 文件编辑专家。重写指定的 spec 部分，保持格式和风格一致。
请以 JSON 格式返回，结构如下：
{
  "summary": "修改概述",
  "changeSummary": "具体变更说明",
  "applyContent": "修改后的完整 spec 文件",
  "riskLevel": "low"
}`

	taskReq := TaskRequest{
		Type:         TaskSpecRewriteSection,
		SystemPrompt: systemPrompt,
		Messages:     []llm.ChatMessage{{Role: "user", Content: context}},
		RepoKey:      req.RepoKey,
	}

	resp, err := s.runner.Chat(ctx, taskReq)
	if err != nil {
		return nil, err
	}

	result := &api.AIDraftResponse{}
	if err := parseJSONResponse(resp.Content, result); err != nil {
		result.Summary = "Spec section modified"
		result.ApplyContent = strings.TrimSpace(resp.Content)

	}
	fillResponseMeta(result, resp.Content, resp.InvocationID)

	return result, nil
}

func (s *Service) RecommendProviderBinding(ctx context.Context, req api.ProviderBindingRequest) (*api.AIAdviceResponse, error) {
	context := BuildProviderBindingContext(req.RemoteRepos, req.LocalRepos, req.ExistingBindings, 20000)

	systemPrompt := `你是一个仓库绑定推荐专家。根据仓库名称、分支模式等特征，推荐可能的本地-远端仓库绑定关系。
请以 JSON 格式返回，结构如下：
{
  "summary": "绑定推荐概述",
  "riskLevel": "low",
  "suggestions": ["推荐绑定：local-repo-1 -> remote-repo-1，置信度：高，理由：..."],
  "references": []
}`

	taskReq := TaskRequest{
		Type:         TaskProviderRecommendation,
		SystemPrompt: systemPrompt,
		Messages:     []llm.ChatMessage{{Role: "user", Content: context}},
	}

	resp, err := s.runner.Chat(ctx, taskReq)
	if err != nil {
		return nil, err
	}

	result := &api.AIAdviceResponse{}
	if err := parseJSONResponse(resp.Content, result); err != nil {
		result.Summary = resp.Content

	}
	fillResponseMeta(result, resp.Content, resp.InvocationID)

	return result, nil
}

func (s *Service) AnalyzePatchRisk(ctx context.Context, req api.PatchAnalysisRequest) (*api.AIDiagnosisResponse, error) {
	context := BuildPatchAnalysisContext(req.PatchContent, req.TargetBranch, req.FileList, 40000)

	systemPrompt := `你是一个 Patch 风险分析专家。分析 patch 是否适合应用到目标分支。
请以 JSON 格式返回，结构如下：
{
  "rootCause": "风险概述",
  "evidence": [],
  "recommendedActions": [],
  "canAutoFix": false,
  "riskLevel": "low"
}`

	taskReq := TaskRequest{
		Type:         TaskPatchRiskAnalysis,
		SystemPrompt: systemPrompt,
		Messages:     []llm.ChatMessage{{Role: "user", Content: context}},
	}

	resp, err := s.runner.Chat(ctx, taskReq)
	if err != nil {
		return nil, err
	}

	result := &api.AIDiagnosisResponse{}
	if err := parseJSONResponse(resp.Content, result); err != nil {
		result.RootCause = resp.Content

	}
	fillResponseMeta(result, resp.Content, resp.InvocationID)

	return result, nil
}

func (s *Service) SummarizeAuditLogs(ctx context.Context, req api.AuditSummaryRequest) (*api.AIAdviceResponse, error) {
	context := BuildAuditSummaryContext(req.Events, req.Stats, req.Anomalies, 30000)

	systemPrompt := `你是一个审计日志分析专家。从大量审计日志中提炼操作主题和异常行为。
请以 JSON 格式返回，结构如下：
{
  "summary": "日志分析概述",
  "riskLevel": "low",
  "suggestions": [],
  "references": []
}`

	taskReq := TaskRequest{
		Type:         TaskAuditSummary,
		SystemPrompt: systemPrompt,
		Messages:     []llm.ChatMessage{{Role: "user", Content: context}},
	}

	resp, err := s.runner.Chat(ctx, taskReq)
	if err != nil {
		return nil, err
	}

	result := &api.AIAdviceResponse{}
	if err := parseJSONResponse(resp.Content, result); err != nil {
		result.Summary = resp.Content

	}
	fillResponseMeta(result, resp.Content, resp.InvocationID)

	return result, nil
}

func (s *Service) AnalyzeStatsInsight(ctx context.Context, req api.StatsInsightRequest) (*api.AIAdviceResponse, error) {
	context := BuildStatsInsightContext(req.Stats, req.Trends, req.AuthorActivity, 30000)

	systemPrompt := `你是一个开发洞察分析专家。从统计数据中提炼有价值的观察和建议。
请以 JSON 格式返回，结构如下：
{
  "summary": "统计洞察概述",
  "riskLevel": "low",
  "suggestions": [],
  "references": []
}`

	taskReq := TaskRequest{
		Type:         TaskStatsInsight,
		SystemPrompt: systemPrompt,
		Messages:     []llm.ChatMessage{{Role: "user", Content: context}},
	}

	resp, err := s.runner.Chat(ctx, taskReq)
	if err != nil {
		return nil, err
	}

	result := &api.AIAdviceResponse{}
	if err := parseJSONResponse(resp.Content, result); err != nil {
		result.Summary = resp.Content

	}
	fillResponseMeta(result, resp.Content, resp.InvocationID)

	return result, nil
}

func (s *Service) AnalyzeWebhookFailure(ctx context.Context, req api.WebhookFailureRequest) (*api.AIDiagnosisResponse, error) {
	context := BuildWebhookFailureContext(req.Payload, req.Response, req.StatusCode, req.EventType, 20000)

	systemPrompt := `你是一个 Webhook 失败诊断专家。分析 webhook 失败原因并给出修复建议。
请以 JSON 格式返回，结构如下：
{
  "rootCause": "失败根因",
  "evidence": [],
  "recommendedActions": [],
  "canAutoFix": false,
  "riskLevel": "low"
}`

	taskReq := TaskRequest{
		Type:         TaskWebhookFailure,
		SystemPrompt: systemPrompt,
		Messages:     []llm.ChatMessage{{Role: "user", Content: context}},
	}

	resp, err := s.runner.Chat(ctx, taskReq)
	if err != nil {
		return nil, err
	}

	result := &api.AIDiagnosisResponse{}
	if err := parseJSONResponse(resp.Content, result); err != nil {
		result.RootCause = resp.Content

	}
	fillResponseMeta(result, resp.Content, resp.InvocationID)

	return result, nil
}

func parseJSONResponse(content string, target interface{}) error {
	matches := jsonBlockRe.FindStringSubmatch(content)
	if len(matches) > 1 {
		return json.Unmarshal([]byte(matches[1]), target)
	}

	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")
	if jsonStart >= 0 && jsonEnd > jsonStart {
		return json.Unmarshal([]byte(content[jsonStart:jsonEnd+1]), target)
	}

	return fmt.Errorf("no JSON found in response")
}

func toInt64(v interface{}) int64 {
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case uint:
		return int64(val)
	default:
		return 0
	}
}

func (s *Service) SubmitUserFeedback(req api.AIFeedbackRequest) error {
	return RecordUserFeedback(req.InvocationID, req.Feedback)
}
