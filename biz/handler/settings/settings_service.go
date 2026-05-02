package settings

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	settings "github.com/yi-nology/git-manage-service/biz/model/settings"
	"github.com/yi-nology/git-manage-service/biz/service/branchrule"
	"github.com/yi-nology/git-manage-service/biz/service/llm"
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
	dto := api.LLMProviderDTO{
		Name:      req.Name,
		Type:      req.Type,
		BaseURL:   req.BaseUrl,
		APIKey:    req.ApiKey,
		Model:     req.Model,
		MaxTokens: int(req.MaxTokens),
		IsDefault: req.IsDefault,
	}
	result, err := llm.CreateProvider(dto)
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
	dto := api.LLMProviderDTO{
		Name:      req.Name,
		Type:      req.Type,
		BaseURL:   req.BaseUrl,
		APIKey:    req.ApiKey,
		Model:     req.Model,
		MaxTokens: int(req.MaxTokens),
		IsDefault: req.IsDefault,
	}
	result, err := llm.UpdateProvider(uint(req.Id), dto)
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

func GetCodeReviewSettings(ctx context.Context, c *app.RequestContext) {
	var req settings.GetCodeReviewSettingsRequest
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	cfg := configs.GlobalConfig.CodeReview
	pkgresponse.Success(c, api.CodeReviewGlobalSettingsDTO{
		Enabled:        cfg.Enabled,
		AutoReviewOnMR: cfg.AutoReviewOnMR,
		BlockOnHigh:    cfg.BlockOnHigh,
		MaxFiles:       cfg.MaxFiles,
		MaxDiffLines:   cfg.MaxDiffLines,
	})
}

func UpdateCodeReviewSettings(ctx context.Context, c *app.RequestContext) {
	var req settings.UpdateCodeReviewSettingsRequest
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, err.Error())
		return
	}
	dto := api.CodeReviewGlobalSettingsDTO{
		Enabled:        req.Enabled,
		AutoReviewOnMR: req.AutoReviewOnMr,
		BlockOnHigh:    req.BlockOnHigh,
		MaxFiles:       int(req.MaxFiles),
		MaxDiffLines:   int(req.MaxDiffLines),
	}
	if dto.MaxFiles <= 0 {
		dto.MaxFiles = configs.GlobalConfig.CodeReview.MaxFiles
	}
	if dto.MaxDiffLines <= 0 {
		dto.MaxDiffLines = configs.GlobalConfig.CodeReview.MaxDiffLines
	}
	if err := settingssvc.SaveCodeReviewSettingsToDB(dto); err != nil {
		pkgresponse.InternalServerError(c, "failed to persist settings: "+err.Error())
		return
	}
	c.Set("audit_details", map[string]interface{}{"enabled": dto.Enabled, "auto_review": dto.AutoReviewOnMR})
	pkgresponse.Success(c, dto)
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
	dto := api.BranchRuleSetDTO{
		Enabled:   req.Enabled,
		Rules:     convertBranchRules(req.Rules),
		Protected: req.ProtectedBranches,
	}
	result, err := branchrule.UpdateGlobalRules(dto)
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	c.Set("audit_details", map[string]interface{}{"rules_count": len(dto.Rules)})
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
	dto := api.RemoteRepoBranchRulesDTO{
		ProviderConfigID: uint(req.ProviderId),
		PlatformOwner:    req.Owner,
		PlatformRepo:     req.Repo,
		UseCustomRules:   req.UseCustomRules,
		Rules:            convertBranchRules(req.Rules),
		Protected:        req.ProtectedBranches,
	}
	result, err := branchrule.UpdateRemoteRepoRules(uint(req.ProviderId), req.Owner, req.Repo, dto)
	if err != nil {
		pkgresponse.InternalServerError(c, err.Error())
		return
	}
	c.Set("audit_target", fmt.Sprintf("provider:%d:%s/%s", uint(req.ProviderId), req.Owner, req.Repo))
	pkgresponse.Success(c, result)
}

func convertBranchRules(rules []*settings.BranchRule) []api.BranchRuleDTO {
	if len(rules) == 0 {
		return nil
	}
	result := make([]api.BranchRuleDTO, 0, len(rules))
	for _, r := range rules {
		result = append(result, api.BranchRuleDTO{
			ID:                uint(r.Id),
			Prefix:            r.Prefix,
			DisplayName:       r.DisplayName,
			SourceBranches:    r.SourceBranches,
			TargetBranches:    r.TargetBranches,
			RequireTaskID:     r.RequireTaskId,
			TaskIDPattern:     r.TaskIdPattern,
			AutoDeleteOnMerge: r.AutoDeleteOnMerge,
			AllowDirectPush:   r.AllowDirectPush,
			RequireCodeReview: r.RequireCodeReview,
			SortOrder:         int(r.SortOrder),
		})
	}
	return result
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
