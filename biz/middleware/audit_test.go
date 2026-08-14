package middleware

import (
	"strings"
	"testing"
)

// TestNamedParamPatternsRoundTrip verifies every named-param pattern produces
// a path that matches an actual auditRoutes key — guarding against the bug
// where multi-segment repl values were applied as first-segment replacements,
// silently dropping audit entries.
func TestNamedParamPatternsRoundTrip(t *testing.T) {
	// Concrete request paths exercising each pattern (segment counts chosen
	// to catch doubled-suffix and multi-segment mismatches).
	cases := []string{
		"/api/v1/repo/my-repo/author/fix",
		"/api/v1/repo/my-repo/author/fix-all",
		"/api/v1/repo/my-repo/author/config",
		"/api/v1/repo/my-repo/maintenance/slim",
		"/api/v1/repo/my-repo/maintenance/gc",
		"/api/v1/reviews/config/my-repo",
		"/api/v1/review/remote-config/3/acme/widgets",
		"/api/v1/branch-rules/remote-config/3/acme/widgets",
		"/api/webhooks/trigger/some-token-value",
		"/api/v1/spec/commit/deep/nested/path.spec",
		"/api/v1/spec/content/deep/nested/path.spec",
		"/api/v1/spec/rules/7",
		"/api/v1/settings/llm-providers/5/default",
		"/api/v1/settings/llm-providers/5/test",
		"/api/v1/settings/review-rules/9",
	}

	for _, path := range cases {
		matched := false
		for _, p := range namedParamPatterns {
			if !strings.HasPrefix(path, p.prefix) {
				continue
			}
			if p.suffix != "" && !strings.Contains(path, p.suffix) {
				continue
			}
			norm := applyNamedPattern(path, p.prefix, p.repl, p.whole)
			if _, ok := auditRoutes["POST:"+norm]; ok {
				matched = true
				break
			}
			if _, ok := auditRoutes["PUT:"+norm]; ok {
				matched = true
				break
			}
			if _, ok := auditRoutes["DELETE:"+norm]; ok {
				matched = true
				break
			}
		}
		if !matched {
			t.Errorf("path %q normalized by patterns but matches NO auditRoutes key — audit entry silently dropped", path)
		}
	}
}
