package codereview

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
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

func runLLMReview(ctx context.Context, files []*FileDiff, repoName, owner string) []*Finding {
	llmProvider, err := llm.GetDefaultProvider()
	if err != nil {
		log.Printf("[CodeReview] LLM skipped: %v", err)
		return nil
	}

	diff := buildDiffString(files)
	if diff == "" {
		return nil
	}

	llmFiles := make([]llm.FileInfo, 0, len(files))
	for _, f := range files {
		if !f.IsDeleted && len(f.RawDiff) < 5000 {
			llmFiles = append(llmFiles, llm.FileInfo{
				Path:      f.NewPath,
				IsNew:     f.IsNew,
				IsDeleted: f.IsDeleted,
			})
		}
	}

	resp, err := llmProvider.Review(ctx, &llm.ReviewRequest{
		Diff:     diff,
		Files:    llmFiles,
		RepoName: repoName,
		Owner:    owner,
	})
	if err != nil {
		log.Printf("[CodeReview] LLM review error: %v", err)
		return nil
	}

	var findings []*Finding
	for _, lf := range resp.Findings {
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
			RuleID:      "llm:" + llmProvider.Name(),
			Source:      "llm",
			Severity:    sev,
			FilePath:    lf.FilePath,
			NewLine:     lf.LineNumber,
			Title:       lf.Title,
			Message:     lf.Message,
			Suggestion:  lf.Suggestion,
			Fingerprint: computeFingerprint("llm:"+llmProvider.Name(), lf.FilePath, lf.LineNumber, lf.Title),
		})
	}

	return findings
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
