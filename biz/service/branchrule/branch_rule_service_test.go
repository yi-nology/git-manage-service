package branchrule

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
	settingsModel "github.com/yi-nology/git-manage-service/biz/model/settings"
)

func TestBranchRuleToProto(t *testing.T) {
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

	proto := branchRuleToProto(rule)
	if proto.GetPrefix() != "feature/" {
		t.Errorf("expected feature/, got %s", proto.GetPrefix())
	}
	if proto.GetDisplayName() != "功能分支" {
		t.Errorf("expected 功能分支, got %s", proto.GetDisplayName())
	}
	if !proto.GetRequireTaskId() {
		t.Error("expected RequireTaskId true")
	}
	if len(proto.SourceBranches) != 1 || proto.SourceBranches[0] != "develop" {
		t.Errorf("expected [develop], got %v", proto.SourceBranches)
	}
}

func TestProtoToRules(t *testing.T) {
	protos := []*settingsModel.BranchRule{
		{
			Prefix:            "feature/",
			DisplayName:       "功能分支",
			SourceBranches:    []string{"develop"},
			TargetBranches:    []string{"develop"},
			RequireTaskId:     true,
			TaskIdPattern:     "[A-Z]+-\\d+",
			AutoDeleteOnMerge: true,
		},
		{
			Prefix:            "hotfix/",
			DisplayName:       "紧急修复",
			SourceBranches:    []string{"main"},
			TargetBranches:    []string{"main"},
			RequireTaskId:     true,
			TaskIdPattern:     "[A-Z]+-\\d+",
			RequireCodeReview: true,
		},
	}
	rules := protoToRules(protos)
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

func TestProtoToRules_Empty(t *testing.T) {
	rules := protoToRules(nil)
	if len(rules) != 0 {
		t.Errorf("expected 0 rules for nil input, got %d", len(rules))
	}
}

func TestGetDefaultRuleSet(t *testing.T) {
	dto := getDefaultRuleSet()
	if !dto.Enabled {
		t.Error("expected default rules to be enabled")
	}
	if len(dto.Rules) == 0 {
		t.Error("expected non-empty default rules")
	}
	if len(dto.ProtectedBranches) == 0 {
		t.Error("expected non-empty protected branches")
	}
}

func TestGetDefaultRuleSet_ContainsStandardPrefixes(t *testing.T) {
	dto := getDefaultRuleSet()
	prefixes := map[string]bool{}
	for _, r := range dto.Rules {
		prefixes[r.GetPrefix()] = true
	}
	expected := []string{"feature/", "bugfix/", "hotfix/", "release/"}
	for _, p := range expected {
		if !prefixes[p] {
			t.Errorf("missing default prefix: %s", p)
		}
	}
}

func TestGetDefaultRuleSet_ProtectedBranches(t *testing.T) {
	dto := getDefaultRuleSet()
	protected := map[string]bool{}
	for _, p := range dto.ProtectedBranches {
		protected[p] = true
	}
	expected := []string{"main", "master", "develop"}
	for _, p := range expected {
		if !protected[p] {
			t.Errorf("missing protected branch: %s", p)
		}
	}
}

func TestRuleSetToProto(t *testing.T) {
	set := &po.BranchRuleSet{
		Enabled: true,
	}
	rule := po.BranchRule{Prefix: "feature/", DisplayName: "功能分支"}
	rule.SetSourceBranches([]string{"develop"})
	set.SetRules([]po.BranchRule{rule})
	set.SetProtected([]string{"main"})

	proto := ruleSetToProto(set)
	if !proto.Enabled {
		t.Error("expected enabled")
	}
	if len(proto.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(proto.Rules))
	}
	if proto.Rules[0].GetPrefix() != "feature/" {
		t.Errorf("expected feature/, got %s", proto.Rules[0].GetPrefix())
	}
	if len(proto.ProtectedBranches) != 1 || proto.ProtectedBranches[0] != "main" {
		t.Errorf("expected [main], got %v", proto.ProtectedBranches)
	}
}

func TestRulesToProtos_Empty(t *testing.T) {
	protos := rulesToProtos(nil)
	if len(protos) != 0 {
		t.Errorf("expected 0 protos, got %d", len(protos))
	}
}

func TestRoundTrip_ProtoToRulesToProto(t *testing.T) {
	original := []*settingsModel.BranchRule{
		{
			Prefix:            "feature/",
			DisplayName:       "功能分支",
			SourceBranches:    []string{"develop"},
			TargetBranches:    []string{"develop"},
			RequireTaskId:     true,
			TaskIdPattern:     "[A-Z]+-\\d+",
			AutoDeleteOnMerge: true,
			AllowDirectPush:   false,
			RequireCodeReview: false,
		},
	}
	rules := protoToRules(original)
	protos := rulesToProtos(rules)
	if protos[0].Prefix != original[0].Prefix {
		t.Errorf("round-trip prefix mismatch: %s vs %s", protos[0].Prefix, original[0].Prefix)
	}
	if protos[0].DisplayName != original[0].DisplayName {
		t.Errorf("round-trip displayName mismatch: %s vs %s", protos[0].DisplayName, original[0].DisplayName)
	}
}

// Helper functions
func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
