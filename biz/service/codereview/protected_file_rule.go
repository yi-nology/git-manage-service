package codereview

import (
	"path/filepath"
	"strings"
)

type ProtectedFileRule struct{}

func (r *ProtectedFileRule) ID() string { return "protected-file" }

var protectedPaths = []struct {
	pattern  string
	severity Severity
	message  string
}{
	{".env", SeverityCritical, "Environment configuration file modified"},
	{".env.production", SeverityCritical, "Production environment file modified"},
	{".env.staging", SeverityHigh, "Staging environment file modified"},
	{".github/workflows/", SeverityHigh, "CI/CD workflow modified"},
	{".gitlab-ci.yml", SeverityHigh, "GitLab CI configuration modified"},
	{"Dockerfile", SeverityMedium, "Dockerfile modified"},
	{"docker-compose.yml", SeverityMedium, "Docker Compose configuration modified"},
	{"docker-compose.yaml", SeverityMedium, "Docker Compose configuration modified"},
	{"Makefile", SeverityLow, "Makefile modified"},
	{"go.mod", SeverityMedium, "Go module definition modified"},
	{"go.sum", SeverityLow, "Go module checksums modified"},
	{"package.json", SeverityMedium, "Node.js package definition modified"},
	{"Cargo.toml", SeverityMedium, "Rust package definition modified"},
	{".pem", SeverityCritical, "Certificate/PEM file modified"},
	{".key", SeverityCritical, "Key file modified"},
	{"id_rsa", SeverityCritical, "SSH private key modified"},
	{"id_ed25519", SeverityCritical, "SSH private key modified"},
	{"known_hosts", SeverityMedium, "SSH known_hosts modified"},
}

func (r *ProtectedFileRule) Check(ctx *RuleContext) ([]*Finding, error) {
	var findings []*Finding
	for _, f := range ctx.Files {
		if f.IsDeleted {
			continue
		}
		path := f.NewPath
		for _, pf := range protectedPaths {
			matched := false
			if strings.Contains(pf.pattern, "/") {
				matched = strings.Contains(path, pf.pattern)
			} else {
				matched = filepath.Base(path) == pf.pattern || strings.HasSuffix(path, "."+pf.pattern) || strings.HasSuffix(path, "/"+pf.pattern)
			}
			if !matched && strings.HasPrefix(pf.pattern, ".") {
				matched = strings.Contains(path, pf.pattern)
			}
			if matched {
				findings = append(findings, &Finding{
					RuleID:      r.ID(),
					Source:      "rule",
					Severity:    pf.severity,
					FilePath:    path,
					Title:       pf.message,
					Message:     pf.message + ". Ensure this change is intentional and reviewed carefully.",
					Suggestion:  "Consider if this change is necessary. Protected files often affect deployment or security.",
					Fingerprint: computeFingerprint(r.ID(), path, 0, pf.pattern),
				})
				break
			}
		}
	}
	return findings, nil
}
