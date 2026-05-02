package lint

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/llm"
	"github.com/yi-nology/git-manage-service/pkg/configs"
)

type LintService struct{}

func NewLintService() *LintService {
	return &LintService{}
}

type LintRequest struct {
	Content string   `json:"content"`
	Rules   []string `json:"rules,omitempty"`
}

type LintResult struct {
	File   string      `json:"file"`
	Issues []LintIssue `json:"issues"`
	Stats  LintStats   `json:"stats"`
}

type LintIssue struct {
	RuleID    string `json:"ruleId"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Line      int    `json:"line"`
	Column    int    `json:"column,omitempty"`
	EndLine   int    `json:"endLine,omitempty"`
	EndColumn int    `json:"endColumn,omitempty"`
	Source    string `json:"source,omitempty"`
	QuickFix  string `json:"quickFix,omitempty"`
}

type LintStats struct {
	ErrorCount   int `json:"errorCount"`
	WarningCount int `json:"warningCount"`
	InfoCount    int `json:"infoCount"`
}

func (s *LintService) Lint(content string, ruleIDs []string) (*LintResult, error) {
	var rules []po.LintRule
	var err error

	if len(ruleIDs) > 0 {
		rules, err = db.NewLintRuleDAO().FindByIDs(ruleIDs)
	} else {
		rules, err = db.NewLintRuleDAO().FindEnabled()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to load lint rules: %v", err)
	}

	result := &LintResult{
		File:   "",
		Issues: []LintIssue{},
		Stats:  LintStats{},
	}

	lines := strings.Split(content, "\n")

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		issues := s.applyRule(lines, rule)
		result.Issues = append(result.Issues, issues...)

		for _, issue := range issues {
			switch issue.Severity {
			case "error":
				result.Stats.ErrorCount++
			case "warning":
				result.Stats.WarningCount++
			case "info":
				result.Stats.InfoCount++
			}
		}
	}

	if configs.GlobalConfig.Lint.EnableRpmlint {
		if rpmlintIssues := s.runRpmlint(content, ""); len(rpmlintIssues) > 0 {
			result.Issues = append(result.Issues, rpmlintIssues...)
			for _, issue := range rpmlintIssues {
				switch issue.Severity {
				case "error":
					result.Stats.ErrorCount++
				case "warning":
					result.Stats.WarningCount++
				case "info":
					result.Stats.InfoCount++
				}
			}
		}
	}

	return result, nil
}

func (s *LintService) LintWithAI(ctx context.Context, content string, ruleIDs []string, mode string) (*LintResult, error) {
	if mode == "" {
		mode = "rule_only"
	}

	result, err := s.Lint(content, ruleIDs)
	if err != nil {
		return nil, err
	}

	if (mode == "rule_and_ai" || mode == "ai_only") && llm.HasDefaultProvider() {
		aiResult, aiErr := AILint(ctx, content, mode)
		if aiErr == nil && aiResult != nil {
			if mode == "ai_only" {
				result = aiResult
			} else {
				result.Issues = append(result.Issues, aiResult.Issues...)
				result.Stats.ErrorCount += aiResult.Stats.ErrorCount
				result.Stats.WarningCount += aiResult.Stats.WarningCount
				result.Stats.InfoCount += aiResult.Stats.InfoCount
			}
		}
	}

	return result, nil
}

func (s *LintService) hasField(lines []string, prefix string) bool {
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			return true
		}
	}
	return false
}

func (s *LintService) hasSection(lines []string, section string) bool {
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), section) {
			return true
		}
	}
	return false
}

func (s *LintService) findSectionLine(lines []string, section string) int {
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), section) {
			return i + 1
		}
	}
	return 0
}

func (s *LintService) runRpmlint(content string, specFilePath string) []LintIssue {
	_, err := exec.LookPath("rpmlint")
	if err != nil {
		return []LintIssue{}
	}

	tmpFile, err := os.CreateTemp("", "spec-*.spec")
	if err != nil {
		return []LintIssue{}
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(content)
	if err != nil {
		tmpFile.Close()
		return []LintIssue{}
	}
	tmpFile.Close()

	if specFilePath == "" {
		specFilePath = tmpFile.Name()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "rpmlint", "-f", "json", specFilePath)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if len(exitErr.Stderr) > 0 {
				_ = exitErr.Stderr
			}
		}
	}

	var rpmlintResult []struct {
		File     string `json:"file"`
		Line     int    `json:"line"`
		Message  string `json:"message"`
		Severity string `json:"severity"`
	}

	if err := json.Unmarshal(output, &rpmlintResult); err != nil {
		return []LintIssue{}
	}

	var issues []LintIssue
	for _, item := range rpmlintResult {
		severity := "info"
		if strings.Contains(strings.ToLower(item.Severity), "error") {
			severity = "error"
		} else if strings.Contains(strings.ToLower(item.Severity), "warning") {
			severity = "warning"
		}

		issues = append(issues, LintIssue{
			RuleID:   "rpmlint",
			Severity: severity,
			Message:  item.Message,
			Line:     item.Line,
		})
	}

	return issues
}
