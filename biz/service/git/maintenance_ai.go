package git

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/service/llm"
)

type MaintenanceAIService struct{}

func NewMaintenanceAIService() *MaintenanceAIService {
	return &MaintenanceAIService{}
}

func (s *MaintenanceAIService) AnalyzeSlimFiles(ctx context.Context, healthReport *api.RepoHealthReport) (*api.MaintenanceAIAnalysisResponse, error) {
	if !llm.HasDefaultProvider() {
		return nil, fmt.Errorf("未配置 LLM 提供商，请在系统设置中配置")
	}
	provider, err := llm.GetDefaultProvider()
	if err != nil {
		return nil, err
	}

	userPrompt := buildSlimAnalysisPrompt(healthReport)

	systemPrompt := `你是一个专业的 Git 仓库瘦身顾问。用户会提供仓库的健康报告和大文件列表，你需要分析每个文件并给出专业建议。

请严格按照以下 JSON 格式返回结果，不要返回其他任何内容：
{
  "summary": "整体分析总结（中文，2-3句话概括仓库瘦身建议）",
  "totalSavings": "预计可释放的空间（如 45.2 MB）",
  "totalSaveBytes": 预计可释放的字节数,
  "recommendations": [
    {
      "path": "文件路径",
      "size": "文件大小",
      "sizeBytes": 文件字节数,
      "recommendation": "safe_to_delete 或 caution 或 keep",
      "category": "binary 或 build_artifact 或 dependency 或 media 或 test_data 或 docs 或 source 或 config",
      "reason": "推荐理由（中文，说明为什么建议删除/保留/谨慎处理）",
      "confidence": "high 或 medium 或 low"
    }
  ]
}

判断标准：
- safe_to_delete: 编译产物(.exe/.bin/.o/.a/.so)、构建缓存、测试二进制、未使用的依赖、临时文件
- caution: 大型媒体文件(可能项目需要)、docs 目录中的大文件、配置文件(可能有特殊用途)
- keep: 源代码文件、重要的配置文件、项目必需的资源

confidence:
- high: 非常确定是可删除的编译产物或临时文件
- medium: 大概率可删除，但需要用户确认
- low: 不确定，建议用户自行判断`

	resp, err := provider.Chat(ctx, &llm.ChatRequest{
		SystemPrompt: systemPrompt,
		Messages:     []llm.ChatMessage{{Role: "user", Content: userPrompt}},
		MaxTokens:    4096,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM 调用失败: %w", err)
	}

	content := resp.Content
	jsonStr := extractJSONFromAI(content)

	var result api.MaintenanceAIAnalysisResponse
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return &api.MaintenanceAIAnalysisResponse{
			Summary:         content,
			Recommendations: []api.FileAIRecommendation{},
		}, nil
	}

	return &result, nil
}

func buildSlimAnalysisPrompt(report *api.RepoHealthReport) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## 仓库健康报告\n"))
	sb.WriteString(fmt.Sprintf("- .git 目录大小: %s\n", report.GitDirSize))
	sb.WriteString(fmt.Sprintf("- 提交数: %d\n", report.CommitCount))
	sb.WriteString(fmt.Sprintf("- 分支数: %d\n", report.BranchCount))
	sb.WriteString(fmt.Sprintf("- 松散对象: %d\n", report.LooseObjects))
	sb.WriteString(fmt.Sprintf("- Pack 文件数: %d\n", report.PackFiles))
	sb.WriteString(fmt.Sprintf("- Pack 内对象数: %d\n", report.InPackObjects))

	if report.GitDirBreakdown != nil {
		sb.WriteString(fmt.Sprintf("\n### .git 空间明细\n"))
		sb.WriteString(fmt.Sprintf("- Pack 文件: %s\n", report.GitDirBreakdown.PackDirSize))
		sb.WriteString(fmt.Sprintf("- 松散对象: %s\n", report.GitDirBreakdown.LooseObjSize))
		sb.WriteString(fmt.Sprintf("- Reflog 日志: %s\n", report.GitDirBreakdown.ReflogSize))
		sb.WriteString(fmt.Sprintf("- 其他: %s\n", report.GitDirBreakdown.OtherSize))
	}

	if len(report.LargeFiles) > 0 {
		sb.WriteString(fmt.Sprintf("\n### 大文件列表 (阈值: %s, 共 %d 个)\n\n", report.ThresholdHuman, len(report.LargeFiles)))
		sb.WriteString("| 路径 | 大小 | 大小(bytes) | 存在 | 来源 | 涉及提交数 |\n")
		sb.WriteString("|------|------|-------------|------|------|------------|\n")
		for _, f := range report.LargeFiles {
			exists := "否"
			if f.Exists {
				exists = "是"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %d | %s | %s | %d |\n", f.Path, f.Size, f.SizeBytes, exists, f.Source, f.CommitCount))
		}
	} else {
		sb.WriteString("\n未发现超过阈值的大文件。\n")
	}

	return sb.String()
}

func extractJSONFromAI(content string) string {
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		return content[start : end+1]
	}
	return content
}
