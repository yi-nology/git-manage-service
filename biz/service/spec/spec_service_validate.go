package spec

import (
	"fmt"
	"regexp"
	"strings"
)

func (s *SpecService) ValidateSpec(content string) SpecValidationResult {
	rules := s.GetBuiltinRules()
	var issues []SpecIssue
	var warnings []SpecIssue

	lines := strings.Split(content, "\n")

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		issues = append(issues, s.applyRule(lines, rule)...)
	}

	for _, issue := range issues {
		if issue.Severity == "error" {
			issues = append(issues, issue)
		} else {
			warnings = append(warnings, issue)
		}
	}

	return SpecValidationResult{
		Valid:    len(issues) == 0,
		Issues:   issues,
		Warnings: warnings,
		Stats: map[string]string{
			"total_lines": fmt.Sprintf("%d", len(lines)),
			"errors":      fmt.Sprintf("%d", len(issues)),
			"warnings":    fmt.Sprintf("%d", len(warnings)),
		},
	}
}

func (s *SpecService) applyRule(lines []string, rule SpecRule) []SpecIssue {
	var issues []SpecIssue

	switch rule.Pattern {
	case "required_name":
		if !s.hasSection(lines, "Name:") {
			issues = append(issues, SpecIssue{
				Line:     1,
				Message:  "Missing required field: Name",
				Severity: rule.Severity,
				Rule:     rule.ID,
				RuleDesc: rule.Description,
			})
		}

	case "required_version":
		if !s.hasSection(lines, "Version:") {
			issues = append(issues, SpecIssue{
				Line:     1,
				Message:  "Missing required field: Version",
				Severity: rule.Severity,
				Rule:     rule.ID,
				RuleDesc: rule.Description,
			})
		}

	case "required_release":
		if !s.hasSection(lines, "Release:") {
			issues = append(issues, SpecIssue{
				Line:     1,
				Message:  "Missing required field: Release",
				Severity: rule.Severity,
				Rule:     rule.ID,
				RuleDesc: rule.Description,
			})
		}

	case "required_summary":
		if !s.hasSection(lines, "Summary:") {
			issues = append(issues, SpecIssue{
				Line:     1,
				Message:  "Missing required field: Summary",
				Severity: rule.Severity,
				Rule:     rule.ID,
				RuleDesc: rule.Description,
			})
		}

	case "required_license":
		if !s.hasSection(lines, "License:") {
			issues = append(issues, SpecIssue{
				Line:     1,
				Message:  "Missing required field: License",
				Severity: rule.Severity,
				Rule:     rule.ID,
				RuleDesc: rule.Description,
			})
		}

	case "changelog_format":
		for i, line := range lines {
			if strings.HasPrefix(line, "%changelog") {
				if i+1 < len(lines) && !strings.HasPrefix(lines[i+1], "*") {
					issues = append(issues, SpecIssue{
						Line:     i + 2,
						Message:  "Changelog entry should start with '*'",
						Severity: rule.Severity,
						Rule:     rule.ID,
						RuleDesc: rule.Description,
					})
				}
			}
		}

	case "no_tabs":
		for i, line := range lines {
			if strings.Contains(line, "\t") {
				issues = append(issues, SpecIssue{
					Line:     i + 1,
					Message:  "Avoid using tabs, use spaces instead",
					Severity: rule.Severity,
					Rule:     rule.ID,
					RuleDesc: rule.Description,
				})
			}
		}

	default:
		if rule.Pattern != "" {
			re, err := regexp.Compile(rule.Pattern)
			if err == nil {
				for i, line := range lines {
					if re.MatchString(line) {
						issues = append(issues, SpecIssue{
							Line:     i + 1,
							Message:  fmt.Sprintf("Line matches rule: %s", rule.Name),
							Severity: rule.Severity,
							Rule:     rule.ID,
							RuleDesc: rule.Description,
						})
					}
				}
			}
		}
	}

	return issues
}

func (s *SpecService) hasSection(lines []string, prefix string) bool {
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), prefix) {
			return true
		}
	}
	return false
}

func (s *SpecService) GetBuiltinRules() []SpecRule {
	return []SpecRule{
		{
			ID:          "required-name",
			Name:        "Required: Name",
			Description: "Spec file must have a Name field",
			Severity:    "error",
			Pattern:     "required_name",
			Enabled:     true,
			Category:    "required",
		},
		{
			ID:          "required-version",
			Name:        "Required: Version",
			Description: "Spec file must have a Version field",
			Severity:    "error",
			Pattern:     "required_version",
			Enabled:     true,
			Category:    "required",
		},
		{
			ID:          "required-release",
			Name:        "Required: Release",
			Description: "Spec file must have a Release field",
			Severity:    "error",
			Pattern:     "required_release",
			Enabled:     true,
			Category:    "required",
		},
		{
			ID:          "required-summary",
			Name:        "Required: Summary",
			Description: "Spec file must have a Summary field",
			Severity:    "error",
			Pattern:     "required_summary",
			Enabled:     true,
			Category:    "required",
		},
		{
			ID:          "required-license",
			Name:        "Required: License",
			Description: "Spec file must have a License field",
			Severity:    "error",
			Pattern:     "required_license",
			Enabled:     true,
			Category:    "required",
		},

		{
			ID:          "changelog-format",
			Name:        "Changelog Format",
			Description: "Changelog entries should follow RPM format",
			Severity:    "warning",
			Pattern:     "changelog_format",
			Enabled:     true,
			Category:    "style",
		},
		{
			ID:          "no-tabs",
			Name:        "No Tabs",
			Description: "Use spaces instead of tabs for consistency",
			Severity:    "info",
			Pattern:     "no_tabs",
			Enabled:     true,
			Category:    "style",
		},

		{
			ID:          "buildroot-usage",
			Name:        "BuildRoot Usage",
			Description: "BuildRoot is deprecated in modern RPM",
			Severity:    "warning",
			Pattern:     "(?i)^BuildRoot:",
			Enabled:     false,
			Category:    "best-practice",
		},
		{
			ID:          "defattr-usage",
			Name:        "%defattr Usage",
			Description: "%defattr is usually not needed in modern RPM",
			Severity:    "info",
			Pattern:     "%defattr",
			Enabled:     false,
			Category:    "best-practice",
		},
	}
}
