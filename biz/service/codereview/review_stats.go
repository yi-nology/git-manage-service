package codereview

import (
	"cmp"
	"slices"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/pkg/timefmt"
)

func GetReviewStats(repoID uint, period string) map[string]interface{} {
	since := parsePeriod(period)

	tasks, _ := db.NewReviewTaskDAO().FindByTimeRange(repoID, since, time.Now())
	findings, _ := db.NewReviewFindingDAO().FindByTimeRange(repoID, since, time.Now())

	totalReviews := len(tasks)
	if totalReviews == 0 {
		return map[string]interface{}{
			"total_reviews":  0,
			"total_findings": 0,
			"pass_rate":      0,
			"by_risk_level":  map[string]int{},
			"by_severity":    map[string]int{},
			"by_source":      map[string]int{},
			"by_rule":        map[string]int{},
			"daily_trend":    []map[string]interface{}{},
			"top_rules":      []map[string]interface{}{},
			"top_files":      []map[string]interface{}{},
		}
	}

	passCount := 0
	blockedCount := 0
	failedCount := 0
	riskCounts := map[string]int{}
	severityCounts := map[string]int{}
	sourceCounts := map[string]int{}
	ruleCounts := map[string]int{}
	fileCounts := map[string]int{}
	dailyData := map[string]*dayBucket{}

	for _, t := range tasks {
		switch t.Status {
		case "success":
			passCount++
		case "blocked":
			blockedCount++
		case "failed":
			failedCount++
		}
		riskCounts[t.RiskLevel]++

		day := t.CreatedAt.Format(timefmt.LayoutDate)
		bucket := dailyData[day]
		if bucket == nil {
			bucket = &dayBucket{Date: day}
			dailyData[day] = bucket
		}
		bucket.Total++
		if t.Status == "success" {
			bucket.Passed++
		}
	}

	for _, f := range findings {
		severityCounts[f.Severity]++
		sourceCounts[f.Source]++
		ruleCounts[f.RuleID]++
		if f.FilePath != "" {
			fileCounts[f.FilePath]++
		}
	}

	passRate := float64(passCount) / float64(totalReviews) * 100

	dailyTrend := sortedDailyTrend(dailyData, since)

	return map[string]interface{}{
		"total_reviews":  totalReviews,
		"total_findings": len(findings),
		"pass_count":     passCount,
		"blocked_count":  blockedCount,
		"failed_count":   failedCount,
		"pass_rate":      passRate,
		"by_risk_level":  riskCounts,
		"by_severity":    severityCounts,
		"by_source":      sourceCounts,
		"by_rule":        topN(ruleCounts, 10),
		"top_files":      topN(fileCounts, 10),
		"daily_trend":    dailyTrend,
	}
}

type dayBucket struct {
	Date   string
	Total  int
	Passed int
}

func parsePeriod(period string) time.Time {
	now := time.Now()
	switch period {
	case "1d":
		return now.AddDate(0, 0, -1)
	case "7d":
		return now.AddDate(0, 0, -7)
	case "30d":
		return now.AddDate(0, 0, -30)
	case "90d":
		return now.AddDate(0, 0, -90)
	default:
		return now.AddDate(0, 0, -7)
	}
}

func sortedDailyTrend(data map[string]*dayBucket, since time.Time) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	for d := since; !d.After(time.Now()); d = d.AddDate(0, 0, 1) {
		key := d.Format(timefmt.LayoutDate)
		bucket := data[key]
		if bucket == nil {
			result = append(result, map[string]interface{}{
				"date":  key,
				"total": 0, "passed": 0, "rate": 0,
			})
		} else {
			rate := 0.0
			if bucket.Total > 0 {
				rate = float64(bucket.Passed) / float64(bucket.Total) * 100
			}
			result = append(result, map[string]interface{}{
				"date":  key,
				"total": bucket.Total, "passed": bucket.Passed, "rate": rate,
			})
		}
	}
	return result
}

func topN(m map[string]int, n int) []map[string]interface{} {
	type kv struct {
		Key   string
		Value int
	}
	var sorted []kv
	for k, v := range m {
		sorted = append(sorted, kv{k, v})
	}
	slices.SortFunc(sorted, func(a, b kv) int { return cmp.Compare(b.Value, a.Value) })
	if len(sorted) > n {
		sorted = sorted[:n]
	}
	result := make([]map[string]interface{}, 0, len(sorted))
	for _, item := range sorted {
		result = append(result, map[string]interface{}{
			"name":  item.Key,
			"count": item.Value,
		})
	}
	return result
}
