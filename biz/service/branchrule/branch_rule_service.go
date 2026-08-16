package branchrule

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	settingsModel "github.com/yi-nology/git-manage-service/biz/model/settings"
	servicePkg "github.com/yi-nology/git-manage-service/pkg/service"
)

func GetGlobalRules() (*settingsModel.BranchRuleSet, error) {
	dao := db.NewBranchRuleSetDAO()
	set, err := dao.FindGlobal()
	if err != nil {
		return getDefaultRuleSet(), nil
	}
	return ruleSetToProto(set), nil
}

func UpdateGlobalRules(req *settingsModel.BranchRuleSet) (*settingsModel.BranchRuleSet, error) {
	dao := db.NewBranchRuleSetDAO()
	set := &po.BranchRuleSet{
		ScopeType: "global",
		ScopeID:   "default",
		Enabled:   req.Enabled,
	}
	rules := protoToRules(req.Rules)
	protected := req.ProtectedBranches
	if protected == nil {
		protected = []string{}
	}
	set.SetRules(rules)
	set.SetProtected(protected)
	if err := dao.Upsert(set); err != nil {
		return nil, fmt.Errorf("failed to save global rules: %w", err)
	}
	return ruleSetToProto(set), nil
}

func GetRemoteRepoRules(providerConfigID uint, platformOwner, platformRepo string) (*settingsModel.RemoteRepoBranchRuleSet, error) {
	overrideDAO := db.NewBranchRuleOverrideDAO()
	override, err := overrideDAO.FindByRemoteRepo(providerConfigID, platformOwner, platformRepo)

	result := &settingsModel.RemoteRepoBranchRuleSet{
		ProviderConfigId: uint64(providerConfigID),
		PlatformOwner:    platformOwner,
		PlatformRepo:     platformRepo,
		UseCustomRules:   false,
	}

	if err == nil {
		result.UseCustomRules = override.UseCustomRules
		if override.UseCustomRules {
			result.Rules = rulesToProtos(override.GetRules())
			result.ProtectedBranches = override.GetProtected()
		} else {
			globalDTO, _ := GetGlobalRules()
			result.Rules = globalDTO.Rules
			result.ProtectedBranches = globalDTO.ProtectedBranches
		}
	} else {
		globalDTO, _ := GetGlobalRules()
		result.Rules = globalDTO.Rules
		result.ProtectedBranches = globalDTO.ProtectedBranches
	}

	result.LinkedRepos = findLinkedRepos(providerConfigID, platformOwner, platformRepo)
	return result, nil
}

func UpdateRemoteRepoRules(providerConfigID uint, platformOwner, platformRepo string, req *settingsModel.RemoteRepoBranchRuleSet) (*settingsModel.RemoteRepoBranchRuleSet, error) {
	override := &po.BranchRuleOverride{
		ProviderConfigID: providerConfigID,
		PlatformOwner:    platformOwner,
		PlatformRepo:     platformRepo,
		UseCustomRules:   req.UseCustomRules,
	}
	if req.UseCustomRules {
		override.SetRules(protoToRules(req.Rules))
		override.SetProtected(req.ProtectedBranches)
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

func ValidateBranchName(repoKey, branchName, baseRef string, skipRules bool) (*settingsModel.ValidateBranchNameResponse, error) {
	if skipRules {
		return &settingsModel.ValidateBranchNameResponse{Valid: true}, nil
	}

	repo, err := servicePkg.GetRepoByKey(repoKey)
	if err != nil {
		return nil, err
	}

	rules, protected := resolveEffectiveRules(repo)

	for _, p := range protected {
		if branchName == p {
			return &settingsModel.ValidateBranchNameResponse{
				Valid:   false,
				Message: fmt.Sprintf("「%s」是保护分支，禁止直接创建", p),
			}, nil
		}
	}

	for _, rule := range rules {
		if !strings.HasPrefix(branchName, rule.Prefix) {
			continue
		}

		if rule.RequireTaskID {
			if rule.TaskIDPattern != "" {
				matched, _ := regexp.MatchString(rule.TaskIDPattern, branchName)
				if !matched {
					return &settingsModel.ValidateBranchNameResponse{
						Valid:   false,
						Message: fmt.Sprintf("分支名需包含任务标识，匹配规则：%s。示例：%sTASK-123-desc", rule.TaskIDPattern, rule.Prefix),
					}, nil
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
				return &settingsModel.ValidateBranchNameResponse{
					Valid:   false,
					Message: fmt.Sprintf("%s 分支应来源于 %s，当前来源为 %s", rule.Prefix, strings.Join(rule.GetSourceBranches(), "/"), baseRef),
				}, nil
			}
		}

		return &settingsModel.ValidateBranchNameResponse{
			Valid:   true,
			Message: rule.DisplayName,
		}, nil
	}

	return &settingsModel.ValidateBranchNameResponse{Valid: true}, nil
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

func findLinkedRepos(providerConfigID uint, platformOwner, platformRepo string) []*settingsModel.LinkedRepo {
	bindingDAO := db.NewRepoProviderBindingDAO()
	bindings, err := bindingDAO.FindByProviderConfigID(providerConfigID)
	if err != nil {
		return nil
	}
	var repos []*settingsModel.LinkedRepo
	for _, b := range bindings {
		if b.PlatformOwner == platformOwner && b.PlatformRepo == platformRepo && b.RepoID > 0 {
			repoDAO := db.NewRepoDAO()
			if r, err := repoDAO.FindByID(b.RepoID); err == nil {
				repos = append(repos, &settingsModel.LinkedRepo{
					RepoKey:  r.Key,
					RepoName: r.Name,
				})
			}
		}
	}
	return repos
}

func ruleSetToProto(set *po.BranchRuleSet) *settingsModel.BranchRuleSet {
	return &settingsModel.BranchRuleSet{
		Enabled:           set.Enabled,
		Rules:             rulesToProtos(set.GetRules()),
		ProtectedBranches: set.GetProtected(),
	}
}

func rulesToProtos(rules []po.BranchRule) []*settingsModel.BranchRule {
	protos := make([]*settingsModel.BranchRule, 0, len(rules))
	for _, r := range rules {
		protos = append(protos, branchRuleToProto(r))
	}
	return protos
}

func branchRuleToProto(r po.BranchRule) *settingsModel.BranchRule {
	return &settingsModel.BranchRule{
		Id:                uint64(r.ID),
		Prefix:            r.Prefix,
		DisplayName:       r.DisplayName,
		SourceBranches:    r.GetSourceBranches(),
		TargetBranches:    r.GetTargetBranches(),
		RequireTaskId:     r.RequireTaskID,
		TaskIdPattern:     r.TaskIDPattern,
		AutoDeleteOnMerge: r.AutoDeleteOnMerge,
		AllowDirectPush:   r.AllowDirectPush,
		RequireCodeReview: r.RequireCodeReview,
		SortOrder:         int32(r.SortOrder),
	}
}

func protoToRules(protos []*settingsModel.BranchRule) []po.BranchRule {
	rules := make([]po.BranchRule, 0, len(protos))
	for i, p := range protos {
		r := po.BranchRule{
			ID:                uint(p.Id),
			Prefix:            p.Prefix,
			DisplayName:       p.DisplayName,
			RequireTaskID:     p.RequireTaskId,
			TaskIDPattern:     p.TaskIdPattern,
			AutoDeleteOnMerge: p.AutoDeleteOnMerge,
			AllowDirectPush:   p.AllowDirectPush,
			RequireCodeReview: p.RequireCodeReview,
			SortOrder:         i,
		}
		r.SetSourceBranches(p.SourceBranches)
		r.SetTargetBranches(p.TargetBranches)
		rules = append(rules, r)
	}
	return rules
}

func getDefaultRuleSet() *settingsModel.BranchRuleSet {
	rules := []*settingsModel.BranchRule{
		{Prefix: "feature/", DisplayName: "功能开发分支", SourceBranches: []string{"develop"}, TargetBranches: []string{"develop"}, RequireTaskId: true, TaskIdPattern: "[A-Z]+-\\d+", AutoDeleteOnMerge: true, AllowDirectPush: false, RequireCodeReview: false},
		{Prefix: "bugfix/", DisplayName: "缺陷修复分支", SourceBranches: []string{"develop"}, TargetBranches: []string{"develop"}, RequireTaskId: true, TaskIdPattern: "[A-Z]+-\\d+", AutoDeleteOnMerge: true, AllowDirectPush: false, RequireCodeReview: false},
		{Prefix: "hotfix/", DisplayName: "紧急修复分支", SourceBranches: []string{"main", "master"}, TargetBranches: []string{"main", "master"}, RequireTaskId: true, TaskIdPattern: "[A-Z]+-\\d+", AutoDeleteOnMerge: true, AllowDirectPush: false, RequireCodeReview: true},
		{Prefix: "release/", DisplayName: "发布分支", SourceBranches: []string{"develop"}, TargetBranches: []string{"main", "master"}, RequireTaskId: false, AutoDeleteOnMerge: false, AllowDirectPush: false, RequireCodeReview: true},
	}
	return &settingsModel.BranchRuleSet{
		Enabled:           true,
		Rules:             rules,
		ProtectedBranches: []string{"main", "master", "develop"},
	}
}
