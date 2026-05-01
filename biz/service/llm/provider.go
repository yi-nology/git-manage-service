package llm

import "context"

type ReviewRequest struct {
	Diff      string
	Files     []FileInfo
	RepoName  string
	Owner     string
	Language  string
}

type FileInfo struct {
	Path      string
	Content   string
	IsNew     bool
	IsDeleted bool
}

type ReviewResponse struct {
	Findings []LLMFinding
	Summary  string
	Raw      string
}

type LLMFinding struct {
	FilePath   string `json:"file_path"`
	LineNumber int    `json:"line_number"`
	Severity   string `json:"severity"`
	Title      string `json:"title"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Messages     []ChatMessage
	SystemPrompt string
	MaxTokens    int
}

type ChatResponse struct {
	Content string
	Raw     string
}

type Provider interface {
	Name() string
	Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error)
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
}
