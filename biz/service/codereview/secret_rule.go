package codereview

import (
	"regexp"
	"strings"
)

type SecretRule struct{}

func (r *SecretRule) ID() string { return "secret-detection" }

var secretPatterns = []struct {
	pattern  *regexp.Regexp
	name     string
	severity Severity
}{
	{regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*['"][^'"]{4,}['"]`), "Hardcoded password", SeverityCritical},
	{regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*['"][^'"]{8,}['"]`), "Hardcoded API key", SeverityCritical},
	{regexp.MustCompile(`(?i)(secret|token)\s*[:=]\s*['"][^'"]{8,}['"]`), "Hardcoded secret/token", SeverityCritical},
	{regexp.MustCompile(`(?i)(AKIA|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[0-9A-Z]{16}`), "AWS Access Key", SeverityCritical},
	{regexp.MustCompile(`-----BEGIN (RSA |EC |DSA )?PRIVATE KEY-----`), "Private key embedded", SeverityCritical},
	{regexp.MustCompile(`(?i)jdbc:[a-z]+://[^\s'"]+:(?:[^\s'"]+)@`), "JDBC connection string with password", SeverityHigh},
	{regexp.MustCompile(`(?i)mongodb(\+srv)?://[^\s'"]+:[^\s'"]+@`), "MongoDB connection string with password", SeverityHigh},
	{regexp.MustCompile(`(?i)mysql://[^\s'"]+:[^\s'"]+@`), "MySQL connection string with password", SeverityHigh},
	{regexp.MustCompile(`(?i)(postgres|postgresql)://[^\s'"]+:[^\s'"]+@`), "PostgreSQL connection string with password", SeverityHigh},
	{regexp.MustCompile(`ghp_[0-9a-zA-Z]{36}`), "GitHub personal access token", SeverityCritical},
	{regexp.MustCompile(`glpat-[0-9a-zA-Z\-]{20,}`), "GitLab personal access token", SeverityCritical},
}

var falsePositivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(example|placeholder|your[_-]?(api[_-]?key|secret|password|token)|xxx+|changeme|todo)`),
	regexp.MustCompile(`(?i)(test|spec|mock|stub|fixture|dummy)`),
}

func (r *SecretRule) Check(ctx *RuleContext) ([]*Finding, error) {
	var findings []*Finding
	for _, f := range ctx.Files {
		if f.IsDeleted {
			continue
		}
		for _, hunk := range f.Hunks {
			for _, line := range hunk.Lines {
				if line.Type != "add" {
					continue
				}
				content := line.Content
				if isFalsePositive(content) {
					continue
				}
				for _, sp := range secretPatterns {
					if sp.pattern.MatchString(content) {
						findings = append(findings, &Finding{
							RuleID:      r.ID(),
							Source:      "rule",
							Severity:    sp.severity,
							FilePath:    line.FilePath,
							NewLine:     line.NewLine,
							Title:       sp.name,
							Message:     "Potential secret or credential detected in code. Use environment variables or a secret manager instead.",
							Suggestion:  "Move this credential to an environment variable or secret manager (e.g. Vault, AWS Secrets Manager).",
							Fingerprint: computeFingerprint(r.ID(), line.FilePath, line.NewLine, content),
						})
						break
					}
				}
			}
		}
	}
	return findings, nil
}

func isFalsePositive(content string) bool {
	lower := strings.ToLower(content)
	for _, fp := range falsePositivePatterns {
		if fp.MatchString(lower) {
			return true
		}
	}
	return false
}
