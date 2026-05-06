package codereview

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/ai"
	"github.com/yi-nology/git-manage-service/biz/service/llm"
	"github.com/yi-nology/git-manage-service/biz/service/provider"
)

func persistFindings(taskID uint, findings []*Finding) error {
	if len(findings) == 0 {
		return nil
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
	return dao.BatchCreate(records)
}

func publishComments(ctx context.Context, p provider.Provider, owner, repo string, mrNum int, taskID uint, result *AggregatedResult) error {
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
		if f.Severity != SeverityCritical && f.Severity != SeverityHigh {
			continue
		}
		body := BuildInlineComment(f)
		discID, dErr := p.CreateDiscussion(ctx, owner, repo, mrNum, provider.DiscussionOptions{
			Body:     body,
			FilePath: f.FilePath,
			NewLine:  f.NewLine,
		})
		if dErr != nil {
			log.Printf("[CodeReview] Failed to create inline discussion: %v", dErr)
			continue
		}
		commentDAO.Create(&po.ReviewComment{
			TaskID:            taskID,
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

func runLLMReview(ctx context.Context, files []*FileDiff, repoName, owner, providerName string) []*Finding {
	diff := buildDiffString(files)
	if diff == "" {
		return nil
	}

	resp, err := ai.NewRunner().Chat(ctx, ai.TaskRequest{
		Type:          ai.TaskCodeReview,
		PromptVersion: "code-review.v1",
		Provider:      ai.ProviderSelection{Name: providerName},
		SystemPrompt:  codeReviewSystemPrompt,
		Messages: []llm.ChatMessage{
			{Role: "user", Content: buildCodeReviewPrompt(files, repoName, owner)},
		},
		MaxTokens: 4096,
	})
	if err != nil {
		log.Printf("[CodeReview] LLM review error: %v", err)
		return nil
	}

	var parsed struct {
		Findings []llm.LLMFinding `json:"findings"`
		Summary  string           `json:"summary"`
	}
	if !ai.DecodeJSON(resp.Content, &parsed) {
		log.Printf("[CodeReview] LLM review returned non-JSON response")
		return nil
	}

	var findings []*Finding
	for _, lf := range parsed.Findings {
		if !isValidLLMFinding(files, lf) {
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

	return findings
}

func isValidLLMFinding(files []*FileDiff, finding llm.LLMFinding) bool {
	if finding.FilePath == "" || finding.LineNumber <= 0 {
		return false
	}
	for _, file := range files {
		if file.NewPath != finding.FilePath && file.OldPath != finding.FilePath {
			continue
		}
		for _, hunk := range file.Hunks {
			for _, line := range hunk.Lines {
				if line.Type == "add" && line.NewLine == finding.LineNumber {
					return true
				}
			}
		}
	}
	return false
}

const codeReviewSystemPrompt = `You are an expert code reviewer. Analyze the provided diff and identify correctness, security, performance, maintainability, and error-handling issues.

Respond ONLY with valid JSON in this exact format:
{
  "findings": [
    {
      "file_path": "path/to/file",
      "line_number": 42,
      "severity": "critical|high|medium|low|info",
      "title": "Brief issue title",
      "message": "Detailed explanation",
      "suggestion": "How to fix"
    }
  ],
  "summary": "Brief overall review summary"
}

Only report issues you can ground in the diff. Do not report formatting or style unless it affects behavior. If no issues are found, return an empty findings array.`

func buildCodeReviewPrompt(files []*FileDiff, repoName, owner string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Review this diff for repository %s/%s:\n\n", owner, repoName))
	b.WriteString("```diff\n")
	b.WriteString(buildDiffString(files))
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
