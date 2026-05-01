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

type Gemini struct {
	baseURL    string
	apiKey     string
	model      string
	maxTokens  int
	httpClient *http.Client
}

func NewGemini(baseURL, apiKey, model string, maxTokens int) *Gemini {
	return &Gemini{
		baseURL:   strings.TrimRight(baseURL, "/"),
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (p *Gemini) Name() string { return "gemini:" + p.model }

func (p *Gemini) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	contents := p.buildContents(req.Messages)
	systemInstruction := ""
	if req.SystemPrompt != "" {
		systemInstruction = req.SystemPrompt
	}
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = p.maxTokens
	}
	return p.doGenerate(ctx, contents, systemInstruction, maxTokens)
}

func (p *Gemini) Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error) {
	prompt := buildReviewPrompt(req)
	contents := []geminiContent{
		{Role: "user", Parts: []geminiPart{{Text: prompt}}},
	}
	resp, err := p.doGenerate(ctx, contents, systemPrompt, p.maxTokens)
	if err != nil {
		return nil, err
	}
	return parseLLMResponse(resp.Content)
}

type geminiRequest struct {
	Contents         []geminiContent    `json:"contents"`
	SystemInstruction *geminiSystemInst `json:"systemInstruction,omitempty"`
	GenerationConfig *geminiGenConfig   `json:"generationConfig,omitempty"`
}

type geminiContent struct {
	Role  string       `json:"role"`
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiSystemInst struct {
	Parts []geminiPart `json:"parts"`
}

type geminiGenConfig struct {
	MaxOutputTokens int     `json:"maxOutputTokens"`
	Temperature     float64 `json:"temperature"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (p *Gemini) buildContents(messages []ChatMessage) []geminiContent {
	contents := make([]geminiContent, 0, len(messages))
	for _, m := range messages {
		role := "user"
		if m.Role == "assistant" {
			role = "model"
		}
		contents = append(contents, geminiContent{
			Role:  role,
			Parts: []geminiPart{{Text: m.Content}},
		})
	}
	return contents
}

func (p *Gemini) doGenerate(ctx context.Context, contents []geminiContent, systemPrompt string, maxTokens int) (*ChatResponse, error) {
	reqBody := geminiRequest{
		Contents: contents,
		GenerationConfig: &geminiGenConfig{
			MaxOutputTokens: maxTokens,
			Temperature:     0.3,
		},
	}
	if systemPrompt != "" {
		reqBody.SystemInstruction = &geminiSystemInst{
			Parts: []geminiPart{{Text: systemPrompt}},
		}
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", p.baseURL, p.model, p.apiKey)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read gemini response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBytes, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode gemini response: %w", err)
	}
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}
	return &ChatResponse{Content: geminiResp.Candidates[0].Content.Parts[0].Text, Raw: string(respBytes)}, nil
}
