package codereview

import (
	"fmt"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func ApplyPolicy(result *AggregatedResult, task *po.ReviewTask) error {
	checkDAO := db.NewMergeCheckResultDAO()

	check := &po.MergeCheckResult{
		RepoID:    task.RepoID,
		MRIID:     task.MRIID,
		CommitSHA: task.CommitSHA,
		CheckType: "code_review",
	}

	if result.Blocked {
		check.Status = "failed"
		check.RiskLevel = string(result.RiskLevel)
		check.Message = result.BlockReason
	} else {
		check.Status = "success"
		check.RiskLevel = string(result.RiskLevel)
		critical := countBySeverity(result.Findings, SeverityCritical)
		high := countBySeverity(result.Findings, SeverityHigh)
		medium := countBySeverity(result.Findings, SeverityMedium)
		low := countBySeverity(result.Findings, SeverityLow)
		check.Message = fmt.Sprintf("Review passed. %d findings: %d critical, %d high, %d medium, %d low",
			len(result.Findings), critical, high, medium, low)
	}

	return checkDAO.Create(check)
}
