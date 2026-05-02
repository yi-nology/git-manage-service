package lint

import (
	"testing"
)

func TestExtractJSON_ValidJSON(t *testing.T) {
	input := `{"issues": [{"line": 1, "severity": "error", "message": "test"}]}`
	result := extractJSON(input)
	if result == "" {
		t.Error("expected non-empty JSON extraction")
	}
}

func TestExtractJSON_JSONInMarkdown(t *testing.T) {
	input := "Here is the result:\n```json\n{\"issues\": []}\n```\nDone."
	result := extractJSON(input)
	if result == "" {
		t.Error("expected JSON extraction from markdown")
	}
}

func TestExtractJSON_NoJSON(t *testing.T) {
	input := "no json here just plain text"
	result := extractJSON(input)
	if result != "" {
		t.Errorf("expected empty extraction, got %q", result)
	}
}

func TestExtractJSON_Empty(t *testing.T) {
	result := extractJSON("")
	if result != "" {
		t.Errorf("expected empty, got %q", result)
	}
}

func TestExtractJSON_NestedBraces(t *testing.T) {
	input := `{"outer": {"inner": "value"}}`
	result := extractJSON(input)
	if result != input {
		t.Errorf("expected full extraction, got %q", result)
	}
}

func TestParseAILintResponse_Valid(t *testing.T) {
	raw := `{"issues": [{"line": 5, "severity": "error", "message": "Missing BuildRequires", "quick_fix": "Add BuildRequires: gcc"}]}`
	result, err := parseAILintResponse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(result.Issues))
	}
	if result.Issues[0].RuleID != "ai-lint" {
		t.Errorf("expected ruleId ai-lint, got %s", result.Issues[0].RuleID)
	}
	if result.Issues[0].Source != "ai" {
		t.Errorf("expected source ai, got %s", result.Issues[0].Source)
	}
	if result.Issues[0].Line != 5 {
		t.Errorf("expected line 5, got %d", result.Issues[0].Line)
	}
	if result.Stats.ErrorCount != 1 {
		t.Errorf("expected 1 error, got %d", result.Stats.ErrorCount)
	}
}

func TestParseAILintResponse_MultipleIssues(t *testing.T) {
	raw := `{
		"issues": [
			{"line": 1, "severity": "error", "message": "err1", "quick_fix": "fix1"},
			{"line": 2, "severity": "warning", "message": "warn1", "quick_fix": "fix2"},
			{"line": 3, "severity": "info", "message": "info1", "quick_fix": "fix3"}
		]
	}`
	result, _ := parseAILintResponse(raw)
	if result.Stats.ErrorCount != 1 || result.Stats.WarningCount != 1 || result.Stats.InfoCount != 1 {
		t.Errorf("stats mismatch: error=%d warning=%d info=%d",
			result.Stats.ErrorCount, result.Stats.WarningCount, result.Stats.InfoCount)
	}
}

func TestParseAILintResponse_EmptyIssues(t *testing.T) {
	raw := `{"issues": []}`
	result, err := parseAILintResponse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Issues) != 0 {
		t.Errorf("expected 0 issues, got %d", len(result.Issues))
	}
}

func TestParseAILintResponse_InvalidSeverity(t *testing.T) {
	raw := `{"issues": [{"line": 1, "severity": "unknown", "message": "test", "quick_fix": "fix"}]}`
	result, _ := parseAILintResponse(raw)
	if len(result.Issues) != 1 {
		t.Fatal("expected 1 issue")
	}
	if result.Issues[0].Severity != "info" {
		t.Errorf("unknown severity should default to info, got %s", result.Issues[0].Severity)
	}
}

func TestParseAILintResponse_NoJSON(t *testing.T) {
	raw := "just plain text no json"
	result, err := parseAILintResponse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Issues) != 0 {
		t.Errorf("expected 0 issues for non-JSON, got %d", len(result.Issues))
	}
}

func TestParseAILintResponse_InvalidJSON(t *testing.T) {
	raw := `{invalid json}`
	result, err := parseAILintResponse(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Issues) != 0 {
		t.Errorf("expected 0 issues for invalid JSON, got %d", len(result.Issues))
	}
}

func TestParseAILintResponse_WithQuickFix(t *testing.T) {
	raw := `{"issues": [{"line": 1, "severity": "error", "message": "test", "quick_fix": "Replace X with Y"}]}`
	result, _ := parseAILintResponse(raw)
	if len(result.Issues) != 1 {
		t.Fatal("expected 1 issue")
	}
	if result.Issues[0].QuickFix != "Replace X with Y" {
		t.Errorf("quick_fix mismatch: %s", result.Issues[0].QuickFix)
	}
}
