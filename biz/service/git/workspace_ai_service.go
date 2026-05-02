package git

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/model/api"
)

type WorkspaceAIService struct{}

func NewWorkspaceAIService() *WorkspaceAIService {
	return &WorkspaceAIService{}
}

const conflictResolveSystemPrompt = `你是一个 Git 冲突解决专家。你需要分析代码冲突并给出最佳合并方案。

规则：
1. 保留两边的有效修改，优先保留更新的功能代码
2. 如果两边修改了同一逻辑，优先保留 theirs（远程）的版本，除非 ours 有明确更好的实现
3. 确保合并后的代码语法正确、逻辑完整
4. 不要遗漏任何一方的功能性修改
5. 返回纯合并后的文件内容，不要包含冲突标记

你必须以 JSON 格式返回：
{
  "resolved_content": "合并后的完整文件内容",
  "explanation": "简要说明合并策略和决策理由",
  "confidence": 0.95
}`

func (s *WorkspaceAIService) AIResolveConflict(ours, theirs, base, filePath, hint string) (*api.AIResolvedFile, error) {
	authorAI := NewAuthorAIService()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("文件路径: %s\n\n", filePath))
	sb.WriteString("=== BASE (共同祖先) ===\n")
	sb.WriteString(base)
	sb.WriteString("\n\n=== OURS (本地当前) ===\n")
	sb.WriteString(ours)
	sb.WriteString("\n\n=== THEIRS (远程变更) ===\n")
	sb.WriteString(theirs)
	if hint != "" {
		sb.WriteString(fmt.Sprintf("\n\n额外提示: %s", hint))
	}

	result, err := authorAI.chat(conflictResolveSystemPrompt, sb.String())
	if err != nil {
		return nil, fmt.Errorf("AI conflict resolve: %w", err)
	}

	jsonStr := extractJSON(result)
	var parsed struct {
		ResolvedContent string  `json:"resolved_content"`
		Explanation     string  `json:"explanation"`
		Confidence      float64 `json:"confidence"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		return &api.AIResolvedFile{
			FilePath:        filePath,
			ResolvedContent: result,
			Explanation:     "AI 返回了非标准格式，已直接使用返回内容",
			Confidence:      0.5,
		}, nil
	}

	return &api.AIResolvedFile{
		FilePath:        filePath,
		ResolvedContent: parsed.ResolvedContent,
		Explanation:     parsed.Explanation,
		Confidence:      parsed.Confidence,
	}, nil
}
