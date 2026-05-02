package commit_analyzer

import (
	"fmt"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func (s *AnalyzerService) GenerateSyncRecommendations(repoKey, taskKey string) (*po.SyncRecommendation, error) {
	stats, err := s.GetCommitStats(repoKey, 30)
	if err != nil {
		return nil, err
	}

	commitCount := stats["commitCount"].(int64)
	commitTypeCount := stats["commitTypeCount"].(map[string]int)

	averageCommitsPerDay := float64(commitCount) / 30.0

	syncFrequency := "daily"
	if averageCommitsPerDay < 0.5 {
		syncFrequency = "weekly"
	} else if averageCommitsPerDay > 5 {
		syncFrequency = "hourly"
	}

	recommendation := fmt.Sprintf("基于过去30天的分析，该仓库平均每天有 %.2f 次提交。", averageCommitsPerDay)
	recommendation += fmt.Sprintf("建议的同步频率为 %s。", syncFrequency)

	if featCount, exists := commitTypeCount["feat"]; exists && featCount > 0 {
		recommendation += fmt.Sprintf(" 包含 %d 个新功能提交。", featCount)
	}
	if fixCount, exists := commitTypeCount["fix"]; exists && fixCount > 0 {
		recommendation += fmt.Sprintf(" 包含 %d 个 bug 修复提交。", fixCount)
	}

	confidence := 0.7
	if commitCount > 20 {
		confidence = 0.9
	} else if commitCount > 5 {
		confidence = 0.8
	}

	syncRecommendation := &po.SyncRecommendation{
		RepoKey:        repoKey,
		TaskKey:        taskKey,
		Recommendation: recommendation,
		SyncFrequency:  syncFrequency,
		Confidence:     confidence,
		LastAnalysis:   time.Now(),
		IsApplied:      false,
	}

	existing, err := s.commitAnalysisDAO.GetSyncRecommendation(repoKey, taskKey)
	if err != nil {
		if err := s.commitAnalysisDAO.CreateSyncRecommendation(syncRecommendation); err != nil {
			_ = err
		}
	} else {
		existing.Recommendation = syncRecommendation.Recommendation
		existing.SyncFrequency = syncRecommendation.SyncFrequency
		existing.Confidence = syncRecommendation.Confidence
		existing.LastAnalysis = syncRecommendation.LastAnalysis
		if err := s.commitAnalysisDAO.UpdateSyncRecommendation(existing); err != nil {
			_ = err
		}
		syncRecommendation = existing
	}

	return syncRecommendation, nil
}

func (s *AnalyzerService) GetSyncRecommendations(repoKey string, limit int) ([]po.SyncRecommendation, error) {
	return s.commitAnalysisDAO.GetSyncRecommendationsByRepo(repoKey, limit)
}

func (s *AnalyzerService) UpdateSyncRecommendation(recommendation *po.SyncRecommendation) error {
	return s.commitAnalysisDAO.UpdateSyncRecommendation(recommendation)
}
