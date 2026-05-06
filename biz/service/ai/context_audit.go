package ai

import (
	"fmt"
	"strings"
)

func BuildAuditSummaryContext(events []string, stats map[string]int, anomalies []string, maxChars int) string {
	b := NewContextBuilder()

	b.AddSection("Event Statistics", formatMap(stats))

	if len(anomalies) > 0 {
		b.AddListSection("Anomalies Detected", anomalies)
	}

	if len(events) > 0 {
		eventsText := strings.Join(events, "\n")
		if len(eventsText) > maxChars {
			eventsText = ClampText(eventsText, maxChars)
		}
		b.AddCodeSection("Recent Audit Events", "text", eventsText)
	}

	return b.Build()
}

func BuildStatsInsightContext(stats map[string]interface{}, trends map[string][]int, authorActivity map[string]int, maxChars int) string {
	b := NewContextBuilder()

	b.AddSection("Statistics Overview", formatInterfaceMap(stats))

	if len(trends) > 0 {
		b.AddSection("Trends", formatTrends(trends))
	}

	if len(authorActivity) > 0 {
		b.AddSection("Author Activity", formatIntMap(authorActivity))
	}

	return b.Build()
}

func BuildWebhookFailureContext(payload, response string, statusCode int, eventType string, maxChars int) string {
	b := NewContextBuilder()

	b.AddSection("Event Type", eventType)
	b.AddSection("HTTP Status", fmt.Sprintf("%d", statusCode))

	if len(payload) > 0 {
		if len(payload) > maxChars/2 {
			payload = ClampText(payload, maxChars/2)
		}
		b.AddCodeSection("Request Payload", "json", payload)
	}

	if len(response) > 0 {
		if len(response) > maxChars/2 {
			response = ClampText(response, maxChars/2)
		}
		b.AddCodeSection("Response Body", "text", response)
	}

	return b.Build()
}

func formatMap(m map[string]int) string {
	var b strings.Builder
	for k, v := range m {
		fmt.Fprintf(&b, "- %s: %d\n", k, v)
	}
	return b.String()
}

func formatIntMap(m map[string]int) string {
	return formatMap(m)
}

func formatInterfaceMap(m map[string]interface{}) string {
	var b strings.Builder
	for k, v := range m {
		fmt.Fprintf(&b, "- %s: %v\n", k, v)
	}
	return b.String()
}

func formatTrends(trends map[string][]int) string {
	var b strings.Builder
	for k, v := range trends {
		fmt.Fprintf(&b, "- %s: %v\n", k, v)
	}
	return b.String()
}
