package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/oekit/pkg/httpjson"
	"net/http"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/pkg/httputil"
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
		// Ollama 是用户自托管服务：只允许本机/内网地址，防止配置被当成
		// SSRF 跳板（详细约束见 httputil 包注释）。
		httpClient: httputil.NewLocalServiceClient(300 * time.Second),
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
	// 迁 oekit/httpjson 单源（issue #18）：手工构造/判状态码/解码三段式
	// 与 httpjson.DoJSON 的纪律单源重复；SSRF 防护用的本地客户端原样
	// 传入，防护不降级。
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	var ollamaResp struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := httpjson.DoJSON(ctx, p.httpClient, httpjson.Request{
		Method: "POST",
		URL:    p.baseURL + "/api/chat",
		Body:   jsonBody,
		Header: func(h http.Header) { h.Set("Content-Type", "application/json") },
	}, &ollamaResp); err != nil {
		return nil, fmt.Errorf("ollama chat error: %w", err)
	}
	return &ChatResponse{Content: ollamaResp.Message.Content, Raw: ollamaResp.Message.Content}, nil
}
