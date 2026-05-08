package ai

import (
	"context"
	"fmt"
	"time"

	"github.com/yi-nology/git-manage-service/biz/service/llm"
)

const defaultTimeout = 60 * time.Second
const defaultMaxInputChars = 60000

type Runner struct {
	resolver *ProviderResolver
}

func NewRunner() *Runner {
	return &Runner{resolver: NewProviderResolver()}
}

func NewRunnerWithResolver(resolver *ProviderResolver) *Runner {
	if resolver == nil {
		resolver = NewProviderResolver()
	}
	return &Runner{resolver: resolver}
}

func (r *Runner) Chat(ctx context.Context, req TaskRequest) (*TaskResponse, error) {
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("AI task %s has no messages", req.Type)
	}
	started := time.Now()

	systemPrompt, messages := prepareInput(req)
	inputSnapshot := serializeInput(systemPrompt, messages)

	provider, err := r.resolver.Resolve(req.Provider)
	if err != nil {
		recordInvocation(req, "", inputSnapshot, "", "failed", err.Error(), started)
		return nil, fmt.Errorf("resolve provider for AI task %s: %w", req.Type, err)
	}

	timeout := req.Timeout
	if timeout == 0 {
		if spec := GetTaskConfig(req.Type); spec.Timeout > 0 {
			timeout = spec.Timeout
		} else {
			timeout = defaultTimeout
		}
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	resp, err := provider.Chat(callCtx, &llm.ChatRequest{
		SystemPrompt: systemPrompt,
		Messages:     messages,
		MaxTokens:    req.MaxTokens,
	})
	if err != nil {
		invID := recordInvocation(req, provider.Name(), inputSnapshot, "", "failed", err.Error(), started)
		_ = invID
		return nil, fmt.Errorf("run AI task %s with provider %s: %w", req.Type, provider.Name(), err)
	}
	invID := recordInvocation(req, provider.Name(), inputSnapshot, resp.Content, "success", "", started)
	return &TaskResponse{
		Content:       resp.Content,
		Raw:           resp.Raw,
		ProviderName:  provider.Name(),
		PromptVersion: req.PromptVersion,
		TaskType:      req.Type,
		InvocationID:  invID,
	}, nil
}

func prepareInput(req TaskRequest) (string, []llm.ChatMessage) {
	maxChars := req.MaxInputChars
	if maxChars == 0 {
		maxChars = defaultMaxInputChars
	}

	systemPrompt := RedactSecrets(req.SystemPrompt)
	messages := make([]llm.ChatMessage, 0, len(req.Messages))
	for _, msg := range req.Messages {
		messages = append(messages, llm.ChatMessage{
			Role:    msg.Role,
			Content: RedactSecrets(msg.Content),
		})
	}

	total := len(systemPrompt)
	for _, msg := range messages {
		total += len(msg.Content)
	}
	if maxChars <= 0 || total <= maxChars {
		return systemPrompt, messages
	}

	remaining := maxChars - len(systemPrompt)
	if remaining < maxChars/3 {
		systemPrompt = ClampText(systemPrompt, maxChars/3)
		remaining = maxChars - len(systemPrompt)
	}
	if remaining <= 0 {
		return systemPrompt, nil
	}

	perMessage := remaining / len(messages)
	if perMessage < 1000 {
		perMessage = 1000
	}
	for i := range messages {
		messages[i].Content = ClampText(messages[i].Content, perMessage)
	}
	return systemPrompt, messages
}
