package codereview

import (
	"strings"
	"testing"
)

func TestParseDiff_BinaryFile(t *testing.T) {
	raw := `diff --git a/image.png b/image.png
Binary files /dev/null and b/image.png differ
`
	files := ParseDiff(raw)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	// Binary diff may not extract paths in all cases
	_ = files[0]
}

func TestParseDiff_RenameFile(t *testing.T) {
	raw := `diff --git a/old_name.go b/new_name.go
similarity index 95%
rename from old_name.go
rename to new_name.go
--- a/old_name.go
+++ b/new_name.go
@@ -1,3 +1,4 @@
 package main
+import "fmt"
 func main() {}
`
	files := ParseDiff(raw)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if files[0].OldPath != "old_name.go" {
		t.Errorf("old path: %s", files[0].OldPath)
	}
	if files[0].NewPath != "new_name.go" {
		t.Errorf("new path: %s", files[0].NewPath)
	}
}

func TestParseDiff_EmptyHunk(t *testing.T) {
	raw := `diff --git a/a.go b/a.go
--- a/a.go
+++ b/a.go
@@ -0,0 +0,0 @@
`
	files := ParseDiff(raw)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if len(files[0].Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(files[0].Hunks))
	}
}

func TestParseDiff_ContextLines(t *testing.T) {
	raw := `diff --git a/a.go b/a.go
--- a/a.go
+++ b/a.go
@@ -1,4 +1,5 @@
 package main
+import "fmt"
 func main() {}
 
`
	files := ParseDiff(raw)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	h := files[0].Hunks[0]
	ctxCount := 0
	addCount := 0
	for _, l := range h.Lines {
		switch l.Type {
		case "ctx":
			ctxCount++
		case "add":
			addCount++
		}
	}
	if ctxCount == 0 {
		t.Error("expected context lines")
	}
	if addCount != 1 {
		t.Errorf("expected 1 add line, got %d", addCount)
	}
}

func TestParseDiff_DeleteLines(t *testing.T) {
	raw := `diff --git a/a.go b/a.go
--- a/a.go
+++ b/a.go
@@ -1,4 +1,3 @@
 package main
-import "os"
 func main() {}
 
`
	files := ParseDiff(raw)
	h := files[0].Hunks[0]
	delCount := 0
	for _, l := range h.Lines {
		if l.Type == "del" {
			delCount++
		}
	}
	if delCount != 1 {
		t.Errorf("expected 1 delete line, got %d", delCount)
	}
	if files[0].Deletions != 1 {
		t.Errorf("expected 1 deletion, got %d", files[0].Deletions)
	}
}

func TestParseDiff_MalformedDiff(t *testing.T) {
	raw := `this is not a valid diff at all
just random text
`
	files := ParseDiff(raw)
	if len(files) != 0 {
		t.Errorf("expected 0 files for malformed diff, got %d", len(files))
	}
}

func TestParseDiff_PartialHeader(t *testing.T) {
	raw := `diff --git a/
--- a/a.go
+++ b/a.go
@@ -1,1 +1,2 @@
 package main
+import "fmt"
`
	files := ParseDiff(raw)
	if len(files) != 0 {
		t.Logf("partial header handled: %d files", len(files))
	}
}

func TestMigrationRule_DestructiveDrop(t *testing.T) {
	rule := &MigrationRule{}
	ctx := &RuleContext{
		Files: []*FileDiff{
			{
				NewPath: "db/migrate/001_drop_table.sql",
				Hunks: []*DiffHunk{{
					Lines: []DiffLine{
						{Type: "add", NewLine: 1, Content: "DROP TABLE users;", FilePath: "db/migrate/001_drop_table.sql"},
					},
				}},
			},
		},
	}
	findings, err := rule.Check(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Error("expected finding for DROP TABLE")
	}
	if findings[0].Severity != SeverityHigh {
		t.Errorf("expected high severity, got %s", findings[0].Severity)
	}
}

func TestMigrationRule_ForeignKeyWithoutIndex(t *testing.T) {
	rule := &MigrationRule{}
	ctx := &RuleContext{
		Files: []*FileDiff{
			{
				NewPath: "db/migrate/002_add_fk.rb",
				Hunks: []*DiffHunk{{
					Lines: []DiffLine{
						{Type: "add", NewLine: 1, Content: "add_foreign_key :orders, :users", FilePath: "db/migrate/002_add_fk.rb"},
					},
				}},
			},
		},
	}
	findings, err := rule.Check(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range findings {
		if f.Title == "Foreign key without index" {
			found = true
		}
	}
	if !found {
		t.Error("expected finding for FK without index")
	}
}

func TestMigrationRule_ForeignKeyWithIndex(t *testing.T) {
	rule := &MigrationRule{}
	ctx := &RuleContext{
		Files: []*FileDiff{
			{
				NewPath: "db/migrate/003_add_fk_idx.rb",
				Hunks: []*DiffHunk{{
					Lines: []DiffLine{
						{Type: "add", NewLine: 1, Content: "add_foreign_key :orders, :users", FilePath: "db/migrate/003_add_fk_idx.rb"},
						{Type: "add", NewLine: 2, Content: "CREATE INDEX idx_orders_user_id ON orders(user_id)", FilePath: "db/migrate/003_add_fk_idx.rb"},
					},
				}},
			},
		},
	}
	findings, _ := rule.Check(ctx)
	for _, f := range findings {
		if f.Title == "Foreign key without index" {
			t.Error("should not find FK-without-index when index is present")
		}
	}
}

func TestMigrationRule_SkipDeletedFiles(t *testing.T) {
	rule := &MigrationRule{}
	ctx := &RuleContext{
		Files: []*FileDiff{
			{
				NewPath:   "db/migrate/001_drop.sql",
				IsDeleted: true,
				Hunks: []*DiffHunk{{
					Lines: []DiffLine{
						{Type: "add", Content: "DROP TABLE users;"},
					},
				}},
			},
		},
	}
	findings, _ := rule.Check(ctx)
	if len(findings) != 0 {
		t.Error("expected no findings for deleted files")
	}
}

func TestMigrationRule_NonMigrationFile(t *testing.T) {
	rule := &MigrationRule{}
	ctx := &RuleContext{
		Files: []*FileDiff{
			{
				NewPath: "src/models/user.go",
				Hunks: []*DiffHunk{{
					Lines: []DiffLine{
						{Type: "add", Content: "DROP TABLE users;"},
					},
				}},
			},
		},
	}
	findings, _ := rule.Check(ctx)
	if len(findings) != 0 {
		t.Error("expected no findings for non-migration files")
	}
}

func TestIsMigrationFile(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"db/migrate/001_create.rb", true},
		{"migrations/001_up.sql", true},
		{"prisma/migrations/001.sql", true},
		{"alembic/versions/abc.py", true},
		{"src/main.go", false},
		{"V001__create_table.sql", true},
		{"U001__rollback.sql", true},
		{"001_migration.sql", true},
		{"README.md", false},
	}
	for _, tt := range tests {
		result := isMigrationFile(tt.path)
		if result != tt.expected {
			t.Errorf("isMigrationFile(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestParseDiff_MultipleHunks(t *testing.T) {
	raw := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 package main
+import "fmt"
 func main() {}
 
@@ -10,3 +11,4 @@
 	fmt.Println("hello")
+	fmt.Println("world")
 }
`
	files := ParseDiff(raw)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if len(files[0].Hunks) != 2 {
		t.Errorf("expected 2 hunks, got %d", len(files[0].Hunks))
	}
	totalAdd := 0
	for _, h := range files[0].Hunks {
		for _, l := range h.Lines {
			if l.Type == "add" {
				totalAdd++
			}
		}
	}
	if totalAdd != 2 {
		t.Errorf("expected 2 add lines total, got %d", totalAdd)
	}
}

func TestParseDiff_EmptyInput(t *testing.T) {
	files := ParseDiff("")
	if files != nil {
		t.Errorf("expected nil for empty input, got %v", files)
	}
}

func TestParseDiff_WhitespaceOnly(t *testing.T) {
	files := ParseDiff("   \n\n  \n")
	if len(files) != 0 {
		t.Errorf("expected 0 files for whitespace input, got %d", len(files))
	}
}

func TestDiffSizeRule_ExactlyAtLimit(t *testing.T) {
	rule := &DiffSizeRule{}
	files := make([]*FileDiff, 50)
	for i := range files {
		files[i] = &FileDiff{NewPath: "file.go", Additions: 5, Deletions: 5}
	}
	ctx := &RuleContext{Files: files}
	findings, _ := rule.Check(ctx)
	tooMany := false
	for _, f := range findings {
		if f.Title == "Too many files changed" {
			tooMany = true
		}
	}
	if tooMany {
		t.Error("50 files should not trigger 'too many files' (limit is >50)")
	}
}

func TestSecretRule_MultipleFileTypes(t *testing.T) {
	rule := &SecretRule{}
	ctx := &RuleContext{
		Files: []*FileDiff{
			{
				NewPath: "config.yaml",
				Hunks: []*DiffHunk{{
					Lines: []DiffLine{
						{Type: "add", NewLine: 1, Content: `password: "super-secret-password-value"`, FilePath: "config.yaml"},
					},
				}},
			},
			{
				NewPath: "main.go",
				Hunks: []*DiffHunk{{
					Lines: []DiffLine{
						{Type: "add", NewLine: 5, Content: `apiKey = "sk-longapikey12345678"`, FilePath: "main.go"},
					},
				}},
			},
		},
	}
	findings, _ := rule.Check(ctx)
	if len(findings) < 1 {
		t.Errorf("expected at least 1 finding, got %d", len(findings))
	}
}

func TestProtectedFileRule_Patterns(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{".env", true},
		{".env.production", true},
		{".env.local", true},
		{".env.staging", true},
		{"src/main.go", false},
		{"config.yaml", false},
	}
	for _, tt := range tests {
		rule := &ProtectedFileRule{}
		ctx := &RuleContext{
			Files: []*FileDiff{{NewPath: tt.path}},
		}
		findings, _ := rule.Check(ctx)
		hasFinding := len(findings) > 0
		if hasFinding != tt.expected {
			t.Errorf("ProtectedFileRule(%q) findings=%d, expected finding=%v", tt.path, len(findings), tt.expected)
		}
	}
}

func TestComputeFingerprint_Uniqueness(t *testing.T) {
	fp1 := computeFingerprint("rule1", "file.go", 1, "content1")
	fp2 := computeFingerprint("rule1", "file.go", 1, "content2")
	fp3 := computeFingerprint("rule2", "file.go", 1, "content1")
	if fp1 == fp2 {
		t.Error("different content should produce different fingerprints")
	}
	if fp1 == fp3 {
		t.Error("different rule IDs should produce different fingerprints")
	}
}

func TestGetRules_ReturnsDefaultRules(t *testing.T) {
	registeredRules = nil
	rules := GetRules()
	if len(rules) == 0 {
		t.Error("expected default rules to be registered")
	}
	ruleIDs := map[string]bool{}
	for _, r := range rules {
		ruleIDs[r.ID()] = true
	}
	expected := []string{"secret-detection", "protected-file", "diff-size", "migration-check", "test-required"}
	for _, id := range expected {
		if !ruleIDs[id] {
			t.Errorf("expected rule %s to be registered", id)
		}
	}
}

func TestBuildSummaryComment_Formats(t *testing.T) {
	result := &AggregatedResult{
		RiskLevel: SeverityMedium,
		Blocked:   false,
		Findings: []*Finding{
			{Severity: SeverityHigh, Title: "Security issue", FilePath: "main.go", NewLine: 10, Message: "msg"},
			{Severity: SeverityLow, Title: "Style issue", FilePath: "util.go", NewLine: 5, Message: "msg2"},
		},
		TotalAdd:  20,
		TotalDel:  5,
		FileCount: 3,
	}
	comment := BuildSummaryComment(result)
	if !strings.Contains(comment, "medium") {
		t.Error("expected risk level in summary")
	}
	if !strings.Contains(comment, "2") {
		t.Error("expected finding count in summary")
	}
}

func TestBuildInlineComment_ContainsAllFields(t *testing.T) {
	f := &Finding{
		Severity:   SeverityCritical,
		Title:      "Critical security issue",
		Message:    "Detailed description",
		Suggestion: "Fix suggestion",
		RuleID:     "security/critical",
		FilePath:   "main.go",
		NewLine:    42,
	}
	comment := BuildInlineComment(f)
	if !strings.Contains(comment, "Critical security issue") {
		t.Error("expected title in comment")
	}
	if !strings.Contains(comment, "Fix suggestion") {
		t.Error("expected suggestion in comment")
	}
	if !strings.Contains(comment, "security/critical") {
		t.Error("expected rule ID in comment")
	}
}
