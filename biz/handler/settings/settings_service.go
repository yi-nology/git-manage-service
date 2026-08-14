package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	settings "github.com/yi-nology/git-manage-service/biz/model/settings"
	"github.com/yi-nology/git-manage-service/biz/service/branchrule"
	"github.com/yi-nology/git-manage-service/biz/service/llm"
	"github.com/yi-nology/git-manage-service/biz/service/rag"
	settingssvc "github.com/yi-nology/git-manage-service/biz/service/settings"
	"github.com/yi-nology/git-manage-service/pkg/configs"
	pkgresponse "github.com/yi-nology/git-manage-service/pkg/response"
)

func ListLLMProviders(ctx context.Context, c *app.RequestContext) {
	var req settings.ListLLMProvidersRequest
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	providers, err := llm.ListProviders()
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, providers)
}

func GetLLMProvider(ctx context.Context, c *app.RequestContext) {
	var req settings.GetLLMProviderRequest
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	provider, err := llm.GetProviderByID(uint(req.Id))
	if err != nil {
		pkgresponse.NotFound(c, err.Error())
		return
	}
	pkgresponse.Success(c, provider)
}

func CreateLLMProvider(ctx context.Context, c *app.RequestContext) {
	var req settings.CreateLLMProviderRequest
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if req.Name == "" || req.Type == "" || req.Model == "" {
		pkgresponse.BadRequest(c, "name, type and model are required")
		return
	}
	if req.BaseUrl == "" {
		pkgresponse.BadRequest(c, "base_url is required")
		return
	}
	providerInfo := &settings.LLMProviderInfo{
		Name:      req.Name,
		Type:      req.Type,
		BaseUrl:   req.BaseUrl,
		ApiKey:    req.ApiKey,
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
		IsDefault: req.IsDefault,
	}
	result, err := llm.CreateProvider(providerInfo)
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func UpdateLLMProvider(ctx context.Context, c *app.RequestContext) {
	var req settings.UpdateLLMProviderRequest
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	providerInfo := &settings.LLMProviderInfo{
		Name:      req.Name,
		Type:      req.Type,
		BaseUrl:   req.BaseUrl,
		ApiKey:    req.ApiKey,
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
		IsDefault: req.IsDefault,
	}
	result, err := llm.UpdateProvider(uint(req.Id), providerInfo)
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func DeleteLLMProvider(ctx context.Context, c *app.RequestContext) {
	var req settings.DeleteLLMProviderRequest
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if err := llm.DeleteProvider(uint(req.Id)); err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, map[string]string{"status": "deleted"})
}

func SetDefaultLLMProvider(ctx context.Context, c *app.RequestContext) {
	var req settings.SetDefaultLLMProviderRequest
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if err := llm.SetDefaultProvider(uint(req.Id)); err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, map[string]string{"status": "ok"})
}

func TestLLMProvider(ctx context.Context, c *app.RequestContext) {
	var req settings.TestLLMProviderRequest
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if err := llm.TestProvider(uint(req.Id)); err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, map[string]string{"status": "ok", "message": "连接测试成功"})
}

func TestEmbedding(ctx context.Context, c *app.RequestContext) {
	var req struct {
		ID uint `path:"id"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	provider, err := llm.GetProviderByID(req.ID)
	if err != nil {
		pkgresponse.NotFound(c, "provider not found: "+err.Error())
		return
	}
	model := provider.EmbeddingModel
	if model == "" {
		switch provider.Type {
		case "ollama":
			model = "nomic-embed-text"
		default:
			model = "text-embedding-3-small"
		}
	}
	client := rag.NewEmbeddingClient(provider.BaseUrl, provider.ApiKey, model, provider.Type)
	_, err = client.EmbedQuery(context.Background(), "Hello, world!")
	if err != nil {
		pkgresponse.InternalServerError(c, "Embedding test failed: "+err.Error())
		return
	}
	pkgresponse.Success(c, map[string]string{"status": "ok", "message": "Embedding 连接测试成功", "model": model})
}

func GetCodeReviewSettings(ctx context.Context, c *app.RequestContext) {
	var req settings.GetCodeReviewSettingsRequest
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	cfg := configs.GetCodeReviewConfig()
	pkgresponse.Success(c, &settings.CodeReviewSettings{
		Enabled:        cfg.Enabled,
		AutoReviewOnMr: cfg.AutoReviewOnMR,
		BlockOnHigh:    cfg.BlockOnHigh,
		MaxFiles:       int32(cfg.MaxFiles),
		MaxDiffLines:   int32(cfg.MaxDiffLines),
		RagEnabled:     cfg.RAG.Enabled,
	})
}

func UpdateCodeReviewSettings(ctx context.Context, c *app.RequestContext) {
	var req settings.UpdateCodeReviewSettingsRequest
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	maxFiles := int(req.MaxFiles)
	maxDiffLines := int(req.MaxDiffLines)
	if maxFiles <= 0 {
		maxFiles = configs.GetCodeReviewConfig().MaxFiles
	}
	if maxDiffLines <= 0 {
		maxDiffLines = configs.GetCodeReviewConfig().MaxDiffLines
	}
	dto := settingssvc.CodeReviewSettings{
		Enabled:        req.Enabled,
		AutoReviewOnMR: req.AutoReviewOnMr,
		BlockOnHigh:    req.BlockOnHigh,
		MaxFiles:       maxFiles,
		MaxDiffLines:   maxDiffLines,
	}
	if err := settingssvc.SaveCodeReviewSettingsToDB(dto); err != nil {
		pkgresponse.InternalServerError(c, "failed to persist settings: "+err.Error())
		return
	}
	c.Set("audit_details", map[string]interface{}{"enabled": dto.Enabled, "auto_review": dto.AutoReviewOnMR})
	pkgresponse.Success(c, &settings.CodeReviewSettings{
		Enabled:        dto.Enabled,
		AutoReviewOnMr: dto.AutoReviewOnMR,
		BlockOnHigh:    dto.BlockOnHigh,
		MaxFiles:       int32(dto.MaxFiles),
		MaxDiffLines:   int32(dto.MaxDiffLines),
		RagEnabled:     dto.RAGEnabled,
	})
}

func GetBranchRules(ctx context.Context, c *app.RequestContext) {
	var req settings.GetBranchRulesRequest
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	result, err := branchrule.GetGlobalRules()
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func UpdateBranchRules(ctx context.Context, c *app.RequestContext) {
	var req settings.UpdateBranchRulesRequest
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	protoReq := &settings.BranchRuleSet{
		Enabled:           req.Enabled,
		Rules:             req.Rules,
		ProtectedBranches: req.ProtectedBranches,
	}
	result, err := branchrule.UpdateGlobalRules(protoReq)
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	c.Set("audit_details", map[string]interface{}{"rules_count": len(protoReq.Rules)})
	pkgresponse.Success(c, result)
}

func ValidateBranchName(ctx context.Context, c *app.RequestContext) {
	var req settings.ValidateBranchNameRequest
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if req.RepoKey == "" || req.BranchName == "" {
		pkgresponse.BadRequest(c, "repo_key and branch_name are required")
		return
	}
	result, err := branchrule.ValidateBranchName(req.RepoKey, req.BranchName, req.BaseRef, req.SkipRules)
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func GetRemoteRepoBranchRules(ctx context.Context, c *app.RequestContext) {
	var req settings.GetRemoteRepoBranchRulesRequest
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	result, err := branchrule.GetRemoteRepoRules(uint(req.ProviderId), req.Owner, req.Repo)
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func UpdateRemoteRepoBranchRules(ctx context.Context, c *app.RequestContext) {
	var req settings.UpdateRemoteRepoBranchRulesRequest
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	protoReq := &settings.RemoteRepoBranchRuleSet{
		ProviderConfigId:  req.ProviderId,
		PlatformOwner:     req.Owner,
		PlatformRepo:      req.Repo,
		UseCustomRules:    req.UseCustomRules,
		Rules:             req.Rules,
		ProtectedBranches: req.ProtectedBranches,
	}
	result, err := branchrule.UpdateRemoteRepoRules(uint(req.ProviderId), req.Owner, req.Repo, protoReq)
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	c.Set("audit_target", fmt.Sprintf("provider:%d:%s/%s", uint(req.ProviderId), req.Owner, req.Repo))
	pkgresponse.Success(c, result)
}

func FetchOllamaModels(ctx context.Context, c *app.RequestContext) {
	baseURL := c.Query("base_url")
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	url := strings.TrimRight(baseURL, "/") + "/api/tags"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		pkgresponse.BadRequest(c, "invalid base_url: "+err.Error())
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		pkgresponse.BadRequest(c, "无法连接 Ollama: "+err.Error())
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		pkgresponse.InternalServerError(c, "解析 Ollama 响应失败")
		return
	}
	names := make([]string, 0, len(result.Models))
	for _, m := range result.Models {
		names = append(names, m.Name)
	}
	pkgresponse.Success(c, names)
}

func ListReviewRules(ctx context.Context, c *app.RequestContext) {
	rules, err := settingssvc.ListReviewRules()
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, rules)
}

func GetReviewRule(ctx context.Context, c *app.RequestContext) {
	ruleID := c.Param("rule_id")
	if ruleID == "" {
		pkgresponse.BadRequest(c, "rule_id is required")
		return
	}
	rule, err := settingssvc.GetReviewRule(ruleID)
	if err != nil {
		pkgresponse.NotFound(c, err.Error())
		return
	}
	pkgresponse.Success(c, rule)
}

func CreateReviewRule(ctx context.Context, c *app.RequestContext) {
	var dto settingssvc.ReviewRuleDTO
	body, err := c.Body()
	if err != nil {
		pkgresponse.BadRequest(c, "failed to read body: "+err.Error())
		return
	}
	if err := json.Unmarshal(body, &dto); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	result, err := settingssvc.CreateReviewRule(dto)
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	c.Set("audit_target", "review_rule:"+result.ID)
	pkgresponse.Success(c, result)
}

func UpdateReviewRule(ctx context.Context, c *app.RequestContext) {
	ruleID := c.Param("rule_id")
	if ruleID == "" {
		pkgresponse.BadRequest(c, "rule_id is required")
		return
	}
	var dto settingssvc.ReviewRuleDTO
	body, err := c.Body()
	if err != nil {
		pkgresponse.BadRequest(c, "failed to read body: "+err.Error())
		return
	}
	if err := json.Unmarshal(body, &dto); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	dto.ID = ruleID
	result, err := settingssvc.UpdateReviewRule(ruleID, dto)
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func DeleteReviewRule(ctx context.Context, c *app.RequestContext) {
	ruleID := c.Param("rule_id")
	if ruleID == "" {
		pkgresponse.BadRequest(c, "rule_id is required")
		return
	}
	if err := settingssvc.DeleteReviewRule(ruleID); err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	pkgresponse.Success(c, map[string]string{"status": "deleted"})
}

func BatchUpdateReviewRules(ctx context.Context, c *app.RequestContext) {
	var dtos []settingssvc.ReviewRuleDTO
	body, err := c.Body()
	if err != nil {
		pkgresponse.BadRequest(c, "failed to read body: "+err.Error())
		return
	}
	if err := json.Unmarshal(body, &dtos); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	if err := settingssvc.BatchUpdateReviewRules(dtos); err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	c.Set("audit_details", map[string]interface{}{"count": len(dtos)})
	rules, _ := settingssvc.ListReviewRules()
	pkgresponse.Success(c, rules)
}
