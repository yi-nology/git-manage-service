package lint

import (
	"testing"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func TestHasField(t *testing.T) {
	s := &LintService{}
	lines := []string{"Name: foo", "Version: 1.0", "Release: 1"}
	if !s.hasField(lines, "Name:") {
		t.Error("expected to find Name:")
	}
	if s.hasField(lines, "URL:") {
		t.Error("should not find URL:")
	}
}

func TestHasField_WithWhitespace(t *testing.T) {
	s := &LintService{}
	lines := []string{"  Name: foo"}
	if !s.hasField(lines, "Name:") {
		t.Error("expected to find Name: with leading whitespace")
	}
}

func TestHasSection(t *testing.T) {
	s := &LintService{}
	lines := []string{"Name: foo", "%description", "test", "%build", "make"}
	if !s.hasSection(lines, "%description") {
		t.Error("expected %description section")
	}
	if !s.hasSection(lines, "%build") {
		t.Error("expected %build section")
	}
	if s.hasSection(lines, "%install") {
		t.Error("should not find %install")
	}
}

func TestFindSectionLine(t *testing.T) {
	s := &LintService{}
	lines := []string{"Name: foo", "%description", "test"}
	line := s.findSectionLine(lines, "%description")
	if line != 2 {
		t.Errorf("expected line 2 (1-indexed), got %d", line)
	}
}

func TestFindSectionLine_NotFound(t *testing.T) {
	s := &LintService{}
	lines := []string{"Name: foo"}
	line := s.findSectionLine(lines, "%build")
	if line != 0 {
		t.Errorf("expected 0 for not found, got %d", line)
	}
}

func TestApplyRule_RequiredName(t *testing.T) {
	s := &LintService{}
	rule := po.LintRule{ID: "test", Pattern: "required_name", Severity: "error", Enabled: true}
	issues := s.applyRule([]string{"Version: 1.0"}, rule)
	if len(issues) == 0 {
		t.Error("expected issue for missing Name")
	}
	issues = s.applyRule([]string{"Name: foo"}, rule)
	if len(issues) != 0 {
		t.Error("expected no issue when Name present")
	}
}

func TestApplyRule_RequiredVersion(t *testing.T) {
	s := &LintService{}
	rule := po.LintRule{ID: "test", Pattern: "required_version", Severity: "error", Enabled: true}
	issues := s.applyRule([]string{"Name: foo"}, rule)
	if len(issues) == 0 {
		t.Error("expected issue for missing Version")
	}
}

func TestApplyRule_RequiredRelease(t *testing.T) {
	s := &LintService{}
	rule := po.LintRule{ID: "test", Pattern: "required_release", Severity: "error", Enabled: true}
	issues := s.applyRule([]string{"Name: foo"}, rule)
	if len(issues) == 0 {
		t.Error("expected issue for missing Release")
	}
}

func TestApplyRule_RequiredSummary(t *testing.T) {
	s := &LintService{}
	rule := po.LintRule{ID: "test", Pattern: "required_summary", Severity: "error", Enabled: true}
	issues := s.applyRule([]string{"Name: foo"}, rule)
	if len(issues) == 0 {
		t.Error("expected issue for missing Summary")
	}
}

func TestApplyRule_RequiredLicense(t *testing.T) {
	s := &LintService{}
	rule := po.LintRule{ID: "test", Pattern: "required_license", Severity: "error", Enabled: true}
	issues := s.applyRule([]string{"Name: foo"}, rule)
	if len(issues) == 0 {
		t.Error("expected issue for missing License")
	}
}

func TestApplyRule_RequiredURL(t *testing.T) {
	s := &LintService{}
	rule := po.LintRule{ID: "test", Pattern: "required_url", Severity: "warning", Enabled: true}
	issues := s.applyRule([]string{"Name: foo"}, rule)
	if len(issues) == 0 {
		t.Error("expected issue for missing URL")
	}
}

func TestApplyRule_RequiredDescription(t *testing.T) {
	s := &LintService{}
	rule := po.LintRule{ID: "test", Pattern: "required_description", Severity: "error", Enabled: true}
	issues := s.applyRule([]string{"Name: foo"}, rule)
	if len(issues) == 0 {
		t.Error("expected issue for missing %description")
	}
}

func TestApplyRule_RequiredPrep(t *testing.T) {
	s := &LintService{}
	rule := po.LintRule{ID: "test", Pattern: "required_prep", Severity: "warning", Enabled: true}
	issues := s.applyRule([]string{"Name: foo"}, rule)
	if len(issues) == 0 {
		t.Error("expected issue for missing %prep")
	}
}

func TestApplyRule_RequiredBuild(t *testing.T) {
	s := &LintService{}
	rule := po.LintRule{ID: "test", Pattern: "required_build", Severity: "warning", Enabled: true}
	issues := s.applyRule([]string{"Name: foo"}, rule)
	if len(issues) == 0 {
		t.Error("expected issue for missing %build")
	}
}

func TestApplyRule_RequiredInstall(t *testing.T) {
	s := &LintService{}
	rule := po.LintRule{ID: "test", Pattern: "required_install", Severity: "warning", Enabled: true}
	issues := s.applyRule([]string{"Name: foo"}, rule)
	if len(issues) == 0 {
		t.Error("expected issue for missing %install")
	}
}

func TestApplyRule_RequiredFiles(t *testing.T) {
	s := &LintService{}
	rule := po.LintRule{ID: "test", Pattern: "required_files", Severity: "error", Enabled: true}
	issues := s.applyRule([]string{"Name: foo"}, rule)
	if len(issues) == 0 {
		t.Error("expected issue for missing %files")
	}
}

func TestApplyRule_EmptySections(t *testing.T) {
	s := &LintService{}
	rule := po.LintRule{ID: "test", Pattern: "empty_sections", Severity: "warning", Enabled: true}
	content := []string{
		"Name: foo",
		"%description",
		"%build",
		"%install",
	}
	issues := s.applyRule(content, rule)
	if len(issues) < 2 {
		t.Errorf("expected issues for empty sections, got %d", len(issues))
	}
}

func TestApplyRule_BuildRootUsage(t *testing.T) {
	s := &LintService{}
	rule := po.LintRule{ID: "test", Pattern: "buildroot_usage", Severity: "warning", Enabled: true}
	issues := s.applyRule([]string{"BuildRoot: /tmp/foo"}, rule)
	if len(issues) == 0 {
		t.Error("expected issue for BuildRoot without %{_tmppath}")
	}
	issues = s.applyRule([]string{"BuildRoot: %{_tmppath}/%{name}"}, rule)
	if len(issues) != 0 {
		t.Error("expected no issue when BuildRoot uses %{_tmppath}")
	}
}

func TestApplyRule_NoTabs(t *testing.T) {
	s := &LintService{}
	rule := po.LintRule{ID: "test", Pattern: "no_tabs", Severity: "info", Enabled: true}
	issues := s.applyRule([]string{"Name:\tfoo"}, rule)
	if len(issues) == 0 {
		t.Error("expected issue for tab usage")
	}
	issues = s.applyRule([]string{"Name: foo"}, rule)
	if len(issues) != 0 {
		t.Error("expected no issue when no tabs")
	}
}

func TestApplyRule_ChangelogFormat(t *testing.T) {
	s := &LintService{}
	rule := po.LintRule{ID: "test", Pattern: "changelog_format", Severity: "warning", Enabled: true}
	lines := []string{"%changelog", "Mon Jan 1 2024", "- update"}
	issues := s.applyRule(lines, rule)
	if len(issues) == 0 {
		t.Error("expected issue for changelog not starting with *")
	}
	lines2 := []string{"%changelog", "* Mon Jan 1 2024 Author", "- update"}
	issues = s.applyRule(lines2, rule)
	if len(issues) != 0 {
		t.Error("expected no issue for correct changelog format")
	}
}

func TestApplyRule_CustomRegex(t *testing.T) {
	s := &LintService{}
	rule := po.LintRule{ID: "test", Pattern: "(?i)^BuildRoot:", Severity: "warning", Enabled: true}
	issues := s.applyRule([]string{"BuildRoot: /tmp/test"}, rule)
	if len(issues) == 0 {
		t.Error("expected regex match on BuildRoot")
	}
	if issues[0].Column == 0 {
		t.Error("expected non-zero column for regex match")
	}
}

func TestApplyRule_DisabledRule(t *testing.T) {
	s := &LintService{}
	rule := po.LintRule{ID: "test", Pattern: "required_name", Severity: "error", Enabled: false}
	issues := s.applyRule([]string{}, rule)
	if len(issues) != 0 {
		t.Error("disabled rule should produce no issues")
	}
}
