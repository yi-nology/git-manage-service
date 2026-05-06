package ai

import (
	"fmt"
	"strings"
)

type DiffSummary struct {
	Patch        string
	RiskLevel    string
	Summary      string
	AddedLines   int
	RemovedLines int
}

func BuildLineDiff(original, modified string) DiffSummary {
	orig := strings.Split(original, "\n")
	next := strings.Split(modified, "\n")
	maxLen := len(orig)
	if len(next) > maxLen {
		maxLen = len(next)
	}

	var lines []string
	added := 0
	removed := 0
	for i := 0; i < maxLen; i++ {
		var oldLine, newLine string
		if i < len(orig) {
			oldLine = orig[i]
		}
		if i < len(next) {
			newLine = next[i]
		}
		if oldLine == newLine {
			continue
		}
		if oldLine != "" {
			lines = append(lines, fmt.Sprintf("-%d %s", i+1, oldLine))
			removed++
		}
		if newLine != "" {
			lines = append(lines, fmt.Sprintf("+%d %s", i+1, newLine))
			added++
		}
	}

	risk := "low"
	total := added + removed
	switch {
	case total > 80:
		risk = "high"
	case total > 25:
		risk = "medium"
	}

	return DiffSummary{
		Patch:        strings.Join(lines, "\n"),
		RiskLevel:    risk,
		Summary:      fmt.Sprintf("AI proposed %d added and %d removed lines", added, removed),
		AddedLines:   added,
		RemovedLines: removed,
	}
}
