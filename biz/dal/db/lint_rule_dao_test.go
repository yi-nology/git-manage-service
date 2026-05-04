package db

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestLintRuleDAO_CRUD(t *testing.T) {
	SetupTestDB(t)
	dao := NewLintRuleDAO()

	rule := &po.LintRule{
		ID:          "test-rule-1",
		Name:        "Test Rule",
		Description: "A test rule",
		Category:    "test",
		Severity:    "error",
		Pattern:     "required_name",
		Enabled:     true,
		Priority:    1,
	}
	if err := dao.Create(rule); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	found, err := dao.FindByID("test-rule-1")
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}
	if found.Name != "Test Rule" {
		t.Errorf("name mismatch: got %s", found.Name)
	}

	found.Enabled = false
	dao.Save(found)
	updated, _ := dao.FindByID("test-rule-1")
	if updated.Enabled {
		t.Error("expected Enabled=false after update")
	}

	if err := dao.Delete("test-rule-1"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	_, err = dao.FindByID("test-rule-1")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestLintRuleDAO_BatchCreate(t *testing.T) {
	SetupTestDB(t)
	dao := NewLintRuleDAO()

	rules := []po.LintRule{
		{ID: "batch-1", Name: "B1", Category: "test", Severity: "error", Pattern: "p1", Enabled: true, Priority: 1},
		{ID: "batch-2", Name: "B2", Category: "test", Severity: "warning", Pattern: "p2", Enabled: true, Priority: 2},
	}
	if err := dao.BatchCreate(rules); err != nil {
		t.Fatalf("BatchCreate failed: %v", err)
	}

	count, _ := dao.Count()
	if count != 2 {
		t.Errorf("expected 2 rules, got %d", count)
	}
}

func TestLintRuleDAO_FindEnabled(t *testing.T) {
	SetupTestDB(t)
	dao := NewLintRuleDAO()
	dao.Create(&po.LintRule{ID: "e1", Name: "E1", Category: "test", Severity: "error", Pattern: "p", Enabled: true, Priority: 1})

	enabled, err := dao.FindEnabled()
	if err != nil {
		t.Fatal(err)
	}
	if len(enabled) < 1 {
		t.Errorf("expected at least 1 enabled, got %d", len(enabled))
	}
}

func TestLintRuleDAO_FindByCategory(t *testing.T) {
	SetupTestDB(t)
	dao := NewLintRuleDAO()
	dao.Create(&po.LintRule{ID: "c1", Name: "C1", Category: "style", Severity: "info", Pattern: "p", Enabled: true, Priority: 1})
	dao.Create(&po.LintRule{ID: "c2", Name: "C2", Category: "required", Severity: "error", Pattern: "p", Enabled: true, Priority: 2})

	styleRules, err := dao.FindByCategory("style")
	if err != nil {
		t.Fatal(err)
	}
	if len(styleRules) != 1 {
		t.Errorf("expected 1 style rule, got %d", len(styleRules))
	}
}

func TestLintRuleDAO_FindByIDs(t *testing.T) {
	SetupTestDB(t)
	dao := NewLintRuleDAO()
	dao.Create(&po.LintRule{ID: "id1", Name: "R1", Category: "test", Severity: "error", Pattern: "p", Enabled: true, Priority: 1})
	dao.Create(&po.LintRule{ID: "id2", Name: "R2", Category: "test", Severity: "error", Pattern: "p", Enabled: true, Priority: 2})
	dao.Create(&po.LintRule{ID: "id3", Name: "R3", Category: "test", Severity: "error", Pattern: "p", Enabled: true, Priority: 3})

	rules, err := dao.FindByIDs([]string{"id1", "id3"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(rules))
	}
}

func TestLintRuleDAO_ExistsByID(t *testing.T) {
	SetupTestDB(t)
	dao := NewLintRuleDAO()
	dao.Create(&po.LintRule{ID: "exists-test", Name: "E", Category: "test", Severity: "error", Pattern: "p", Enabled: true, Priority: 1})

	exists, _ := dao.ExistsByID("exists-test")
	if !exists {
		t.Error("expected ID to exist")
	}
	exists, _ = dao.ExistsByID("nonexistent")
	if exists {
		t.Error("expected ID not to exist")
	}
}

func TestLintRuleDAO_FindAll_OrderedByPriority(t *testing.T) {
	SetupTestDB(t)
	dao := NewLintRuleDAO()
	dao.Create(&po.LintRule{ID: "p3", Name: "P3", Category: "test", Severity: "error", Pattern: "p", Enabled: true, Priority: 3})
	dao.Create(&po.LintRule{ID: "p1", Name: "P1", Category: "test", Severity: "error", Pattern: "p", Enabled: true, Priority: 1})
	dao.Create(&po.LintRule{ID: "p2", Name: "P2", Category: "test", Severity: "error", Pattern: "p", Enabled: true, Priority: 2})

	all, err := dao.FindAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(all))
	}
	if all[0].Priority > all[1].Priority {
		t.Errorf("expected ordered by priority, got %d before %d", all[0].Priority, all[1].Priority)
	}
}
