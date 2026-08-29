package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/pkg/configs"
)

type EmbeddingClient struct {
	baseURL      string
	apiKey       string
	model        string
	providerType string
	httpClient   *http.Client
}

func NewEmbeddingClient(baseURL, apiKey, model, providerType string) *EmbeddingClient {
	if model == "" {
		switch providerType {
		case "ollama":
			model = "nomic-embed-text"
		default:
			model = "text-embedding-3-small"
		}
	}
	url := strings.TrimRight(baseURL, "/")
	if providerType != "ollama" && !strings.HasSuffix(url, "/v1") {
		url += "/v1"
	}
	return &EmbeddingClient{
		baseURL:      url,
		apiKey:       apiKey,
		model:        model,
		providerType: providerType,
		httpClient:   &http.Client{Timeout: 120 * time.Second},
	}
}

func NewEmbeddingClientFromDB() *EmbeddingClient {
	provider, err := db.NewLLMProviderDAO().FindEmbeddingProvider()
	if err == nil && provider != nil {
		model := provider.EmbeddingModel
		if model == "" {
			switch provider.Type {
			case "ollama":
				model = "nomic-embed-text"
			default:
				model = "text-embedding-3-small"
			}
		}
		log.Printf("[RAG] Using embedding provider: %s (type=%s, model=%s)", provider.Name, provider.Type, model)
		return NewEmbeddingClient(provider.BaseURL, provider.APIKey, model, provider.Type)
	}

	return NewEmbeddingClientFromConfig()
}

func NewEmbeddingClientFromConfig() *EmbeddingClient {
	cfg := configs.GetCodeReviewConfig()
	if cfg.DefaultLLM == "" {
		return nil
	}
	for _, p := range cfg.LLMProviders {
		if p.Name == cfg.DefaultLLM {
			return NewEmbeddingClient(p.BaseURL, p.APIKey, "text-embedding-3-small", p.Type)
		}
	}
	return nil
}

func (c *EmbeddingClient) Embed(ctx context.Context, texts []string) ([][]float64, error) {
	if c == nil || c.baseURL == "" {
		return nil, fmt.Errorf("embedding client not configured")
	}
	if len(texts) == 0 {
		return nil, nil
	}

	if c.providerType == "ollama" {
		return c.embedOllama(ctx, texts)
	}
	return c.embedOpenAI(ctx, texts)
}

func (c *EmbeddingClient) embedOpenAI(ctx context.Context, texts []string) ([][]float64, error) {
	batchSize := 20
	var allEmbeddings [][]float64

	for i := 0; i < len(texts); i += batchSize {
		end := i + batchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[i:end]

		embeddings, err := c.embedOpenAIBatch(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("openai embedding batch %d-%d: %w", i, end, err)
		}
		allEmbeddings = append(allEmbeddings, embeddings...)
	}
	return allEmbeddings, nil
}

type openAIEmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type openAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

func (c *EmbeddingClient) embedOpenAIBatch(ctx context.Context, texts []string) ([][]float64, error) {
	reqBody := openAIEmbeddingRequest{Model: c.model, Input: texts}
	jsonBody, _ := json.Marshal(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/embeddings", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai embedding API call failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read openai embedding response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("openai embedding API returned %d: %s", resp.StatusCode, truncateString(string(body), 500))
	}

	var embResp openAIEmbeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("parse openai embedding response: %w", err)
	}

	result := make([][]float64, len(texts))
	for _, d := range embResp.Data {
		if d.Index < len(result) {
			result[d.Index] = d.Embedding
		}
	}
	for i, r := range result {
		if r == nil {
			return nil, fmt.Errorf("missing embedding for index %d", i)
		}
	}
	return result, nil
}

type ollamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
}

func (c *EmbeddingClient) embedOllama(ctx context.Context, texts []string) ([][]float64, error) {
	url := strings.TrimRight(c.baseURL, "/")
	if !strings.HasSuffix(url, "/api") {
		url += "/api"
	}

	allEmbeddings := make([][]float64, 0, len(texts))
	for i, text := range texts {
		reqBody := ollamaEmbeddingRequest{Model: c.model, Prompt: text}
		jsonBody, _ := json.Marshal(reqBody)

		req, err := http.NewRequestWithContext(ctx, "POST", url+"/embeddings", bytes.NewReader(jsonBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("ollama embedding %d: %w", i, err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("ollama embedding returned %d: %s", resp.StatusCode, truncateString(string(body), 300))
		}

		var embResp ollamaEmbeddingResponse
		if err := json.Unmarshal(body, &embResp); err != nil {
			return nil, fmt.Errorf("parse ollama embedding: %w", err)
		}
		if len(embResp.Embedding) == 0 {
			return nil, fmt.Errorf("ollama returned empty embedding for text %d", i)
		}
		allEmbeddings = append(allEmbeddings, embResp.Embedding)
	}
	return allEmbeddings, nil
}

func (c *EmbeddingClient) EmbedQuery(ctx context.Context, query string) ([]float64, error) {
	embeddings, err := c.Embed(ctx, []string{query})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("no embedding returned for query")
	}
	return embeddings[0], nil
}

func (c *EmbeddingClient) ProviderType() string {
	if c == nil {
		return ""
	}
	return c.providerType
}

func (c *EmbeddingClient) Model() string {
	if c == nil {
		return ""
	}
	return c.model
}

func truncateString(s string, maxLen int) string {
	return s[:min(len(s), maxLen)]
}

func init() {
	log.Printf("[RAG] Vector store initialized")
}
