package codereview

import "fmt"

type ProcessStep struct {
	Name   string
	Status string
	Detail string
}

type AggregatedResult struct {
	Findings    []*Finding
	TotalAdd    int
	TotalDel    int
	FileCount   int
	RiskLevel   Severity
	Blocked     bool
	BlockReason string
	ProcessLog  []*ProcessStep
}

func Aggregate(findings []*Finding, totalAdd, totalDel, fileCount int, blockOnHigh bool, processLog []*ProcessStep) *AggregatedResult {
	findings = deduplicate(findings)

	risk := calculateRisk(findings)
	blocked := false
	reason := ""

	if blockOnHigh && (risk == SeverityCritical || risk == SeverityHigh) {
		blocked = true
		count := countBySeverity(findings, SeverityCritical) + countBySeverity(findings, SeverityHigh)
		reason = fmt.Sprintf("Code review blocked: %d critical/high severity findings", count)
	}

	return &AggregatedResult{
		Findings:    findings,
		TotalAdd:    totalAdd,
		TotalDel:    totalDel,
		FileCount:   fileCount,
		RiskLevel:   risk,
		Blocked:     blocked,
		BlockReason: reason,
		ProcessLog:  processLog,
	}
}

func deduplicate(findings []*Finding) []*Finding {
	seen := make(map[string]bool, len(findings))
	result := make([]*Finding, 0, len(findings))
	for _, f := range findings {
		if seen[f.Fingerprint] {
			continue
		}
		seen[f.Fingerprint] = true
		result = append(result, f)
	}
	return result
}

func calculateRisk(findings []*Finding) Severity {
	if len(findings) == 0 {
		return SeverityInfo
	}

	counts := map[Severity]int{}
	for _, f := range findings {
		counts[f.Severity]++
	}

	if counts[SeverityCritical] > 0 {
		return SeverityCritical
	}
	if counts[SeverityHigh] > 0 {
		return SeverityHigh
	}
	if counts[SeverityMedium] >= 3 {
		return SeverityHigh
	}
	if counts[SeverityMedium] > 0 {
		return SeverityMedium
	}
	if counts[SeverityLow] >= 5 {
		return SeverityMedium
	}
	if counts[SeverityLow] > 0 {
		return SeverityLow
	}
	return SeverityInfo
}

func countBySeverity(findings []*Finding, sev Severity) int {
	n := 0
	for _, f := range findings {
		if f.Severity == sev {
			n++
		}
	}
	return n
}
