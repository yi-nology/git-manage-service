package git

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/service/llm"
)

type AuthorAIService struct{}

func NewAuthorAIService() *AuthorAIService {
	return &AuthorAIService{}
}

func (s *AuthorAIService) getProvider() (llm.Provider, error) {
	if !llm.HasDefaultProvider() {
		return nil, fmt.Errorf("未配置 LLM 提供商，请在系统设置中配置")
	}
	return llm.GetDefaultProvider()
}

func (s *AuthorAIService) chat(systemPrompt string, userPrompt string) (string, error) {
	p, err := s.getProvider()
	if err != nil {
		return "", err
	}
	resp, err := p.Chat(context.Background(), &llm.ChatRequest{
		SystemPrompt: systemPrompt,
		Messages:     []llm.ChatMessage{{Role: "user", Content: userPrompt}},
		MaxTokens:    4096,
	})
	if err != nil {
		return "", fmt.Errorf("LLM 调用失败: %w", err)
	}
	return resp.Content, nil
}

func (s *AuthorAIService) chatMulti(systemPrompt string, history []api.ChatMessageDTO, prompt string) (string, error) {
	p, err := s.getProvider()
	if err != nil {
		return "", err
	}
	var msgs []llm.ChatMessage
	for _, h := range history {
		msgs = append(msgs, llm.ChatMessage{Role: h.Role, Content: h.Content})
	}
	msgs = append(msgs, llm.ChatMessage{Role: "user", Content: prompt})
	resp, err := p.Chat(context.Background(), &llm.ChatRequest{
		SystemPrompt: systemPrompt,
		Messages:     msgs,
		MaxTokens:    4096,
	})
	if err != nil {
		return "", fmt.Errorf("LLM 调用失败: %w", err)
	}
	return resp.Content, nil
}

func extractJSON(raw string) string {
	start := strings.Index(raw, "{")
	if start == -1 {
		return raw
	}
	depth := 0
	for i := start; i < len(raw); i++ {
		if raw[i] == '{' {
			depth++
		} else if raw[i] == '}' {
			depth--
			if depth == 0 {
				return raw[start : i+1]
			}
		}
	}
	return raw[start:]
}

func (s *AuthorAIService) SmartSuggest(repoPath string) (*api.AliasSuggestionResult, error) {
	identities, err := db.NewAuthorIdentityDAO().ListAll()
	if err != nil {
		return nil, err
	}
	var existingIdentities []string
	for _, id := range identities {
		var aliases []string
		for _, a := range IdentityToDTO(&id).Aliases {
			aliases = append(aliases, fmt.Sprintf("%s <%s>", a.Name, a.Email))
		}
		existingIdentities = append(existingIdentities,
			fmt.Sprintf("身份 #%d: %s <%s> (别名: %s)", id.ID, id.CanonicalName, id.CanonicalEmail, strings.Join(aliases, ", ")))
	}

	cmd := exec.Command("git", "log", "--all", "--format=%an <%ae>")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}
	authorSet := make(map[string]bool)
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line != "" {
			authorSet[line] = true
		}
	}
	var allAuthors []string
	for a := range authorSet {
		allAuthors = append(allAuthors, a)
	}

	systemPrompt := `你是一个 Git 作者身份分析专家。用户会给你一个 git 仓库中所有提交作者列表和已有的身份配置。
请分析哪些作者实际上属于同一人（考虑：相似的用户名、同一邮箱的不同显示名、常见的命名模式如拼音缩写等）。
对于每个推荐，给出置信度(high/medium/low)和理由。

请以 JSON 格式返回，格式如下：
{
  "suggestions": [
    {
      "identityId": 1,
      "identityName": "身份的主名",
      "aliasName": "推荐的别名名称",
      "aliasEmail": "推荐的别名邮箱",
      "confidence": "high",
      "reason": "为什么认为这是同一人"
    }
  ],
  "summary": "总结分析结果，用中文描述"
}

注意：
- 只推荐当前不在已有别名中的作者
- identityId 必须是已有身份的真实 ID
- 不要将明显不同的人合并
- 如果没有推荐的别名，返回空数组`

	userPrompt := fmt.Sprintf("已有身份配置:\n%s\n\n仓库中的所有提交作者:\n%s",
		strings.Join(existingIdentities, "\n"),
		strings.Join(allAuthors, "\n"))

	raw, err := s.chat(systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	var result api.AliasSuggestionResult
	jsonStr := extractJSON(raw)
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return &api.AliasSuggestionResult{
			Suggestions: []api.AliasSuggestion{},
			Summary:     raw,
		}, nil
	}
	return &result, nil
}

func (s *AuthorAIService) AnalyzeScan(scan *api.AuthorScanResult, identities []api.AuthorIdentityDTO) (string, error) {
	systemPrompt := `你是一个 Git 仓库管理助手。用户会给你一次作者扫描的结果，包括不匹配的提交列表和身份配置。
请用简洁的中文总结分析结果，给出可操作的建议。格式：
1. 一句话总结
2. 按身份分组的问题描述
3. 建议的修复操作（选择修复 vs 全部修复，是否需要 push）
请直接输出中文文本，不要用 JSON。`

	var commits []string
	for _, c := range scan.Commits {
		commits = append(commits, fmt.Sprintf("- %s (%s): %s <%s> → %s <%s> [%s] %s",
			c.ShortHash, c.Date[:10], c.AuthorName, c.AuthorEmail,
			c.TargetName, c.TargetEmail, c.MatchType, c.Message))
	}
	var ids []string
	for _, id := range identities {
		ids = append(ids, fmt.Sprintf("- %s <%s> (默认: %v, %d 个别名)", id.CanonicalName, id.CanonicalEmail, id.IsDefault, len(id.Aliases)))
	}

	userPrompt := fmt.Sprintf("扫描结果: 共 %d 个提交, %d 个不匹配\n\n身份配置:\n%s\n\n不匹配的提交:\n%s",
		scan.TotalCommits, scan.MatchCount,
		strings.Join(ids, "\n"),
		strings.Join(commits, "\n"))

	return s.chat(systemPrompt, userPrompt)
}

func (s *AuthorAIService) SuggestMerges(identities []api.AuthorIdentityDTO) (*api.MergeSuggestionResult, error) {
	systemPrompt := `你是一个 Git 作者身份分析专家。用户会给你所有身份及其别名的列表。
请分析哪些身份实际上属于同一个人（考虑：相同的真实姓名、相似的邮箱模式、别名的交叉引用等）。

请以 JSON 格式返回：
{
  "merges": [
    {
      "keepId": 3,
      "keepName": "保留的身份名",
      "mergeIds": [1, 2],
      "mergeNames": "要合并的身份名",
      "reason": "为什么建议合并"
    }
  ],
  "summary": "总结分析结果"
}

注意：
- keepId 是建议保留的身份 ID，mergeIds 是建议合并进去的身份 ID
- 合并后 mergeIds 的别名会全部转移到 keepId
- 如果不需要合并，返回空数组
- 不要将明显不同的人合并`

	var ids []string
	for _, id := range identities {
		var aliases []string
		for _, a := range id.Aliases {
			aliases = append(aliases, fmt.Sprintf("%s <%s>", a.Name, a.Email))
		}
		ids = append(ids, fmt.Sprintf("身份 #%d: %s <%s> (默认: %v, 别名: %s)",
			id.ID, id.CanonicalName, id.CanonicalEmail, id.IsDefault, strings.Join(aliases, ", ")))
	}

	userPrompt := fmt.Sprintf("所有身份:\n%s", strings.Join(ids, "\n"))

	raw, err := s.chat(systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	var result api.MergeSuggestionResult
	jsonStr := extractJSON(raw)
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return &api.MergeSuggestionResult{
			Merges:  []api.MergeCandidate{},
			Summary: raw,
		}, nil
	}
	return &result, nil
}

func (s *AuthorAIService) AssessRisk(commits []api.MismatchedCommit, repoPath string) (*api.RiskAssessmentResult, error) {
	branches, _ := exec.Command("git", "branch", "-a").CombinedOutput()
	branchCount := len(strings.Split(strings.TrimSpace(string(branches)), "\n"))

	remotes, _ := exec.Command("git", "remote").CombinedOutput()
	remoteList := strings.TrimSpace(string(remotes))
	hasRemote := remoteList != ""

	systemPrompt := `你是一个 Git 仓库安全分析专家。用户即将执行 git filter-branch 重写提交历史。
请根据提供的信息评估风险等级，给出具体的风险因素和建议。

请以 JSON 格式返回：
{
  "riskLevel": "low|medium|high",
  "summary": "简短总结",
  "factors": [
    {"level": "warning", "description": "描述", "recommendation": "建议"}
  ],
  "recommendations": ["建议1", "建议2"]
}

考虑：
- 受影响的提交数量
- 是否有远程仓库（force push 风险）
- 分支数量（越多风险越高）
- 是否有协作者
- 是否涉及 tag`

	var commitInfos []string
	for _, c := range commits {
		commitInfos = append(commitInfos, fmt.Sprintf("- %s: %s <%s> → %s <%s>",
			c.ShortHash, c.AuthorName, c.AuthorEmail, c.TargetName, c.TargetEmail))
	}

	userPrompt := fmt.Sprintf("即将修改 %d 个提交:\n%s\n\n仓库信息:\n- 分支数: %d\n- 远程: %s\n- 远程列表: %s",
		len(commits),
		strings.Join(commitInfos, "\n"),
		branchCount,
		fmt.Sprintf("%v", hasRemote),
		remoteList)

	raw, err := s.chat(systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	var result api.RiskAssessmentResult
	jsonStr := extractJSON(raw)
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return &api.RiskAssessmentResult{
			RiskLevel: "medium",
			Summary:   raw,
		}, nil
	}
	return &result, nil
}

func (s *AuthorAIService) AuthorChat(repoPath string, prompt string, history []api.ChatMessageDTO, scan *api.AuthorScanResult) (string, error) {
	systemPrompt := `你是一个 Git 作者身份管理助手。帮助用户理解和管理 Git 提交作者信息。
你可以回答关于：
- 如何配置 Git 作者信息
- git filter-branch 的作用和风险
- 作者身份和别名的作用
- 修复历史提交的最佳实践
- git config 的使用方法

请用中文简洁回答。如果用户的问题与当前仓库的扫描结果相关，可以参考扫描数据给出针对性建议。`

	if scan != nil && scan.MatchCount > 0 {
		var commits []string
		for _, c := range scan.Commits {
			commits = append(commits, fmt.Sprintf("- %s: %s <%s>", c.ShortHash, c.AuthorName, c.AuthorEmail))
		}
		systemPrompt += fmt.Sprintf("\n\n当前仓库扫描结果: 共 %d 个提交, %d 个不匹配:\n%s",
			scan.TotalCommits, scan.MatchCount, strings.Join(commits, "\n"))
	}

	return s.chatMulti(systemPrompt, history, prompt)
}
