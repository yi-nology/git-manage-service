package ai

func BuildSyncFailureContext(logs, stderr string, currentBranch, trackingBranch string, recentActions []string, maxChars int) string {
	b := NewContextBuilder()

	b.AddSection("Current Branch", currentBranch)
	b.AddSection("Tracking Branch", trackingBranch)

	if len(recentActions) > 0 {
		b.AddListSection("Recent Actions", recentActions)
	}

	if len(stderr) > 0 {
		if len(stderr) > maxChars/2 {
			stderr = ClampText(stderr, maxChars/2)
		}
		b.AddCodeSection("Git Error Output", "text", stderr)
	}

	if len(logs) > 0 {
		if len(logs) > maxChars/2 {
			logs = ClampText(logs, maxChars/2)
		}
		b.AddCodeSection("Sync Logs", "text", logs)
	}

	return b.Build()
}
