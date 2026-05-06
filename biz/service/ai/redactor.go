package ai

import (
	"regexp"
	"strings"
)

var assignmentSecretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(api[_-]?key\s*[:=]\s*)[^\s"'` + "`" + `]+`),
	regexp.MustCompile(`(?i)(secret\s*[:=]\s*)[^\s"'` + "`" + `]+`),
	regexp.MustCompile(`(?i)(password\s*[:=]\s*)[^\s"'` + "`" + `]+`),
	regexp.MustCompile(`(?i)(token\s*[:=]\s*)[^\s"'` + "`" + `]+`),
	regexp.MustCompile(`(?i)(authorization:\s*bearer\s+)[^\s"'` + "`" + `]+`),
}

var standaloneSecretPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)AKIA[0-9A-Z]{16}`),
	regexp.MustCompile(`(?i)sk-[a-z0-9]{20,}`),
}

func RedactSecrets(text string) string {
	redacted := text
	for _, pattern := range assignmentSecretPatterns {
		redacted = pattern.ReplaceAllString(redacted, "${1}[REDACTED]")
	}
	for _, pattern := range standaloneSecretPatterns {
		redacted = pattern.ReplaceAllString(redacted, "[REDACTED]")
	}
	return redacted
}

func ClampText(text string, maxChars int) string {
	if maxChars <= 0 || len(text) <= maxChars {
		return text
	}
	head := maxChars / 2
	tail := maxChars - head
	var b strings.Builder
	b.Grow(maxChars + 64)
	b.WriteString(text[:head])
	b.WriteString("\n\n... [truncated by AI input budget] ...\n\n")
	b.WriteString(text[len(text)-tail:])
	return b.String()
}
