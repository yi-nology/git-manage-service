package branchrule

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func GetGlobalRules() (*api.BranchRuleSetDTO, error) {
	dao := db.NewBranchRuleSetDAO()
	set, err := dao.FindGlobal()
	if err != nil {
		return getDefaultRuleSetDTO(), nil
	}
	return ruleSetToDTO(set), nil
}

func UpdateGlobalRules(req api.BranchRuleSetDTO) (*api.BranchRuleSetDTO, error) {
	dao := db.NewBranchRuleSetDAO()
	set := &po.BranchRuleSet{
		ScopeType: "global",
		ScopeID:   "default",
		Enabled:   req.Enabled,
	}
	rules := dtoToRules(req.Rules)
	protected := req.Protected
	if protected == nil {
		protected = []string{}
	}
	set.SetRules(rules)
	set.SetProtected(protected)
	if err := dao.Upsert(set); err != nil {
		return nil, fmt.Errorf("failed to save global rules: %w", err)
	}
	return ruleSetToDTO(set), nil
}

func GetRemoteRepoRules(providerConfigID uint, platformOwner, platformRepo string) (*api.RemoteRepoBranchRulesDTO, error) {
	overrideDAO := db.NewBranchRuleOverrideDAO()
	override, err := overrideDAO.FindByRemoteRepo(providerConfigID, platformOwner, platformRepo)

	result := &api.RemoteRepoBranchRulesDTO{
		ProviderConfigID: providerConfigID,
		PlatformOwner:    platformOwner,
		PlatformRepo:     platformRepo,
		UseCustomRules:   false,
	}

	if err == nil {
		result.UseCustomRules = override.UseCustomRules
		if override.UseCustomRules {
			result.Rules = rulesToDTOs(override.GetRules())
			result.Protected = override.GetProtected()
		} else {
			globalDTO, _ := GetGlobalRules()
			result.Rules = globalDTO.Rules
			result.Protected = globalDTO.Protected
		}
	} else {
		globalDTO, _ := GetGlobalRules()
		result.Rules = globalDTO.Rules
		result.Protected = globalDTO.Protected
	}

	result.LinkedRepos = findLinkedRepos(providerConfigID, platformOwner, platformRepo)
	return result, nil
}

func UpdateRemoteRepoRules(providerConfigID uint, platformOwner, platformRepo string, req api.RemoteRepoBranchRulesDTO) (*api.RemoteRepoBranchRulesDTO, error) {
	override := &po.BranchRuleOverride{
		ProviderConfigID: providerConfigID,
		PlatformOwner:    platformOwner,
		PlatformRepo:     platformRepo,
		UseCustomRules:   req.UseCustomRules,
	}
	if req.UseCustomRules {
		override.SetRules(dtoToRules(req.Rules))
		override.SetProtected(req.Protected)
	} else {
		override.RulesJSON = ""
		override.ProtectedJSON = ""
	}
	dao := db.NewBranchRuleOverrideDAO()
	if err := dao.Upsert(override); err != nil {
		return nil, fmt.Errorf("failed to save override: %w", err)
	}
	return GetRemoteRepoRules(providerConfigID, platformOwner, platformRepo)
}

func ValidateBranchName(repoKey, branchName, baseRef string, skipRules bool) (*api.BranchValidationResult, error) {
	if skipRules {
		return &api.BranchValidationResult{Valid: true}, nil
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		return nil, fmt.Errorf("repo not found: %w", err)
	}

	rules, protected := resolveEffectiveRules(repo)

	for _, p := range protected {
		if branchName == p {
			return &api.BranchValidationResult{
				Valid: false,
				Errors: []api.BranchValidationError{{
					Field:   "name",
					Message: fmt.Sprintf("「%s」是保护分支，禁止直接创建", p),
				}},
			}, nil
		}
	}

	for _, rule := range rules {
		if !strings.HasPrefix(branchName, rule.Prefix) {
			continue
		}

		var errors []api.BranchValidationError

		if rule.RequireTaskID {
			if rule.TaskIDPattern != "" {
				matched, _ := regexp.MatchString(rule.TaskIDPattern, branchName)
				if !matched {
					errors = append(errors, api.BranchValidationError{
						Field:   "name",
						Message: fmt.Sprintf("分支名需包含任务标识，匹配规则：%s。示例：%sTASK-123-desc", rule.TaskIDPattern, rule.Prefix),
					})
				}
			}
		}

		if baseRef != "" && len(rule.GetSourceBranches()) > 0 {
			baseMatch := false
			for _, src := range rule.GetSourceBranches() {
				if baseRef == src {
					baseMatch = true
					break
				}
			}
			if !baseMatch {
				errors = append(errors, api.BranchValidationError{
					Field:   "base_ref",
					Message: fmt.Sprintf("%s 分支应来源于 %s，当前来源为 %s", rule.Prefix, strings.Join(rule.GetSourceBranches(), "/"), baseRef),
				})
			}
		}

		if len(errors) > 0 {
			return &api.BranchValidationResult{
				Valid:    false,
				RuleName: rule.DisplayName,
				Errors:   errors,
			}, nil
		}

		return &api.BranchValidationResult{
			Valid:    true,
			RuleName: rule.DisplayName,
		}, nil
	}

	return &api.BranchValidationResult{Valid: true}, nil
}

func resolveEffectiveRules(repo *po.Repo) ([]po.BranchRule, []string) {
	if repo.ProviderConfigID != 0 && repo.PlatformOwner != "" && repo.PlatformRepo != "" {
		overrideDAO := db.NewBranchRuleOverrideDAO()
		override, err := overrideDAO.FindByRemoteRepo(repo.ProviderConfigID, repo.PlatformOwner, repo.PlatformRepo)
		if err == nil && override.UseCustomRules {
			return override.GetRules(), override.GetProtected()
		}
	}

	setDAO := db.NewBranchRuleSetDAO()
	set, err := setDAO.FindGlobal()
	if err != nil || !set.Enabled {
		return nil, nil
	}
	return set.GetRules(), set.GetProtected()
}

func findLinkedRepos(providerConfigID uint, platformOwner, platformRepo string) []api.LinkedRepoDTO {
	bindingDAO := db.NewRepoProviderBindingDAO()
	bindings, err := bindingDAO.FindByProviderConfigID(providerConfigID)
	if err != nil {
		return nil
	}
	var repos []api.LinkedRepoDTO
	for _, b := range bindings {
		if b.PlatformOwner == platformOwner && b.PlatformRepo == platformRepo && b.RepoID > 0 {
			repoDAO := db.NewRepoDAO()
			if r, err := repoDAO.FindByID(b.RepoID); err == nil {
				repos = append(repos, api.LinkedRepoDTO{ID: r.ID, Key: r.Key, Name: r.Name})
			}
		}
	}
	return repos
}

func ruleSetToDTO(set *po.BranchRuleSet) *api.BranchRuleSetDTO {
	return &api.BranchRuleSetDTO{
		Enabled:   set.Enabled,
		Rules:     rulesToDTOs(set.GetRules()),
		Protected: set.GetProtected(),
	}
}

func rulesToDTOs(rules []po.BranchRule) []api.BranchRuleDTO {
	dtos := make([]api.BranchRuleDTO, 0, len(rules))
	for _, r := range rules {
		dtos = append(dtos, BranchRuleToDTO(r))
	}
	return dtos
}

func BranchRuleToDTO(r po.BranchRule) api.BranchRuleDTO {
	return api.BranchRuleDTO{
		ID:                r.ID,
		Prefix:            r.Prefix,
		DisplayName:       r.DisplayName,
		SourceBranches:    r.GetSourceBranches(),
		TargetBranches:    r.GetTargetBranches(),
		RequireTaskID:     r.RequireTaskID,
		TaskIDPattern:     r.TaskIDPattern,
		AutoDeleteOnMerge: r.AutoDeleteOnMerge,
		AllowDirectPush:   r.AllowDirectPush,
		RequireCodeReview: r.RequireCodeReview,
		SortOrder:         r.SortOrder,
	}
}

func dtoToRules(dtos []api.BranchRuleDTO) []po.BranchRule {
	rules := make([]po.BranchRule, 0, len(dtos))
	for i, d := range dtos {
		r := po.BranchRule{
			ID:                d.ID,
			Prefix:            d.Prefix,
			DisplayName:       d.DisplayName,
			RequireTaskID:     d.RequireTaskID,
			TaskIDPattern:     d.TaskIDPattern,
			AutoDeleteOnMerge: d.AutoDeleteOnMerge,
			AllowDirectPush:   d.AllowDirectPush,
			RequireCodeReview: d.RequireCodeReview,
			SortOrder:         i,
		}
		r.SetSourceBranches(d.SourceBranches)
		r.SetTargetBranches(d.TargetBranches)
		rules = append(rules, r)
	}
	return rules
}

func getDefaultRuleSetDTO() *api.BranchRuleSetDTO {
	rules := []api.BranchRuleDTO{
		{Prefix: "feature/", DisplayName: "功能开发分支", SourceBranches: []string{"develop"}, TargetBranches: []string{"develop"}, RequireTaskID: true, TaskIDPattern: "[A-Z]+-\\d+", AutoDeleteOnMerge: true, AllowDirectPush: false, RequireCodeReview: false},
		{Prefix: "bugfix/", DisplayName: "缺陷修复分支", SourceBranches: []string{"develop"}, TargetBranches: []string{"develop"}, RequireTaskID: true, TaskIDPattern: "[A-Z]+-\\d+", AutoDeleteOnMerge: true, AllowDirectPush: false, RequireCodeReview: false},
		{Prefix: "hotfix/", DisplayName: "紧急修复分支", SourceBranches: []string{"main", "master"}, TargetBranches: []string{"main", "master"}, RequireTaskID: true, TaskIDPattern: "[A-Z]+-\\d+", AutoDeleteOnMerge: true, AllowDirectPush: false, RequireCodeReview: true},
		{Prefix: "release/", DisplayName: "发布分支", SourceBranches: []string{"develop"}, TargetBranches: []string{"main", "master"}, RequireTaskID: false, AutoDeleteOnMerge: false, AllowDirectPush: false, RequireCodeReview: true},
	}
	b, _ := json.Marshal(rules)
	_ = b
	return &api.BranchRuleSetDTO{
		Enabled:   true,
		Rules:     rules,
		Protected: []string{"main", "master", "develop"},
	}
}
