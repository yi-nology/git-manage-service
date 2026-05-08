package codereview

import (
	"testing"
)

func TestParseDiff_Empty(t *testing.T) {
	result := ParseDiff("")
	if result != nil {
		t.Errorf("expected nil for empty input, got %v", result)
	}
}

func TestParseDiff_SingleFile(t *testing.T) {
	raw := "diff --git a/main.go b/main.go\n--- a/main.go\n+++ b/main.go\n@@ -10,6 +10,8 @@ func main() {\n import \"fmt\"\n import \"os\"\n \n+password := \"super-secret-key\"\n+apiKey := \"ak-1234567890abcdef\"\n fmt.Println(\"hello\")\n fmt.Println(\"world\")\n }\n"
	files := ParseDiff(raw)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	f := files[0]
	if f.OldPath != "main.go" || f.NewPath != "main.go" {
		t.Errorf("paths: old=%s new=%s", f.OldPath, f.NewPath)
	}
	if f.Additions != 2 {
		t.Errorf("expected 2 additions, got %d", f.Additions)
	}
	if len(f.Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(f.Hunks))
	}
	h := f.Hunks[0]
	addLines := 0
	for _, l := range h.Lines {
		if l.Type == "add" {
			addLines++
		}
	}
	if addLines != 2 {
		t.Fatalf("expected 2 add lines, got %d (total lines: %d)", addLines, len(h.Lines))
	}
	found := false
	for _, l := range h.Lines {
		if l.Type == "add" && l.Content == `password := "super-secret-key"` {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("password line not found in add lines")
	}
}

func TestParseDiff_NewFile(t *testing.T) {
	raw := `diff --git a/newdir/config.yaml b/newdir/config.yaml
new file mode 100644
--- /dev/null
+++ b/newdir/config.yaml
@@ -0,0 +1,3 @@
+server:
+  port: 8080
+  host: localhost
`
	files := ParseDiff(raw)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if !files[0].IsNew {
		t.Error("expected IsNew=true")
	}
	if files[0].NewPath != "newdir/config.yaml" {
		t.Errorf("path: %s", files[0].NewPath)
	}
}

func TestParseDiff_DeletedFile(t *testing.T) {
	raw := `diff --git a/old.go b/old.go
deleted file mode 100644
--- a/old.go
+++ /dev/null
@@ -1,2 +0,0 @@
-package main
-func main() {}
`
	files := ParseDiff(raw)
	if len(files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(files))
	}
	if !files[0].IsDeleted {
		t.Error("expected IsDeleted=true")
	}
}

func TestParseDiff_MultipleFiles(t *testing.T) {
	raw := `diff --git a/a.go b/a.go
--- a/a.go
+++ b/a.go
@@ -1,3 +1,4 @@
 package a
+import "fmt"
 func A() {}
 
diff --git a/b.go b/b.go
--- a/b.go
+++ b/b.go
@@ -1,3 +1,4 @@
 package b
+import "os"
 func B() {}
`
	files := ParseDiff(raw)
	if len(files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(files))
	}
}

func TestSecretRule_DetectsPassword(t *testing.T) {
	rule := &SecretRule{}
	ctx := &RuleContext{
		Files: []*FileDiff{
			{
				NewPath: "main.go",
				Hunks: []*DiffHunk{{
					Lines: []DiffLine{
						{Type: "add", NewLine: 10, Content: `password = "my-super-secret-pass"`, FilePath: "main.go"},
						{Type: "add", NewLine: 11, Content: `apiKey = "ak-1234567890abcdef12345678"`, FilePath: "main.go"},
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
		t.Fatal("expected at least 1 finding")
	}
	if findings[0].Severity != SeverityCritical {
		t.Errorf("expected critical severity, got %s", findings[0].Severity)
	}
}

func TestSecretRule_SkipsFalsePositive(t *testing.T) {
	rule := &SecretRule{}
	ctx := &RuleContext{
		Files: []*FileDiff{
			{
				NewPath: "example_test.go",
				Hunks: []*DiffHunk{{
					Lines: []DiffLine{
						{Type: "add", NewLine: 5, Content: `password := "changeme"`, FilePath: "example_test.go"},
						{Type: "add", NewLine: 6, Content: `apiKey := "your_api_key_here"`, FilePath: "example_test.go"},
					},
				}},
			},
		},
	}
	findings, _ := rule.Check(ctx)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for false positives, got %d", len(findings))
	}
}

func TestSecretRule_IgnoresContextAndDeleteLines(t *testing.T) {
	rule := &SecretRule{}
	ctx := &RuleContext{
		Files: []*FileDiff{
			{
				NewPath: "main.go",
				Hunks: []*DiffHunk{{
					Lines: []DiffLine{
						{Type: "ctx", Content: `password := "should-not-detect"`, FilePath: "main.go"},
						{Type: "del", Content: `api_key := "should-not-detect"`, FilePath: "main.go"},
					},
				}},
			},
		},
	}
	findings, _ := rule.Check(ctx)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings for non-add lines, got %d", len(findings))
	}
}

func TestProtectedFileRule_EnvFile(t *testing.T) {
	rule := &ProtectedFileRule{}
	ctx := &RuleContext{
		Files: []*FileDiff{
			{NewPath: ".env"},
			{NewPath: "src/main.go"},
			{NewPath: ".env.production"},
		},
	}
	findings, err := rule.Check(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 2 {
		t.Errorf("expected 2 findings (.env + .env.production), got %d", len(findings))
	}
}

func TestDiffSizeRule_LargeMR(t *testing.T) {
	rule := &DiffSizeRule{}
	files := make([]*FileDiff, 60)
	for i := range files {
		files[i] = &FileDiff{NewPath: "file.go", Additions: 10, Deletions: 5}
	}
	ctx := &RuleContext{Files: files}
	findings, err := rule.Check(ctx)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range findings {
		if f.Title == "Too many files changed" {
			found = true
		}
	}
	if !found {
		t.Error("expected 'too many files' finding")
	}
}

func TestTestRequiredRule_NoTest(t *testing.T) {
	rule := &TestRequiredRule{}
	ctx := &RuleContext{
		Files: []*FileDiff{
			{NewPath: "src/service/user.go", Additions: 10},
		},
	}
	findings, err := rule.Check(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) == 0 {
		t.Fatal("expected finding for missing test")
	}
	if findings[0].Severity != SeverityLow {
		t.Errorf("expected low severity, got %s", findings[0].Severity)
	}
}

func TestTestRequiredRule_WithTest(t *testing.T) {
	rule := &TestRequiredRule{}
	ctx := &RuleContext{
		Files: []*FileDiff{
			{NewPath: "src/service/user.go", Additions: 10},
			{NewPath: "src/service/user_test.go", Additions: 10},
		},
	}
	findings, _ := rule.Check(ctx)
	if len(findings) != 0 {
		t.Errorf("expected 0 findings when test exists, got %d", len(findings))
	}
}

func TestAggregator_Deduplication(t *testing.T) {
	findings := []*Finding{
		{Fingerprint: "abc", Severity: SeverityHigh, Title: "A"},
		{Fingerprint: "abc", Severity: SeverityHigh, Title: "A (dup)"},
		{Fingerprint: "def", Severity: SeverityLow, Title: "B"},
	}
	result := Aggregate(findings, 10, 5, 3, true, nil)
	if len(result.Findings) != 2 {
		t.Errorf("expected 2 deduplicated findings, got %d", len(result.Findings))
	}
}

func TestAggregator_BlockedOnHigh(t *testing.T) {
	findings := []*Finding{
		{Fingerprint: "abc", Severity: SeverityCritical, Title: "Critical issue"},
	}
	result := Aggregate(findings, 10, 5, 3, true, nil)
	if !result.Blocked {
		t.Error("expected blocked=true")
	}
	if result.RiskLevel != SeverityCritical {
		t.Errorf("expected critical risk, got %s", result.RiskLevel)
	}
}

func TestAggregator_PassesOnLow(t *testing.T) {
	findings := []*Finding{
		{Fingerprint: "abc", Severity: SeverityLow, Title: "Minor issue"},
	}
	result := Aggregate(findings, 10, 5, 3, true, nil)
	if result.Blocked {
		t.Error("expected blocked=false for low severity")
	}
}

func TestAggregator_NoBlockWhenDisabled(t *testing.T) {
	findings := []*Finding{
		{Fingerprint: "abc", Severity: SeverityCritical, Title: "Critical issue"},
	}
	result := Aggregate(findings, 10, 5, 3, false, nil)
	if result.Blocked {
		t.Error("expected blocked=false when blockOnHigh=false")
	}
}

func TestComputeFingerprint_Deterministic(t *testing.T) {
	fp1 := computeFingerprint("rule1", "file.go", 10, "content")
	fp2 := computeFingerprint("rule1", "file.go", 10, "content")
	if fp1 != fp2 {
		t.Error("expected deterministic fingerprints")
	}
	fp3 := computeFingerprint("rule1", "file.go", 10, "different")
	if fp1 == fp3 {
		t.Error("expected different fingerprints for different content")
	}
}

func TestBuildSummaryComment(t *testing.T) {
	result := &AggregatedResult{
		RiskLevel: SeverityMedium,
		Blocked:   false,
		Findings: []*Finding{
			{Severity: SeverityHigh, Title: "Test issue", FilePath: "main.go", NewLine: 10, Message: "msg"},
		},
		TotalAdd:  20,
		TotalDel:  5,
		FileCount: 3,
	}
	comment := BuildSummaryComment(result)
	if comment == "" {
		t.Fatal("expected non-empty comment")
	}
}

func TestBuildInlineComment(t *testing.T) {
	f := &Finding{
		Severity:   SeverityHigh,
		Title:      "SQL injection risk",
		Message:    "User input directly concatenated into query",
		Suggestion: "Use parameterized queries",
		RuleID:     "security/sql-injection",
	}
	comment := BuildInlineComment(f)
	if comment == "" {
		t.Fatal("expected non-empty inline comment")
	}
}
