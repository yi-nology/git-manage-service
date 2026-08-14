package ai

func BuildConflictContext(conflictDiff, oursBranch, theirsBranch string, maxChars int) string {
	b := NewContextBuilder()

	b.AddSection("Ours Branch", oursBranch)
	b.AddSection("Theirs Branch", theirsBranch)

	if len(conflictDiff) > maxChars {
		conflictDiff = ClampText(conflictDiff, maxChars)
	}

	b.AddCodeSection("Conflict Diff", "diff", conflictDiff)
	return b.Build()
}
