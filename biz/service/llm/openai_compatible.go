package llm

import (
	"context"
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
