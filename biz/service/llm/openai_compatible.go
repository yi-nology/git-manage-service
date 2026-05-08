package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	oai "github.com/sashabaranov/go-openai"
)

type OpenAICompatible struct {
	client    *oai.Client
	model     string
	maxTokens int
	name      string
}

func NewOpenAICompatible(baseURL, apiKey, model string, maxTokens int, name string) *OpenAICompatible {
	cfg := oai.DefaultConfig(apiKey)
	if baseURL != "" {
		cfg.BaseURL = strings.TrimRight(baseURL, "/")
		if !strings.Contains(cfg.BaseURL, "/v1") && !strings.Contains(cfg.BaseURL, "/v4") {
			cfg.BaseURL += "/v1"
		}
	}
	return &OpenAICompatible{
		client:    oai.NewClientWithConfig(cfg),
		model:     model,
		maxTokens: maxTokens,
		name:      name,
	}
}

func (p *OpenAICompatible) Name() string { return p.name }

func (p *OpenAICompatible) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	msgs := make([]oai.ChatCompletionMessage, 0, len(req.Messages)+1)
	if req.SystemPrompt != "" {
		msgs = append(msgs, oai.ChatCompletionMessage{Role: oai.ChatMessageRoleSystem, Content: req.SystemPrompt})
	}
	for _, m := range req.Messages {
		msgs = append(msgs, oai.ChatCompletionMessage{Role: m.Role, Content: m.Content})
	}
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = p.maxTokens
	}
	if maxTokens < 8192 {
		maxTokens = 8192
	}
	resp, err := p.client.CreateChatCompletion(ctx, oai.ChatCompletionRequest{
		Model:       p.model,
		Messages:    msgs,
		Temperature: 0.3,
		MaxTokens:   maxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("openai compatible chat error: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}
	content := resp.Choices[0].Message.Content
	return &ChatResponse{Content: content, Raw: content}, nil
}

func (p *OpenAICompatible) Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error) {
	prompt := buildReviewPrompt(req)

	maxTokens := p.maxTokens
	if maxTokens < 8192 {
		maxTokens = 8192
	}
	resp, err := p.client.CreateChatCompletion(ctx, oai.ChatCompletionRequest{
		Model: p.model,
		Messages: []oai.ChatCompletionMessage{
			{Role: oai.ChatMessageRoleSystem, Content: systemPrompt},
			{Role: oai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: 0.2,
		MaxTokens:   maxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("openai compatible API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from LLM")
	}

	content := resp.Choices[0].Message.Content
	return parseLLMResponse(content)
}

const systemPrompt = `You are an expert code reviewer. Analyze the provided diff and identify issues.

Respond ONLY with valid JSON in this exact format:
{
  "findings": [
    {
      "file_path": "path/to/file",
      "line_number": 42,
      "severity": "high",
      "title": "Brief issue title",
      "message": "Detailed explanation",
      "suggestion": "How to fix"
    }
  ],
  "summary": "Brief overall review summary"
}

Severity levels: critical, high, medium, low, info.
Focus on: bugs, security issues, performance, maintainability, error handling.
Do NOT report style/formatting issues unless they affect correctness.
If no issues found, return empty findings array.`

func buildReviewPrompt(req *ReviewRequest) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Review this diff for repository %s/%s", req.Owner, req.RepoName))
	if req.Language != "" {
		b.WriteString(fmt.Sprintf(" (%s)", req.Language))
	}
	b.WriteString(":\n\n```\n")
	b.WriteString(req.Diff)
	b.WriteString("\n```\n")

	if len(req.Files) > 0 {
		b.WriteString("\n\nFiles for context:\n")
		for _, f := range req.Files {
			if f.Content == "" || len(f.Content) > 5000 {
				continue
			}
			b.WriteString(fmt.Sprintf("\n--- %s ---\n```\n%s\n```\n", f.Path, f.Content))
		}
	}

	return b.String()
}

func parseLLMResponse(content string) (*ReviewResponse, error) {
	content = strings.TrimSpace(content)

	jsonStart := strings.Index(content, "{")
	jsonEnd := strings.LastIndex(content, "}")
	if jsonStart < 0 || jsonEnd < 0 || jsonEnd <= jsonStart {
		return &ReviewResponse{Findings: nil, Summary: "LLM response was not valid JSON", Raw: content}, nil
	}

	jsonStr := content[jsonStart : jsonEnd+1]

	var raw struct {
		Findings []LLMFinding `json:"findings"`
		Summary  string       `json:"summary"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &raw); err != nil {
		return &ReviewResponse{Findings: nil, Summary: "Failed to parse LLM JSON response", Raw: content}, nil
	}

	return &ReviewResponse{
		Findings: raw.Findings,
		Summary:  raw.Summary,
		Raw:      content,
	}, nil
}
