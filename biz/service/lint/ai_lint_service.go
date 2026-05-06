package lint

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/yi-nology/git-manage-service/biz/service/ai"
	"github.com/yi-nology/git-manage-service/biz/service/llm"
)

const aiLintSystemPrompt = `你是一个 RPM Spec 文件审查专家。用户会给你一个 .spec 文件的完整内容。
请分析这个文件，找出以下类型的语义问题（规则引擎无法检测到的）：

1. **依赖完整性**: BuildRequires 和 Requires 是否覆盖了编译和运行时所需的所有依赖
2. **%files 与 %install 一致性**: %files 段列出的文件路径是否在 %install 段中有对应的安装操作
3. **宏使用正确性**: 是否正确使用了 %configure、%make_build、%make_install 等推荐宏
4. **Source URL 有效性**: Source0 等源码 URL 是否是有效的模式
5. **许可证标识**: License 字段是否使用了 SPDX 标准标识符
6. **安全性问题**: %post/%postun 等脚本段是否存在危险操作
7. **构建流程逻辑**: %build 段的编译命令是否完整正确
8. **版本号规范**: Version/Release 字段是否符合 RPM 版本规范
9. **补丁管理**: Patch 字段声明但未在 %prep 中应用，或反之

请用中文回答。响应格式必须是 JSON：
{
  "issues": [
    {
      "line": 行号(整数，不确定时为0),
      "severity": "error|warning|info",
      "message": "问题描述",
      "quick_fix": "修复建议的具体操作描述，如可能则给出可替换的代码片段"
    }
  ]
}

注意：
- 只报告你确定存在的问题，不要猜测
- severity 标准：error=必须修复的严重问题，warning=强烈建议修复，info=改进建议
- 每个 issue 必须有明确的 quick_fix 描述
- 不要报告规则引擎已覆盖的简单缺失字段问题（如 Name/Version 等）`

func AILint(ctx context.Context, content string, mode string) (*LintResult, error) {
	resp, err := ai.NewRunner().Chat(ctx, ai.TaskRequest{
		Type:          ai.TaskSpecLint,
		PromptVersion: "spec-lint.v1",
		SystemPrompt:  aiLintSystemPrompt,
		Messages: []llm.ChatMessage{
			{Role: "user", Content: "请审查以下 spec 文件：\n\n" + content},
		},
		MaxTokens: 4096,
	})
	if err != nil {
		return nil, fmt.Errorf("AI lint request failed: %w", err)
	}

	return parseAILintResponse(resp.Content)
}

func parseAILintResponse(raw string) (*LintResult, error) {
	result := &LintResult{
		Issues: []LintIssue{},
		Stats:  LintStats{},
	}

	jsonStr := ai.ExtractJSON(raw)
	if jsonStr == "" {
		return result, nil
	}

	var parsed struct {
		Issues []struct {
			Line     int    `json:"line"`
			Severity string `json:"severity"`
			Message  string `json:"message"`
			QuickFix string `json:"quick_fix"`
		} `json:"issues"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return result, nil
	}

	for _, issue := range parsed.Issues {
		severity := issue.Severity
		if severity != "error" && severity != "warning" && severity != "info" {
			severity = "info"
		}
		li := LintIssue{
			RuleID:   "ai-lint",
			Severity: severity,
			Message:  issue.Message,
			Line:     issue.Line,
			Source:   "ai",
			QuickFix: issue.QuickFix,
		}
		result.Issues = append(result.Issues, li)
		switch severity {
		case "error":
			result.Stats.ErrorCount++
		case "warning":
			result.Stats.WarningCount++
		case "info":
			result.Stats.InfoCount++
		}
	}

	return result, nil
}

func AIFix(ctx context.Context, content string, issue string, line int, severity string) (string, error) {
	prompt := fmt.Sprintf("在以下 spec 文件中，有一个问题需要修复：\n\n问题：%s\n行号：%d\n严重级别：%s\n\n请输出修复后的完整 spec 文件内容。只修改与问题相关的部分，保持其余内容完全不变。不要用代码块包裹输出。", issue, line, severity)

	resp, err := ai.NewRunner().Chat(ctx, ai.TaskRequest{
		Type:          ai.TaskSpecFix,
		PromptVersion: "spec-fix.v1",
		SystemPrompt:  "你是一个 RPM Spec 文件专家。用户会给你一个有问题的 spec 文件和问题描述，请修复问题后输出完整的 spec 文件内容。只修改有问题的部分，保持其余内容不变。直接输出文件内容，不要加任何解释或代码块标记。",
		Messages: []llm.ChatMessage{
			{Role: "user", Content: content + "\n\n---\n" + prompt},
		},
		MaxTokens: 8192,
	})
	if err != nil {
		return "", fmt.Errorf("AI fix request failed: %w", err)
	}

	return ai.StripFencedCode(resp.Content, "spec", "rpm"), nil
}

func extractJSON(text string) string {
	return ai.ExtractJSON(text)
}
