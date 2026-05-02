package commit_analyzer

import (
	"encoding/json"
	"fmt"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

func (s *AnalyzerService) updateCommitPatterns(repoKey string, analysis *po.CommitAnalysis) {
	dayOfWeek := analysis.CommitTime.Weekday().String()
	s.updatePattern(repoKey, "daily", dayOfWeek, analysis)

	weekNumber := (analysis.CommitTime.YearDay()-1)/7 + 1
	weekKey := fmt.Sprintf("Week %d", weekNumber)
	s.updatePattern(repoKey, "weekly", weekKey, analysis)

	monthKey := analysis.CommitTime.Format("2006-01")
	s.updatePattern(repoKey, "monthly", monthKey, analysis)

	var languageChanges []LanguageChange
	if err := json.Unmarshal([]byte(analysis.LanguageChanges), &languageChanges); err == nil {
		for _, langChange := range languageChanges {
			s.updatePattern(repoKey, "language", langChange.Language, analysis)
		}
	}
}

func (s *AnalyzerService) updatePattern(repoKey, patternType, patternValue string, analysis *po.CommitAnalysis) {
	pattern, err := s.commitAnalysisDAO.GetLatestCommitPattern(repoKey, patternType, patternValue)
	if err != nil {
		pattern = &po.CommitPattern{
			RepoKey:      repoKey,
			PatternType:  patternType,
			PatternValue: patternValue,
			CommitCount:  1,
			ChangeCount:  analysis.TotalChanges,
			StartDate:    analysis.CommitTime,
			EndDate:      analysis.CommitTime,
		}
		if err := s.commitAnalysisDAO.CreateCommitPattern(pattern); err != nil {
			_ = err
		}
	} else {
		pattern.CommitCount++
		pattern.ChangeCount += analysis.TotalChanges
		pattern.EndDate = analysis.CommitTime
		if err := s.commitAnalysisDAO.UpdateCommitPattern(pattern); err != nil {
			_ = err
		}
	}
}
