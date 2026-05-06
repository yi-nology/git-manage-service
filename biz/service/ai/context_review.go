package ai

func BuildReviewContext(diffContent string, changedFiles []string, existingFindings []string, repoLanguage string, maxChars int) string {
	b := NewContextBuilder()

	b.AddListSection("Changed Files", changedFiles)

	if len(existingFindings) > 0 {
		b.AddListSection("Existing Lint Findings (skip reporting these)", existingFindings)
	}

	b.AddSection("Repository Language", repoLanguage)

	if len(diffContent) > maxChars {
		diffContent = ClampText(diffContent, maxChars)
	}

	b.AddCodeSection("Full Diff", "diff", diffContent)
	return b.Build()
}

func BuildReviewReplyContext(reviewSummary string, reviewerComments []string, authorResponseDraft string, maxChars int) string {
	b := NewContextBuilder()

	b.AddSection("Review Summary", reviewSummary)

	if len(reviewerComments) > 0 {
		b.AddListSection("Reviewer Comments", reviewerComments)
	}

	if authorResponseDraft != "" {
		b.AddSection("Author Response Draft", authorResponseDraft)
	}

	return b.Build()
}
