package codereview

import (
	"path/filepath"
	"strings"
)

type TestRequiredRule struct{}

func (r *TestRequiredRule) ID() string { return "test-required" }

func (r *TestRequiredRule) Check(ctx *RuleContext) ([]*Finding, error) {
	var findings []*Finding
	srcFiles := map[string]bool{}
	testFiles := map[string]bool{}

	for _, f := range ctx.Files {
		if f.IsDeleted {
			continue
		}
		path := f.NewPath
		if isSourceFile(path) && !isTestFile(path) {
			srcFiles[path] = true
		}
		if isTestFile(path) {
			testFiles[path] = true
		}
	}

	for srcPath := range srcFiles {
		base := fileBaseWithoutExt(srcPath)
		hasTest := false
		for testPath := range testFiles {
			testBase := fileBaseWithoutExt(testPath)
			if strings.Contains(testBase, base) || strings.Contains(testBase, strings.ReplaceAll(base, ".", "_")) {
				hasTest = true
				break
			}
		}
		if !hasTest {
			findings = append(findings, &Finding{
				RuleID:      r.ID(),
				Source:      "rule",
				Severity:    SeverityLow,
				FilePath:    srcPath,
				Title:       "Source file changed without corresponding tests",
				Message:     "No test file found for " + srcPath + ". Consider adding tests for new or modified logic.",
				Suggestion:  "Add unit tests to cover the changes in this file.",
				Fingerprint: computeFingerprint(r.ID(), srcPath, 0, "no-test"),
			})
		}
	}

	return findings, nil
}

func isSourceFile(path string) bool {
	ext := filepath.Ext(path)
	switch ext {
	case ".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".java", ".rs", ".rb":
		return true
	}
	return false
}

func isTestFile(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	dir := strings.ToLower(filepath.Dir(path))
	return strings.Contains(base, "_test.") ||
		strings.Contains(base, ".test.") ||
		strings.Contains(base, ".spec.") ||
		strings.Contains(base, "_spec.") ||
		strings.Contains(dir, "test") ||
		strings.Contains(dir, "spec") ||
		strings.Contains(dir, "__tests__")
}

func fileBaseWithoutExt(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	return base[:len(base)-len(ext)]
}
