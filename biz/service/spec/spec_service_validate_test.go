package spec

import (
	"strings"
	"testing"
)

func TestValidateSpec_MissingName(t *testing.T) {
	s := NewSpecService()
	content := "Version: 1.0\nRelease: 1\nSummary: test\nLicense: MIT\n"
	result := s.ValidateSpec(content)
	if result.Valid {
		t.Error("expected invalid for missing Name")
	}
	found := false
	for _, issue := range result.Issues {
		if strings.Contains(issue.Message, "Name") {
			found = true
		}
	}
	if !found {
		t.Error("expected issue about missing Name")
	}
}

func TestValidateSpec_MissingVersion(t *testing.T) {
	s := NewSpecService()
	content := "Name: foo\nRelease: 1\nSummary: test\nLicense: MIT\n"
	result := s.ValidateSpec(content)
	if result.Valid {
		t.Error("expected invalid for missing Version")
	}
}

func TestValidateSpec_MissingRelease(t *testing.T) {
	s := NewSpecService()
	content := "Name: foo\nVersion: 1.0\nSummary: test\nLicense: MIT\n"
	result := s.ValidateSpec(content)
	if result.Valid {
		t.Error("expected invalid for missing Release")
	}
}

func TestValidateSpec_MissingSummary(t *testing.T) {
	s := NewSpecService()
	content := "Name: foo\nVersion: 1.0\nRelease: 1\nLicense: MIT\n"
	result := s.ValidateSpec(content)
	if result.Valid {
		t.Error("expected invalid for missing Summary")
	}
}

func TestValidateSpec_MissingLicense(t *testing.T) {
	s := NewSpecService()
	content := "Name: foo\nVersion: 1.0\nRelease: 1\nSummary: test\n"
	result := s.ValidateSpec(content)
	if result.Valid {
		t.Error("expected invalid for missing License")
	}
}

func TestValidateSpec_AllRequired(t *testing.T) {
	s := NewSpecService()
	content := "Name: foo\nVersion: 1.0\nRelease: 1\nSummary: test\nLicense: MIT\n"
	result := s.ValidateSpec(content)
	if !result.Valid {
		t.Errorf("expected valid spec, got issues: %v", result.Issues)
	}
}

func TestValidateSpec_ChangelogFormat(t *testing.T) {
	s := NewSpecService()
	content := "Name: foo\nVersion: 1.0\nRelease: 1\nSummary: test\nLicense: MIT\n%changelog\nMon Jan 1 - test\n- update\n"
	result := s.ValidateSpec(content)
	changelogIssue := false
	for _, issue := range result.Issues {
		if strings.Contains(issue.Message, "Changelog") {
			changelogIssue = true
		}
	}
	if !changelogIssue {
		t.Error("expected changelog format issue")
	}
}

func TestValidateSpec_ChangelogCorrect(t *testing.T) {
	s := NewSpecService()
	content := "Name: foo\nVersion: 1.0\nRelease: 1\nSummary: test\nLicense: MIT\n%changelog\n* Mon Jan 1 2024 Author\n- update\n"
	result := s.ValidateSpec(content)
	for _, issue := range result.Issues {
		if strings.Contains(issue.Message, "Changelog") {
			t.Error("should not have changelog issue for correct format")
		}
	}
}

func TestValidateSpec_NoTabs(t *testing.T) {
	s := NewSpecService()
	content := "Name:\tfoo\nVersion: 1.0\nRelease: 1\nSummary: test\nLicense: MIT\n"
	result := s.ValidateSpec(content)
	tabIssue := false
	for _, issue := range result.Issues {
		if strings.Contains(issue.Message, "tab") {
			tabIssue = true
		}
	}
	if !tabIssue {
		t.Error("expected tab issue")
	}
}

func TestValidateSpec_StatsPopulated(t *testing.T) {
	s := NewSpecService()
	content := "Name: foo\nVersion: 1.0\nRelease: 1\nSummary: test\nLicense: MIT\n"
	result := s.ValidateSpec(content)
	if result.Stats == nil {
		t.Fatal("expected stats to be populated")
	}
	if result.Stats["total_lines"] == "" {
		t.Error("expected total_lines in stats")
	}
}

func TestGetBuiltinRules_NonEmpty(t *testing.T) {
	s := NewSpecService()
	rules := s.GetBuiltinRules()
	if len(rules) == 0 {
		t.Error("expected non-empty builtin rules")
	}
}

func TestGetBuiltinRules_RequiredRules(t *testing.T) {
	s := NewSpecService()
	rules := s.GetBuiltinRules()
	patterns := map[string]bool{}
	for _, r := range rules {
		patterns[r.Pattern] = true
	}
	required := []string{"required_name", "required_version", "required_release", "required_summary", "required_license"}
	for _, p := range required {
		if !patterns[p] {
			t.Errorf("missing required rule pattern: %s", p)
		}
	}
}

func TestGetBuiltinRules_EnabledByDefault(t *testing.T) {
	s := NewSpecService()
	rules := s.GetBuiltinRules()
	for _, r := range rules {
		if r.Pattern == "required_name" && !r.Enabled {
			t.Error("required_name should be enabled by default")
		}
	}
}

func TestHasSection(t *testing.T) {
	s := &SpecService{}
	lines := []string{"Name: foo", "Version: 1.0"}
	if !s.hasSection(lines, "Name:") {
		t.Error("expected to find Name:")
	}
	if s.hasSection(lines, "URL:") {
		t.Error("should not find URL:")
	}
}

func TestApplyRule_CustomRegex(t *testing.T) {
	s := &SpecService{}
	rule := SpecRule{
		ID:          "test-custom",
		Name:        "Custom Test",
		Pattern:     "(?i)^BuildRoot:",
		Severity:    "warning",
		Enabled:     true,
		Description: "Test custom regex",
	}
	lines := []string{"BuildRoot: /tmp/test"}
	issues := s.applyRule(lines, rule)
	if len(issues) == 0 {
		t.Error("expected custom regex to match")
	}
}

func TestValidateSpec_DisabledRules(t *testing.T) {
	s := NewSpecService()
	rules := s.GetBuiltinRules()
	for _, r := range rules {
		if r.Pattern == "buildroot-usage" && r.Enabled {
			t.Error("buildroot-usage should be disabled by default")
		}
	}
}
