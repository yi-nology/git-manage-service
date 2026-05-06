package ai

import "fmt"

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

func BuildMergePlanContext(oursBranch, theirsBranch, baseBranch string, diff string, conflicts int, maxChars int) string {
	b := NewContextBuilder()

	b.AddSection("Ours Branch", oursBranch)
	b.AddSection("Theirs Branch", theirsBranch)
	b.AddSection("Base Branch", baseBranch)
	b.AddSection("Conflicts Detected", fmt.Sprintf("%d", conflicts))

	if len(diff) > maxChars {
		diff = ClampText(diff, maxChars)
	}

	b.AddCodeSection("Merge Diff Preview", "diff", diff)
	return b.Build()
}
