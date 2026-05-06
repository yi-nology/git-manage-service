package ai

import (
	"fmt"
)

func BuildSpecContext(specContent string, sectionPath string, lintErrors []string, maxChars int) string {
	b := NewContextBuilder()

	if sectionPath != "" {
		b.AddSection("Target Section", sectionPath)
	}

	if len(lintErrors) > 0 {
		b.AddListSection("Lint Errors", lintErrors)
	}

	if len(specContent) > maxChars {
		specContent = ClampText(specContent, maxChars)
	}

	b.AddCodeSection("Specification File", "yaml", specContent)
	return b.Build()
}

func BuildSpecChatContext(specContent, userQuestion, historySummary string, maxChars int) string {
	b := NewContextBuilder()

	if historySummary != "" {
		b.AddSection("Conversation History", historySummary)
	}

	if len(specContent) > maxChars {
		specContent = ClampText(specContent, maxChars)
	}

	b.AddCodeSection("Specification File", "yaml", specContent)
	b.AddSection("User Question", userQuestion)

	return b.Build()
}

func BuildSpecFixContext(specContent string, lineStart, lineEnd int, issueDescription string, maxChars int) string {
	b := NewContextBuilder()

	b.AddSection("Issue to Fix", fmt.Sprintf("Lines %d-%d: %s", lineStart, lineEnd, issueDescription))

	if len(specContent) > maxChars {
		specContent = ClampText(specContent, maxChars)
	}

	b.AddCodeSection("Specification File", "yaml", specContent)
	return b.Build()
}
