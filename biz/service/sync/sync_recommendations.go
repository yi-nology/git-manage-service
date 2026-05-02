package sync

import (
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func (s *SyncService) GetSyncRecommendations(repoKey, taskKey string) (*po.SyncRecommendation, error) {
	return s.commitAnalyzer.GenerateSyncRecommendations(repoKey, taskKey)
}

func (s *SyncService) GetSyncRecommendationsByRepo(repoKey string, limit int) ([]po.SyncRecommendation, error) {
	return s.commitAnalyzer.GetSyncRecommendations(repoKey, limit)
}

func (s *SyncService) ApplySyncRecommendation(repoKey, taskKey string) error {
	recommendation, err := s.commitAnalyzer.GenerateSyncRecommendations(repoKey, taskKey)
	if err != nil {
		return err
	}

	task, err := s.syncTaskDAO.FindByKey(taskKey)
	if err != nil {
		return err
	}

	cronExpr := task.Cron
	switch recommendation.SyncFrequency {
	case "hourly":
		cronExpr = "0 * * * *"
	case "daily":
		cronExpr = "0 0 * * *"
	case "weekly":
		cronExpr = "0 0 * * 0"
	}

	task.Cron = cronExpr
	if err := s.syncTaskDAO.Save(task); err != nil {
		return err
	}

	recommendation.IsApplied = true
	if err := s.commitAnalyzer.UpdateSyncRecommendation(recommendation); err != nil {
		_ = err
	}

	return nil
}

func (s *SyncService) AnalyzeRepoForSync(repoPath, repoKey string) error {
	return s.commitAnalyzer.AnalyzeRepo(repoPath, repoKey)
}
