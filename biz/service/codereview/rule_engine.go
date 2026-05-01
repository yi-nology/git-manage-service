package codereview

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

type Finding struct {
	RuleID      string   `json:"rule_id"`
	Source      string   `json:"source"`
	Severity    Severity `json:"severity"`
	FilePath    string   `json:"file_path"`
	OldLine     int      `json:"old_line"`
	NewLine     int      `json:"new_line"`
	Title       string   `json:"title"`
	Message     string   `json:"message"`
	Suggestion  string   `json:"suggestion"`
	Fingerprint string   `json:"fingerprint"`
}

type Rule interface {
	ID() string
	Check(ctx *RuleContext) ([]*Finding, error)
}

type RuleContext struct {
	Files    []*FileDiff
	Provider string
	RepoKey  string
	Owner    string
	Repo     string
	MRIID    string
}

func computeFingerprint(ruleID, filePath string, line int, content string) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s:%s:%d:%s", ruleID, filePath, line, strings.TrimSpace(content))))
	return fmt.Sprintf("%x", h.Sum(nil))[:16]
}

var registeredRules []Rule

func RegisterRule(r Rule) {
	registeredRules = append(registeredRules, r)
}

func GetRules() []Rule {
	if len(registeredRules) == 0 {
		initDefaultRules()
	}
	return registeredRules
}

func initDefaultRules() {
	RegisterRule(&SecretRule{})
	RegisterRule(&ProtectedFileRule{})
	RegisterRule(&DiffSizeRule{})
	RegisterRule(&MigrationRule{})
	RegisterRule(&TestRequiredRule{})
}
