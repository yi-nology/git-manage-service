package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Anthropic struct {
	baseURL    string
	apiKey     string
	model      string
	maxTokens  int
	httpClient *http.Client
}

func NewAnthropic(baseURL, apiKey, model string, maxTokens int) *Anthropic {
	return &Anthropic{
		baseURL:   strings.TrimRight(baseURL, "/"),
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (p *Anthropic) Name() string { return "anthropic:" + p.model }

func (p *Anthropic) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	messages := make([]anthropicMessage, 0, len(req.Messages))
	for _, m := range req.Messages {
		messages = append(messages, anthropicMessage{Role: m.Role, Content: m.Content})
	}
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = p.maxTokens
	}
	body := anthropicRequest{
		Model:     p.model,
		MaxTokens: maxTokens,
		Messages:  messages,
	}
	if req.SystemPrompt != "" {
		body.System = req.SystemPrompt
	}
	return p.doRequest(ctx, body)
}

func (p *Anthropic) Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error) {
	prompt := buildReviewPrompt(req)
	body := anthropicRequest{
		Model:     p.model,
		MaxTokens: p.maxTokens,
		System:    systemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
	}
	resp, err := p.doRequest(ctx, body)
	if err != nil {
		return nil, err
	}
	return parseLLMResponse(resp.Content)
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func (p *Anthropic) doRequest(ctx context.Context, reqBody anthropicRequest) (*ChatResponse, error) {
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := p.baseURL + "/v1/messages"
	if !strings.HasPrefix(p.baseURL, "https://api.anthropic.com") &&
		!strings.Contains(p.baseURL, "/v1") &&
		!strings.Contains(p.baseURL, "/messages") {
		if strings.Contains(p.baseURL, "anthropic") {
			url = p.baseURL
		}
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic API error: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read anthropic response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var anthropicResp anthropicResponse
	if err := json.Unmarshal(respBytes, &anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode anthropic response: %w", err)
	}
	if len(anthropicResp.Content) == 0 {
		return nil, fmt.Errorf("no response from Anthropic")
	}
	return &ChatResponse{Content: anthropicResp.Content[0].Text, Raw: string(respBytes)}, nil
}
