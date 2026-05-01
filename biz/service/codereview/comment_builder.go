package codereview

import (
	"fmt"
	"strings"
)

func BuildSummaryComment(result *AggregatedResult) string {
	var b strings.Builder

	emoji := "✅"
	if result.RiskLevel == SeverityCritical || result.RiskLevel == SeverityHigh {
		emoji = "🚫"
	} else if result.RiskLevel == SeverityMedium {
		emoji = "⚠️"
	}

	b.WriteString(fmt.Sprintf("## %s Code Review Summary\n\n", emoji))

	if result.Blocked {
		b.WriteString(fmt.Sprintf("**Status: BLOCKED** — %s\n\n", result.BlockReason))
	} else {
		b.WriteString(fmt.Sprintf("**Status: Passed** (risk level: %s)\n\n", result.RiskLevel))
	}

	b.WriteString(fmt.Sprintf("- **Files changed:** %d\n", result.FileCount))
	b.WriteString(fmt.Sprintf("- **Lines:** +%d / -%d\n", result.TotalAdd, result.TotalDel))
	b.WriteString(fmt.Sprintf("- **Findings:** %d\n\n", len(result.Findings)))

	critical := countBySeverity(result.Findings, SeverityCritical)
	high := countBySeverity(result.Findings, SeverityHigh)
	medium := countBySeverity(result.Findings, SeverityMedium)
	low := countBySeverity(result.Findings, SeverityLow)
	info := countBySeverity(result.Findings, SeverityInfo)

	if len(result.Findings) > 0 {
		b.WriteString("| Severity | Count |\n|----------|-------|\n")
		if critical > 0 {
			b.WriteString(fmt.Sprintf("| 🔴 Critical | %d |\n", critical))
		}
		if high > 0 {
			b.WriteString(fmt.Sprintf("| 🟠 High | %d |\n", high))
		}
		if medium > 0 {
			b.WriteString(fmt.Sprintf("| 🟡 Medium | %d |\n", medium))
		}
		if low > 0 {
			b.WriteString(fmt.Sprintf("| 🔵 Low | %d |\n", low))
		}
		if info > 0 {
			b.WriteString(fmt.Sprintf("| ⚪ Info | %d |\n", info))
		}
		b.WriteString("\n")
	}

	grouped := groupFindingsBySeverity(result.Findings)
	order := []Severity{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo}
	for _, sev := range order {
		findings, ok := grouped[sev]
		if !ok || len(findings) == 0 {
			continue
		}
		b.WriteString(fmt.Sprintf("### %s\n\n", strings.Title(string(sev))))
		for _, f := range findings {
			b.WriteString(formatFinding(f))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n---\n*Powered by git-manage-service code review*\n")
	return b.String()
}

func BuildInlineComment(f *Finding) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("**[%s] %s**\n\n", strings.ToUpper(string(f.Severity)), f.Title))
	b.WriteString(f.Message)
	b.WriteString("\n")
	if f.Suggestion != "" {
		b.WriteString(fmt.Sprintf("\n💡 **Suggestion:** %s\n", f.Suggestion))
	}
	b.WriteString(fmt.Sprintf("\n*Rule: `%s`*\n", f.RuleID))
	return b.String()
}

func formatFinding(f *Finding) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("- **%s** ", f.Title))
	if f.FilePath != "" {
		b.WriteString(fmt.Sprintf("in `%s`", f.FilePath))
	}
	if f.NewLine > 0 {
		b.WriteString(fmt.Sprintf(":%d", f.NewLine))
	}
	b.WriteString("\n")
	if f.Message != "" {
		b.WriteString(fmt.Sprintf("  > %s\n", f.Message))
	}
	if f.Suggestion != "" {
		b.WriteString(fmt.Sprintf("  > 💡 %s\n", f.Suggestion))
	}
	b.WriteString("\n")
	return b.String()
}

func groupFindingsBySeverity(findings []*Finding) map[Severity][]*Finding {
	m := make(map[Severity][]*Finding)
	for _, f := range findings {
		m[f.Severity] = append(m[f.Severity], f)
	}
	return m
}
