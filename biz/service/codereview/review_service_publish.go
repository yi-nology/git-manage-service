package codereview

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/ai"
	"github.com/yi-nology/git-manage-service/biz/service/llm"
	"github.com/yi-nology/git-manage-service/biz/service/rag"
	"github.com/yi-nology/git-manage-service/pkg/logger"
	"github.com/yi-nology/git-platform-sdk/provider"
)

func persistFindings(taskID uint, findings []*Finding) (map[string]uint, error) {
	idMap := make(map[string]uint, len(findings))
	if len(findings) == 0 {
		return idMap, nil
	}
	dao := db.NewReviewFindingDAO()
	if err := dao.DeleteByTaskID(taskID); err != nil {
		return idMap, fmt.Errorf("failed to clean old findings: %w", err)
	}
	records := make([]po.ReviewFinding, 0, len(findings))
	for _, f := range findings {
		raw, _ := json.Marshal(f)
		records = append(records, po.ReviewFinding{
			TaskID:      taskID,
			Source:      f.Source,
			RuleID:      f.RuleID,
			Severity:    string(f.Severity),
			FilePath:    f.FilePath,
			OldLine:     f.OldLine,
			NewLine:     f.NewLine,
			Title:       f.Title,
			Message:     f.Message,
			Suggestion:  f.Suggestion,
			Fingerprint: f.Fingerprint,
			RawPayload:  string(raw),
		})
	}
	if err := dao.BatchCreate(records); err != nil {
		return idMap, err
	}
	for i, f := range findings {
		idMap[f.Fingerprint] = records[i].ID
	}
	return idMap, nil
}

func cleanupOldComments(ctx context.Context, p provider.DiffManager, owner, repo string, mrNum string, providerConfigID uint, mrIID string) {
	oldComments, err := db.NewReviewCommentDAO().FindSummaryCommentsByMRIID(providerConfigID, mrIID)
	if err != nil {
		return
	}
	for _, c := range oldComments {
		if c.ProviderCommentID != "" {
			if dErr := p.DeleteNote(ctx, owner, repo, mrNum, c.ProviderCommentID); dErr != nil {
				logger.ErrorWithErr("Failed to delete old summary note", dErr, logrus.Fields{"note_id": c.ProviderCommentID})
			}
		}
	}
}

func publishComments(ctx context.Context, p provider.DiffManager, owner, repo string, mrNum string, taskID uint, commitSHA string, result *AggregatedResult, findingIDMap map[string]uint) error {
	commentDAO := db.NewReviewCommentDAO()
	summary := BuildSummaryComment(result)

	// Collect filtered findings and build inline review comments
	type inlineEntry struct {
		finding *Finding
		comment provider.ReviewComment
	}
	var entries []inlineEntry
	for _, f := range result.Findings {
		if f.FilePath == "" || f.NewLine == 0 {
			continue
		}
		if f.Severity != SeverityCritical && f.Severity != SeverityHigh && f.Severity != SeverityMedium {
			continue
		}
		entries = append(entries, inlineEntry{
			finding: f,
			comment: provider.ReviewComment{
				Path: f.FilePath,
				Body: BuildInlineComment(f),
				Line: f.NewLine,
			},
		})
	}

	reviewComments := make([]provider.ReviewComment, 0, len(entries))
	for _, e := range entries {
		reviewComments = append(reviewComments, e.comment)
	}

	// Single API call: summary body + all inline comments
	rm, ok := p.(provider.ReviewManager)
	if !ok {
		return fmt.Errorf("provider does not support ReviewManager interface")
	}
	reviewResult, err := rm.CreateReview(ctx, owner, repo, mrNum, provider.CreateReviewOptions{
		CommitID: commitSHA,
		Event:    "COMMENT",
		Body:     summary,
		Comments: reviewComments,
	})
	if err != nil {
		return fmt.Errorf("failed to create review: %w", err)
	}

	commentDAO.Create(&po.ReviewComment{
		TaskID:            taskID,
		ProviderCommentID: reviewResult.ID,
		CommentType:       "summary",
		Body:              summary,
		Status:            "posted",
	})

	for i, e := range entries {
		externalID := ""
		if reviewResult.Comments != nil && i < len(reviewResult.Comments) {
			externalID = reviewResult.Comments[i].ExternalID
		}
		findingID := findingIDMap[e.finding.Fingerprint]
		commentDAO.Create(&po.ReviewComment{
			TaskID:            taskID,
			FindingID:         findingID,
			ProviderCommentID: externalID,
			CommentType:       "inline",
			FilePath:          e.finding.FilePath,
			LineNumber:        e.finding.NewLine,
			Body:              e.comment.Body,
			Status:            "posted",
		})
	}

	return nil
}

func runLLMReview(ctx context.Context, files []*FileDiff, repoName, owner, providerName string, repoID uint, repoCfg *po.ReviewRepoConfig) ([]*Finding, *ProcessStep) {
	diff := buildDiffString(files)
	if diff == "" {
		return nil, nil
	}

	ragContext := retrieveRAGContext(ctx, files, repoID)

	systemPrompt := buildSystemPromptWithRules(repoCfg)

	userPrompt := buildCodeReviewPrompt(files, repoName, owner)
	if ragContext != "" {
		userPrompt = "## Relevant Code Context\n\n" + ragContext + "\n\n---\n\n" + userPrompt
	}

	resp, err := callLLMWithRetry(ctx, ai.NewRunner(), ai.TaskRequest{
		Type:          ai.TaskCodeReview,
		PromptVersion: "code-review.v1",
		Provider:      ai.ProviderSelection{Name: providerName},
		SystemPrompt:  systemPrompt,
		Messages: []llm.ChatMessage{
			{Role: "user", Content: userPrompt},
		},
		MaxTokens: 4096,
	})
	if err != nil {
		logger.ErrorWithErr("LLM review error", err, logrus.Fields{"provider": providerName})
		return nil, &ProcessStep{
			Name:   "LLM Review",
			Status: "error",
			Detail: fmt.Sprintf("LLM call failed (provider=%s): %v", providerName, err),
		}
	}

	var parsed struct {
		Findings []llm.LLMFinding `json:"findings"`
		Summary  string           `json:"summary"`
	}

	if !ai.DecodeJSON(resp.Content, &parsed) {
		repairJSON := strings.TrimRight(resp.Content, " \t\n\r`\"}")
		repairJSON += "]}"
		if ai.DecodeJSON(repairJSON, &parsed) {
			logger.Info("LLM response repaired (truncated JSON)", logrus.Fields{
				"provider":       resp.ProviderName,
				"findings_count": len(parsed.Findings),
			})
		} else {
			logger.Warn("LLM review returned non-JSON response", logrus.Fields{
				"provider":        resp.ProviderName,
				"content_preview": resp.Content[:min(300, len(resp.Content))],
			})
			return nil, &ProcessStep{
				Name:   "LLM Review",
				Status: "error",
				Detail: fmt.Sprintf("LLM returned non-JSON response (provider=%s)", resp.ProviderName),
			}
		}
	}
	logger.Info("LLM review parsed", logrus.Fields{
		"provider":       resp.ProviderName,
		"findings_count": len(parsed.Findings),
		"summary_len":    len(parsed.Summary),
	})

	var findings []*Finding
	filtered := 0
	for _, lf := range parsed.Findings {
		if !isValidLLMFinding(files, lf) {
			filtered++
			logger.Warn("LLM finding filtered (path not in diff)", logrus.Fields{
				"file_path": lf.FilePath, "line": lf.LineNumber, "title": lf.Title,
			})
			continue
		}
		sev := SeverityMedium
		switch lf.Severity {
		case "critical":
			sev = SeverityCritical
		case "high":
			sev = SeverityHigh
		case "medium":
			sev = SeverityMedium
		case "low":
			sev = SeverityLow
		case "info":
			sev = SeverityInfo
		}
		findings = append(findings, &Finding{
			RuleID:      "llm:" + resp.ProviderName,
			Source:      "llm",
			Severity:    sev,
			FilePath:    lf.FilePath,
			NewLine:     lf.LineNumber,
			Title:       lf.Title,
			Message:     lf.Message,
			Suggestion:  lf.Suggestion,
			Fingerprint: computeFingerprint("llm:"+resp.ProviderName, lf.FilePath, lf.LineNumber, ""),
		})
	}

	detail := fmt.Sprintf("%d findings from LLM (provider=%s", len(findings), resp.ProviderName)
	if filtered > 0 {
		detail += fmt.Sprintf(", %d filtered by validation", filtered)
	}
	detail += ")"
	return findings, &ProcessStep{
		Name:   "LLM Review",
		Status: "ok",
		Detail: detail,
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func isValidLLMFinding(files []*FileDiff, finding llm.LLMFinding) bool {
	if finding.FilePath == "" {
		return false
	}
	fp := normalizePath(finding.FilePath)
	// strip leading ./  that LLMs sometimes add
	fp = strings.TrimPrefix(fp, "./")

	for _, file := range files {
		newP := normalizePath(file.NewPath)
		oldP := normalizePath(file.OldPath)
		// exact match
		if newP == fp || oldP == fp {
			return true
		}
		// suffix: /path/to/file.go matches file.go
		if strings.HasSuffix(newP, "/"+fp) || strings.HasSuffix(oldP, "/"+fp) {
			return true
		}
		// finding path is a suffix of diff path: deploy/Dockerfile matches Dockerfile
		if strings.HasSuffix(fp, "/"+newP) || strings.HasSuffix(fp, "/"+oldP) {
			return true
		}
		// basename match with directory overlap
		fBase := filepath.Base(fp)
		if fBase == filepath.Base(newP) || fBase == filepath.Base(oldP) {
			return true
		}
	}
	return false
}

func normalizePath(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	for strings.HasPrefix(p, "./") {
		p = p[2:]
	}
	return p
}

var lineCommentPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(?:file|文件)[:\s]+(\S+?)[,\s]+(?:line|行)[:\s]+(\d+)`),
	regexp.MustCompile(`(\S+?):(\d+):?\s`),
	regexp.MustCompile(`(\S+?)[,\s]+(?:line|行)\s*(\d+)`),
}

const systemPromptPrefix = `你是一位资深代码审查专家。

输出格式：
仅返回合法 JSON，不要用 markdown 代码块包裹。严格使用以下格式：
{"findings":[{"file_path":"actual/file/path.go","line_number":42,"severity":"critical|high|medium|low|info","title":"问题标题","message":"详细说明","suggestion":"修复建议"}],"summary":"整体审查总结"}

行号规则（关键）：
- line_number 必须是 diff 中右侧（"+" 号一侧）的行号。
- 例如 "@@ -10,5 +20,7 @@" 中，新增行从第 20 行开始，每个 "+" 行递增 1。
- 不要使用相对位置、顺序计数或原始文件中的行号。
- 查看 @@ -X,Y +N,M @@ 标记，N 是新文件的起始行号，从这里开始计数。

文件路径规则：
- file_path 必须是 diff 中的实际文件路径（例如 "internal/service/crawler.go"），不要编造路径。`

const intentAnalysisPrompt = `变更意图分析：
在列出发现之前，你必须先分析并理解作者本次变更的目的：
1. 本次变更的整体目的是什么？（Bug 修复、新功能、重构、优化等）
2. 作者要解决什么问题？
3. 变更后的预期行为是什么？

将意图分析写在 "summary" 字段中。这有助于你在上下文中审查代码，而不是标记有意为之的内容。

审查时请区分：
- 真实的 Bug 或安全问题（必须报告）
- 与作者意图冲突的风格偏好（标记为 low/info）
- 实现方式可以改进以更好达成意图的地方（提供建设性建议）`

const systemPromptSuffix = `最终规则：
- 只报告 diff 中可见的真实问题，不要对文件组织方式发表意见
- 如果没有发现真实问题，返回 {"findings":[],"summary":"未发现问题"}
- 所有 title、message、suggestion 和 summary 字段必须使用中文撰写
- file_path 和 line_number 必须准确——不要编造数值`

func GetPromptStructure() map[string]interface{} {
	return map[string]interface{}{
		"prefix": systemPromptPrefix,
		"intent": intentAnalysisPrompt,
		"suffix": systemPromptSuffix,
	}
}

func buildSystemPromptWithRules(repoCfg ...*po.ReviewRepoConfig) string {
	prefix := systemPromptPrefix
	intent := intentAnalysisPrompt
	if len(repoCfg) > 0 && repoCfg[0] != nil {
		if repoCfg[0].PromptPrefixOverride != "" {
			prefix = repoCfg[0].PromptPrefixOverride
		}
		if repoCfg[0].PromptIntentOverride != "" {
			intent = repoCfg[0].PromptIntentOverride
		}
	}
	var b strings.Builder
	b.WriteString(prefix)
	b.WriteString("\n\n")
	b.WriteString(intent)
	rules, err := db.NewReviewRuleDAO().FindEnabledPromptRules()
	if err == nil && len(rules) > 0 {
		b.WriteString("\n\nCUSTOM REVIEW RULES (from user configuration):\n")
		for i, r := range rules {
			b.WriteString(fmt.Sprintf("%d. [%s] %s\n", i+1, r.Name, r.PromptText))
		}
	}
	b.WriteString("\n\n")
	b.WriteString(systemPromptSuffix)
	return b.String()
}

func buildCodeReviewPrompt(files []*FileDiff, repoName, owner string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Review this diff for repository %s/%s:\n\n", owner, repoName))
	b.WriteString("Files in this diff:\n")
	for i, f := range files {
		if i > 30 {
			b.WriteString(fmt.Sprintf("... and %d more files\n", len(files)-31))
			break
		}
		b.WriteString(fmt.Sprintf("- %s\n", f.NewPath))
	}
	b.WriteString("\n```diff\n")
	diffStr := buildDiffString(files)
	const maxDiffChars = 50000
	if len(diffStr) > maxDiffChars {
		diffStr = diffStr[:maxDiffChars] + "\n\n... (diff truncated: showing first ~50K chars of " + fmt.Sprintf("%d", len(diffStr)) + " total)"
	}
	b.WriteString(diffStr)
	b.WriteString("\n```\n")
	return b.String()
}

func buildDiffString(files []*FileDiff) string {
	var b strings.Builder
	for _, f := range files {
		if f.IsDeleted || f.RawDiff == "" {
			continue
		}
		b.WriteString(f.RawDiff)
	}
	return b.String()
}

func retrieveRAGContext(ctx context.Context, files []*FileDiff, repoID uint) string {
	svc := rag.DefaultService()
	if !svc.IsAvailable() || repoID == 0 {
		return ""
	}

	changedPaths := make([]string, 0, len(files))
	for _, f := range files {
		if f.NewPath != "" && !f.IsDeleted {
			changedPaths = append(changedPaths, f.NewPath)
		}
	}
	if len(changedPaths) == 0 {
		return ""
	}

	context, err := svc.RetrieveForReview(ctx, repoID, changedPaths)
	if err != nil {
		logger.Warn("RAG context retrieval failed (non-fatal)", logrus.Fields{"repo_id": repoID, "error": err.Error()})
		return ""
	}
	return context
}

func callLLMWithRetry(ctx context.Context, runner *ai.Runner, req ai.TaskRequest) (*ai.TaskResponse, error) {
	const maxAttempts = 3
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt*attempt) * time.Second
			logger.Info("LLM retry attempt", logrus.Fields{"attempt": attempt + 1, "backoff": backoff.String()})
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}
		resp, err := runner.Chat(ctx, req)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		logger.ErrorWithErr("LLM attempt failed", err, logrus.Fields{"attempt": attempt + 1})
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
	}
	return nil, lastErr
}
