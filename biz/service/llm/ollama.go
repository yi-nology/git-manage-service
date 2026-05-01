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

type Ollama struct {
	baseURL    string
	model      string
	maxTokens  int
	httpClient *http.Client
}

func NewOllama(baseURL, model string, maxTokens int) *Ollama {
	return &Ollama{
		baseURL:   strings.TrimRight(baseURL, "/"),
		model:     model,
		maxTokens: maxTokens,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

func (p *Ollama) Name() string { return "ollama:" + p.model }

func (p *Ollama) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	messages := make([]map[string]string, 0, len(req.Messages)+1)
	if req.SystemPrompt != "" {
		messages = append(messages, map[string]string{"role": "system", "content": req.SystemPrompt})
	}
	for _, m := range req.Messages {
		messages = append(messages, map[string]string{"role": m.Role, "content": m.Content})
	}
	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = p.maxTokens
	}
	body := map[string]interface{}{
		"model":    p.model,
		"messages": messages,
		"stream":   false,
		"options": map[string]interface{}{
			"num_predict": maxTokens,
		},
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	url := p.baseURL + "/api/chat"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama chat error: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(respBody))
	}
	var ollamaResp struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode ollama chat response: %w", err)
	}
	return &ChatResponse{Content: ollamaResp.Message.Content, Raw: ollamaResp.Message.Content}, nil
}

func (p *Ollama) Review(ctx context.Context, req *ReviewRequest) (*ReviewResponse, error) {
	prompt := buildReviewPrompt(req)

	body := map[string]interface{}{
		"model":  p.model,
		"prompt": systemPrompt + "\n\n" + prompt,
		"stream": false,
		"format": "json",
		"options": map[string]interface{}{
			"temperature": 0.2,
			"num_predict": p.maxTokens,
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := p.baseURL + "/api/generate"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama API error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var ollamaResp struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode ollama response: %w", err)
	}

	return parseLLMResponse(ollamaResp.Response)
}
