package lint

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func (s *LintService) applyRule(lines []string, rule po.LintRule) []LintIssue {
	var issues []LintIssue

	switch rule.Pattern {
	case "required_name":
		if !s.hasField(lines, "Name:") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing required field: Name",
				Line:     1,
			})
		}

	case "required_version":
		if !s.hasField(lines, "Version:") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing required field: Version",
				Line:     1,
			})
		}

	case "required_release":
		if !s.hasField(lines, "Release:") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing required field: Release",
				Line:     1,
			})
		}

	case "required_summary":
		if !s.hasField(lines, "Summary:") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing required field: Summary",
				Line:     1,
			})
		}

	case "required_license":
		if !s.hasField(lines, "License:") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing required field: License",
				Line:     1,
			})
		}

	case "required_url":
		if !s.hasField(lines, "URL:") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing recommended field: URL",
				Line:     1,
			})
		}

	case "required_description":
		if !s.hasSection(lines, "%description") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing required section: %description",
				Line:     1,
			})
		}

	case "required_prep":
		if !s.hasSection(lines, "%prep") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing recommended section: %prep",
				Line:     1,
			})
		}

	case "required_build":
		if !s.hasSection(lines, "%build") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing recommended section: %build",
				Line:     1,
			})
		}

	case "required_install":
		if !s.hasSection(lines, "%install") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing recommended section: %install",
				Line:     1,
			})
		}

	case "required_files":
		if !s.hasSection(lines, "%files") {
			issues = append(issues, LintIssue{
				RuleID:   rule.ID,
				Severity: rule.Severity,
				Message:  "Missing required section: %files",
				Line:     1,
			})
		}

	case "empty_sections":
		sections := []string{"%description", "%prep", "%build", "%install", "%files"}
		for _, section := range sections {
			sectionLine := s.findSectionLine(lines, section)
			if sectionLine > 0 {
				isEmpty := true
				for i := sectionLine; i < len(lines); i++ {
					line := strings.TrimSpace(lines[i])
					if line == "" {
						continue
					}
					if strings.HasPrefix(line, "%") && i > sectionLine {
						break
					}
					if !strings.HasPrefix(line, "%") {
						isEmpty = false
						break
					}
				}
				if isEmpty {
					issues = append(issues, LintIssue{
						RuleID:   rule.ID,
						Severity: rule.Severity,
						Message:  fmt.Sprintf("Section %s is empty", section),
						Line:     sectionLine + 1,
					})
				}
			}
		}

	case "buildroot_usage":
		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "BuildRoot:") {
				if !strings.Contains(line, "%{_tmppath}") {
					issues = append(issues, LintIssue{
						RuleID:   rule.ID,
						Severity: rule.Severity,
						Message:  "BuildRoot should use %{_tmppath} macro",
						Line:     i + 1,
					})
				}
			}
		}

	case "macro_consistency":
		reBraces := regexp.MustCompile(`%\{[a-zA-Z_][a-zA-Z0-9_]*\}`)
		reNoBraces := regexp.MustCompile(`%[a-zA-Z_][a-zA-Z0-9_]*[^{a-zA-Z0-9_]`)
		for i, line := range lines {
			braces := reBraces.FindAllString(line, -1)
			noBraces := reNoBraces.FindAllString(line, -1)
			if len(braces) > 0 && len(noBraces) > 0 {
				issues = append(issues, LintIssue{
					RuleID:   rule.ID,
					Severity: rule.Severity,
					Message:  "Inconsistent macro usage: use either %{macro} or %macro consistently",
					Line:     i + 1,
				})
				break
			}
		}

	case "changelog_format":
		for i, line := range lines {
			if strings.HasPrefix(line, "%changelog") {
				if i+1 < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i+1]), "*") {
					issues = append(issues, LintIssue{
						RuleID:   rule.ID,
						Severity: rule.Severity,
						Message:  "Changelog entry should start with '*'",
						Line:     i + 2,
					})
				}
			}
		}

	case "no_tabs":
		for i, line := range lines {
			if strings.Contains(line, "\t") {
				issues = append(issues, LintIssue{
					RuleID:   rule.ID,
					Severity: rule.Severity,
					Message:  "Avoid using tabs, use spaces instead",
					Line:     i + 1,
				})
			}
		}

	default:
		if rule.Pattern != "" {
			re, err := regexp.Compile(rule.Pattern)
			if err == nil {
				for i, line := range lines {
					loc := re.FindStringIndex(line)
					if loc != nil {
						issues = append(issues, LintIssue{
							RuleID:    rule.ID,
							Severity:  rule.Severity,
							Message:   fmt.Sprintf("Line matches rule: %s", rule.Name),
							Line:      i + 1,
							Column:    loc[0] + 1,
							EndLine:   i + 1,
							EndColumn: loc[1],
						})
					}
				}
			}
		}
	}

	return issues
}
