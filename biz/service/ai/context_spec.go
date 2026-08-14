package ai

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
