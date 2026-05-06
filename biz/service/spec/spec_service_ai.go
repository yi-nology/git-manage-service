package spec

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/service/ai"
	"github.com/yi-nology/git-manage-service/biz/service/llm"
)

type ChatMessage = llm.ChatMessage

func (s *SpecService) AIAssist(ctx context.Context, content string, prompt string, action string, history []llm.ChatMessage) (string, string, error) {
	var systemPrompt string
	taskType := ai.TaskSpecChat
	if action == "agent" {
		taskType = ai.TaskSpecPatch
		systemPrompt = `你是一个 RPM Spec 文件编辑 Agent。用户的 spec 文件内容如下：

` + "```" + `
` + content + `
` + "```" + `

你的任务是直接修改这个 spec 文件。

规则：
1. 只需要输出修改后的完整 spec 文件内容，不要有任何其他文字、解释或代码块标记。
2. 第一行必须是 Name: 开头。
3. 严格遵循 RPM 打包最佳实践。
4. 保持原有的格式风格和缩进。
5. 不要删除或修改用户没有要求改动的部分。`
	} else {
		systemPrompt = `你是一个 RPM Spec 文件专家。用户会给你一个 .spec 文件的内容和一个问题或请求。
请用中文回答，提供专业、准确的建议或修改方案。

规则：
1. 如果是修改请求，请直接输出修改后的完整 spec 文件内容，不要用代码块包裹。
2. 如果是检查或分析请求，请用清晰的列表格式列出问题。
3. 始终遵循 RPM 打包最佳实践。
4. 用户当前正在编辑的 spec 文件内容如下：` + "\n```\n" + content + "\n```"
	}

	var chatMsgs []llm.ChatMessage
	if len(history) > 0 {
		maxHistory := 10
		if len(history) > maxHistory {
			history = history[len(history)-maxHistory:]
		}
		chatMsgs = append(chatMsgs, history...)
	}
	chatMsgs = append(chatMsgs, llm.ChatMessage{Role: "user", Content: prompt})

	resp, err := ai.NewRunner().Chat(ctx, ai.TaskRequest{
		Type:          taskType,
		PromptVersion: "spec-assist.v1",
		SystemPrompt:  systemPrompt,
		Messages:      chatMsgs,
		MaxTokens:     4096,
	})
	if err != nil {
		return "", "", fmt.Errorf("LLM request failed: %w", err)
	}

	applyContent := ""
	if action == "complete" || action == "generate" || action == "agent" {
		applyContent = extractSpecContent(resp.Content)
	}

	if action == "agent" && applyContent != "" {
		return "", applyContent, nil
	}

	return resp.Content, applyContent, nil
}

func extractSpecContent(text string) string {
	codeBlockRe := regexp.MustCompile("(?s)```(?:spec|rpm)?\\n(.*?)```")
	if matches := codeBlockRe.FindStringSubmatch(text); len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}
	lines := strings.Split(text, "\n")
	var specLines []string
	inSpec := false
	for _, line := range lines {
		stripped := strings.TrimSpace(line)
		if strings.HasPrefix(stripped, "Name:") ||
			strings.HasPrefix(stripped, "Version:") ||
			strings.HasPrefix(stripped, "Release:") ||
			strings.HasPrefix(stripped, "Summary:") ||
			strings.HasPrefix(stripped, "%prep") ||
			strings.HasPrefix(stripped, "%build") ||
			strings.HasPrefix(stripped, "%install") ||
			strings.HasPrefix(stripped, "%clean") ||
			strings.HasPrefix(stripped, "%files") ||
			strings.HasPrefix(stripped, "%changelog") ||
			strings.HasPrefix(stripped, "%description") {
			inSpec = true
		}
		if inSpec {
			specLines = append(specLines, line)
		}
	}
	if len(specLines) > 3 {
		return strings.Join(specLines, "\n")
	}
	return ""
}
