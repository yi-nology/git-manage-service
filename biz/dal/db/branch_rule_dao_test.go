package db

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestBranchRuleSetDAO_UpsertAndFind(t *testing.T) {
	SetupTestDB(t)
	dao := NewBranchRuleSetDAO()

	set := &po.BranchRuleSet{
		ScopeType: "global",
		ScopeID:   "default",
		Enabled:   true,
	}
	rule := po.BranchRule{Prefix: "feature/", DisplayName: "功能分支"}
	set.SetRules([]po.BranchRule{rule})
	set.SetProtected([]string{"main", "develop"})

	if err := dao.Upsert(set); err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	found, err := dao.FindGlobal()
	if err != nil {
		t.Fatalf("FindGlobal failed: %v", err)
	}
	if !found.Enabled {
		t.Error("expected enabled")
	}
	rules := found.GetRules()
	if len(rules) != 1 || rules[0].Prefix != "feature/" {
		t.Errorf("rules mismatch: %v", rules)
	}
	protected := found.GetProtected()
	if len(protected) != 2 {
		t.Errorf("expected 2 protected, got %d", len(protected))
	}
}

func TestBranchRuleSetDAO_UpdateExisting(t *testing.T) {
	SetupTestDB(t)
	dao := NewBranchRuleSetDAO()

	set1 := &po.BranchRuleSet{ScopeType: "global", ScopeID: "default", Enabled: true}
	set1.SetRules([]po.BranchRule{{Prefix: "feature/"}})
	dao.Upsert(set1)

	set2 := &po.BranchRuleSet{ScopeType: "global", ScopeID: "default", Enabled: false}
	set2.SetRules([]po.BranchRule{{Prefix: "hotfix/"}})
	dao.Upsert(set2)

	found, _ := dao.FindGlobal()
	if found.Enabled {
		t.Error("expected disabled after second upsert")
	}
	rules := found.GetRules()
	if len(rules) != 1 || rules[0].Prefix != "hotfix/" {
		t.Errorf("expected hotfix/ after update, got %v", rules)
	}
}

func TestBranchRuleOverrideDAO_UpsertAndFind(t *testing.T) {
	SetupTestDB(t)
	dao := NewBranchRuleOverrideDAO()

	override := &po.BranchRuleOverride{
		ProviderConfigID: 1,
		PlatformOwner:    "myorg",
		PlatformRepo:     "myrepo",
		UseCustomRules:   true,
	}
	override.SetRules([]po.BranchRule{{Prefix: "release/", DisplayName: "Release"}})
	override.SetProtected([]string{"main"})

	if err := dao.Upsert(override); err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}

	found, err := dao.FindByRemoteRepo(1, "myorg", "myrepo")
	if err != nil {
		t.Fatalf("FindByRemoteRepo failed: %v", err)
	}
	if !found.UseCustomRules {
		t.Error("expected UseCustomRules=true")
	}
	rules := found.GetRules()
	if len(rules) != 1 || rules[0].Prefix != "release/" {
		t.Errorf("rules mismatch: %v", rules)
	}
}

func TestBranchRuleOverrideDAO_Delete(t *testing.T) {
	SetupTestDB(t)
	dao := NewBranchRuleOverrideDAO()
	override := &po.BranchRuleOverride{
		ProviderConfigID: 1,
		PlatformOwner:    "org",
		PlatformRepo:     "repo",
	}
	dao.Upsert(override)

	found, _ := dao.FindByRemoteRepo(1, "org", "repo")
	dao.Delete(found.ID)

	_, err := dao.FindByRemoteRepo(1, "org", "repo")
	if err == nil {
		t.Error("expected error after delete")
	}
}
