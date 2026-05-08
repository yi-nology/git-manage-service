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
	"github.com/yi-nology/git-manage-service/biz/service/provider"
	"github.com/yi-nology/git-manage-service/biz/service/rag"
	"github.com/yi-nology/git-manage-service/pkg/logger"
)

func persistFindings(taskID uint, findings []*Finding) (map[string]uint, error) {
	idMap := make(map[string]uint, len(findings))
	if len(findings) == 0 {
		return idMap, nil
	}
	dao := db.NewReviewFindingDAO()
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

func cleanupOldComments(ctx context.Context, p provider.Provider, owner, repo string, mrNum int, providerConfigID uint, mrIID string) {
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

func publishComments(ctx context.Context, p provider.Provider, owner, repo string, mrNum int, taskID uint, result *AggregatedResult, findingIDMap map[string]uint) error {
	commentDAO := db.NewReviewCommentDAO()

	summary := BuildSummaryComment(result)
	noteID, err := p.CreateNote(ctx, owner, repo, mrNum, summary)
	if err != nil {
		return fmt.Errorf("failed to post summary comment: %w", err)
	}
	commentDAO.Create(&po.ReviewComment{
		TaskID:            taskID,
		ProviderCommentID: noteID,
		CommentType:       "summary",
		Body:              summary,
		Status:            "posted",
	})

	for _, f := range result.Findings {
		if f.FilePath == "" || f.NewLine == 0 {
			continue
		}
		if f.Severity != SeverityCritical && f.Severity != SeverityHigh && f.Severity != SeverityMedium {
			continue
		}
		body := BuildInlineComment(f)
		discID, dErr := p.CreateDiscussion(ctx, owner, repo, mrNum, provider.DiscussionOptions{
			Body:     body,
			FilePath: f.FilePath,
			NewLine:  f.NewLine,
		})
		if dErr != nil {
			logger.ErrorWithErr("Failed to create inline discussion", dErr, logrus.Fields{"file": f.FilePath, "line": f.NewLine})
			continue
		}
		findingID := findingIDMap[f.Fingerprint]
		commentDAO.Create(&po.ReviewComment{
			TaskID:            taskID,
			FindingID:         findingID,
			ProviderCommentID: discID,
			CommentType:       "inline",
			FilePath:          f.FilePath,
			LineNumber:        f.NewLine,
			Body:              body,
			Status:            "posted",
		})
	}

	return nil
}

func runLLMReview(ctx context.Context, files []*FileDiff, repoName, owner, providerName string, repoID uint) ([]*Finding, *ProcessStep) {
	diff := buildDiffString(files)
	if diff == "" {
		return nil, nil
	}

	ragContext := retrieveRAGContext(ctx, files, repoID)

	systemPrompt := buildSystemPromptWithRules()

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
			Fingerprint: computeFingerprint("llm:"+resp.ProviderName, lf.FilePath, lf.LineNumber, lf.Title),
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

const systemPromptPrefix = `You are an expert code reviewer.

OUTPUT FORMAT:
Respond ONLY with valid JSON. Do NOT wrap in markdown code blocks. Use this exact format:
{"findings":[{"file_path":"actual/file/path.go","line_number":42,"severity":"critical|high|medium|low|info","title":"问题标题","message":"详细说明","suggestion":"修复建议"}],"summary":"整体审查总结"}

LINE NUMBER RULES (CRITICAL):
- line_number MUST be the line number shown in the RIGHT side of the diff (the "+" line number from the @@ header range).
- For example, in "@@ -10,5 +20,7 @@", added lines start at line 20. Each "+" line increments by 1.
- DO NOT use relative positions, sequential counters, or line numbers from the original file.
- Look at the @@ -X,Y +N,M @@ headers. N is the starting new-file line number. Count from there.

FILE PATH RULES:
- file_path MUST be the actual file path from the diff (e.g., "internal/service/crawler.go"), NOT placeholder or invented paths.`

const intentAnalysisPrompt = `CHANGE INTENT ANALYSIS:
Before listing findings, you MUST first analyze and understand what the author is trying to accomplish with this change:
1. What is the overall purpose of this change? (bug fix, new feature, refactor, optimization, etc.)
2. What problem is the author solving?
3. What is the expected behavior after this change?

Include your intent analysis in the "summary" field. This helps you review the code in context rather than flagging things that are intentional.

When reviewing, distinguish between:
- Actual bugs or security issues (MUST report)
- Style preferences that conflict with the author's intent (report as low/info)
- Areas where the implementation could better achieve the stated intent (report with constructive suggestions)`

const systemPromptSuffix = `FINAL RULES:
- Only report real issues visible in the diff, not opinions about file organization
- If no real issues found, return {"findings":[],"summary":"未发现问题"}
- All title, message, suggestion and summary fields MUST be written in Chinese (中文)
- file_path and line_number must be accurate — do NOT fabricate values`

func buildSystemPromptWithRules() string {
	var b strings.Builder
	b.WriteString(systemPromptPrefix)
	b.WriteString("\n\n")
	b.WriteString(intentAnalysisPrompt)
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
