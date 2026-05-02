package branchrule

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestBranchRuleToDTO(t *testing.T) {
	rule := po.BranchRule{
		ID:                1,
		Prefix:            "feature/",
		DisplayName:       "功能分支",
		RequireTaskID:     true,
		TaskIDPattern:     "[A-Z]+-\\d+",
		AutoDeleteOnMerge: true,
		AllowDirectPush:   false,
		RequireCodeReview: false,
		SortOrder:         0,
	}
	rule.SetSourceBranches([]string{"develop"})
	rule.SetTargetBranches([]string{"develop"})

	dto := BranchRuleToDTO(rule)
	if dto.Prefix != "feature/" {
		t.Errorf("expected feature/, got %s", dto.Prefix)
	}
	if dto.DisplayName != "功能分支" {
		t.Errorf("expected 功能分支, got %s", dto.DisplayName)
	}
	if !dto.RequireTaskID {
		t.Error("expected RequireTaskID true")
	}
	if len(dto.SourceBranches) != 1 || dto.SourceBranches[0] != "develop" {
		t.Errorf("expected [develop], got %v", dto.SourceBranches)
	}
}

func TestDtoToRules(t *testing.T) {
	dtos := []api.BranchRuleDTO{
		{
			Prefix:            "feature/",
			DisplayName:       "功能分支",
			SourceBranches:    []string{"develop"},
			TargetBranches:    []string{"develop"},
			RequireTaskID:     true,
			TaskIDPattern:     "[A-Z]+-\\d+",
			AutoDeleteOnMerge: true,
		},
		{
			Prefix:            "hotfix/",
			DisplayName:       "紧急修复",
			SourceBranches:    []string{"main"},
			TargetBranches:    []string{"main"},
			RequireTaskID:     true,
			TaskIDPattern:     "[A-Z]+-\\d+",
			RequireCodeReview: true,
		},
	}
	rules := dtoToRules(dtos)
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(rules))
	}
	if rules[0].Prefix != "feature/" {
		t.Errorf("expected feature/, got %s", rules[0].Prefix)
	}
	if rules[0].SortOrder != 0 {
		t.Errorf("expected SortOrder 0, got %d", rules[0].SortOrder)
	}
	if rules[1].SortOrder != 1 {
		t.Errorf("expected SortOrder 1, got %d", rules[1].SortOrder)
	}
	srcs := rules[0].GetSourceBranches()
	if len(srcs) != 1 || srcs[0] != "develop" {
		t.Errorf("expected [develop], got %v", srcs)
	}
}

func TestDtoToRules_Empty(t *testing.T) {
	rules := dtoToRules(nil)
	if len(rules) != 0 {
		t.Errorf("expected 0 rules for nil input, got %d", len(rules))
	}
}

func TestGetDefaultRuleSetDTO(t *testing.T) {
	dto := getDefaultRuleSetDTO()
	if !dto.Enabled {
		t.Error("expected default rules to be enabled")
	}
	if len(dto.Rules) == 0 {
		t.Error("expected non-empty default rules")
	}
	if len(dto.Protected) == 0 {
		t.Error("expected non-empty protected branches")
	}
}

func TestGetDefaultRuleSetDTO_ContainsStandardPrefixes(t *testing.T) {
	dto := getDefaultRuleSetDTO()
	prefixes := map[string]bool{}
	for _, r := range dto.Rules {
		prefixes[r.Prefix] = true
	}
	expected := []string{"feature/", "bugfix/", "hotfix/", "release/"}
	for _, p := range expected {
		if !prefixes[p] {
			t.Errorf("missing default prefix: %s", p)
		}
	}
}

func TestGetDefaultRuleSetDTO_ProtectedBranches(t *testing.T) {
	dto := getDefaultRuleSetDTO()
	protected := map[string]bool{}
	for _, p := range dto.Protected {
		protected[p] = true
	}
	expected := []string{"main", "master", "develop"}
	for _, p := range expected {
		if !protected[p] {
			t.Errorf("missing protected branch: %s", p)
		}
	}
}

func TestRuleSetToDTO(t *testing.T) {
	set := &po.BranchRuleSet{
		Enabled: true,
	}
	rule := po.BranchRule{Prefix: "feature/", DisplayName: "功能分支"}
	rule.SetSourceBranches([]string{"develop"})
	set.SetRules([]po.BranchRule{rule})
	set.SetProtected([]string{"main"})

	dto := ruleSetToDTO(set)
	if !dto.Enabled {
		t.Error("expected enabled")
	}
	if len(dto.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(dto.Rules))
	}
	if dto.Rules[0].Prefix != "feature/" {
		t.Errorf("expected feature/, got %s", dto.Rules[0].Prefix)
	}
	if len(dto.Protected) != 1 || dto.Protected[0] != "main" {
		t.Errorf("expected [main], got %v", dto.Protected)
	}
}

func TestRulesToDTOs_Empty(t *testing.T) {
	dtos := rulesToDTOs(nil)
	if len(dtos) != 0 {
		t.Errorf("expected 0 DTOs, got %d", len(dtos))
	}
}

func TestRoundTrip_DTOToRulesToDTO(t *testing.T) {
	original := []api.BranchRuleDTO{
		{
			Prefix:            "feature/",
			DisplayName:       "功能分支",
			SourceBranches:    []string{"develop"},
			TargetBranches:    []string{"develop"},
			RequireTaskID:     true,
			TaskIDPattern:     "[A-Z]+-\\d+",
			AutoDeleteOnMerge: true,
			AllowDirectPush:   false,
			RequireCodeReview: false,
		},
	}
	rules := dtoToRules(original)
	dtos := rulesToDTOs(rules)
	if dtos[0].Prefix != original[0].Prefix {
		t.Errorf("round-trip prefix mismatch: %s vs %s", dtos[0].Prefix, original[0].Prefix)
	}
	if dtos[0].DisplayName != original[0].DisplayName {
		t.Errorf("round-trip displayName mismatch: %s vs %s", dtos[0].DisplayName, original[0].DisplayName)
	}
}
