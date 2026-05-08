package codereview

import (
	"fmt"
)

type DiffSizeRule struct{}

func (r *DiffSizeRule) ID() string { return "diff-size" }

var (
	defaultMaxFileLines  = 500
	defaultMaxTotalLines = 3000
	defaultMaxFilesPerMR = 50
)

func getMaxFileLines() int {
	cfg := getConfig()
	if cfg.MaxDiffLines > 0 {
		return cfg.MaxDiffLines / 6
	}
	return defaultMaxFileLines
}

func getMaxTotalLines() int {
	cfg := getConfig()
	if cfg.MaxDiffLines > 0 {
		return cfg.MaxDiffLines
	}
	return defaultMaxTotalLines
}

func getMaxFilesPerMR() int {
	cfg := getConfig()
	if cfg.MaxFiles > 0 {
		return cfg.MaxFiles
	}
	return defaultMaxFilesPerMR
}

func (r *DiffSizeRule) Check(ctx *RuleContext) ([]*Finding, error) {
	var findings []*Finding
	totalAdd, totalDel := 0, 0
	maxFileLines := getMaxFileLines()
	maxTotalLines := getMaxTotalLines()
	maxFilesPerMR := getMaxFilesPerMR()

	if len(ctx.Files) > maxFilesPerMR {
		findings = append(findings, &Finding{
			RuleID:      r.ID(),
			Source:      "rule",
			Severity:    SeverityMedium,
			Title:       "Too many files changed",
			Message:     fmt.Sprintf("MR contains %d files (max recommended: %d). Consider splitting into smaller MRs.", len(ctx.Files), maxFilesPerMR),
			Suggestion:  "Break this MR into smaller, focused changes for better review quality.",
			Fingerprint: computeFingerprint(r.ID(), "", 0, fmt.Sprintf("filecount:%d", len(ctx.Files))),
		})
	}

	for _, f := range ctx.Files {
		changes := f.Additions + f.Deletions
		totalAdd += f.Additions
		totalDel += f.Deletions

		if changes > maxFileLines {
			findings = append(findings, &Finding{
				RuleID:      r.ID(),
				Source:      "rule",
				Severity:    SeverityMedium,
				FilePath:    f.NewPath,
				Title:       "File diff too large",
				Message:     fmt.Sprintf("File %s has %d changed lines (max recommended: %d).", f.NewPath, changes, maxFileLines),
				Suggestion:  "Consider splitting large file changes into separate MRs.",
				Fingerprint: computeFingerprint(r.ID(), f.NewPath, 0, fmt.Sprintf("lines:%d", changes)),
			})
		}
	}

	total := totalAdd + totalDel
	if total > maxTotalLines {
		findings = append(findings, &Finding{
			RuleID:      r.ID(),
			Source:      "rule",
			Severity:    SeverityHigh,
			Title:       "MR diff too large for effective review",
			Message:     fmt.Sprintf("Total diff is +%d/-%d (%d lines). Reviews lose effectiveness above %d lines.", totalAdd, totalDel, total, maxTotalLines),
			Suggestion:  "Split this MR into multiple smaller MRs for thorough review.",
			Fingerprint: computeFingerprint(r.ID(), "", 0, fmt.Sprintf("total:%d", total)),
		})
	}

	return findings, nil
}
