package codereview

import "context"

type CLIService interface {
	ReviewCode(ctx context.Context, req *CLIReviewRequest) (*CLIReviewResult, error)
	ValidateInstallation() error
	GetVersion() (string, error)
}

type CLIReviewRequest struct {
	RepoPath     string
	CommitRange  string
	CustomPrompt string
	MaxTokens    int
	WorkingDir   string
}

type CLIReviewResult struct {
	Content  string
	Score    int
	Issues   []CLIReviewIssue
	Summary  string
	Duration int
}

type CLIReviewIssue struct {
	FilePath   string `json:"file_path"`
	LineNumber int    `json:"line_number"`
	Severity   string `json:"severity"`
	Title      string `json:"title"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
	RuleID     string `json:"rule_id"`
}
