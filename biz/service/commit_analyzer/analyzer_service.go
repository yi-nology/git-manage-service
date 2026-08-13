package commit_analyzer

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/git"
)

type AnalyzerService struct {
	commitAnalysisDAO *db.CommitAnalysisDAO
	gitService        *git.GitService
}

func NewAnalyzerService() *AnalyzerService {
	return &AnalyzerService{
		commitAnalysisDAO: db.NewCommitAnalysisDAO(),
		gitService:        git.NewGitService(),
	}
}

type FileChange struct {
	File       string `json:"file"`
	Added      int    `json:"added"`
	Deleted    int    `json:"deleted"`
	ChangeType string `json:"changeType"`
}

type LanguageChange struct {
	Language string `json:"language"`
	Added    int    `json:"added"`
	Deleted  int    `json:"deleted"`
}

func (s *AnalyzerService) AnalyzeCommit(repoPath, repoKey, commitHash string) (*po.CommitAnalysis, error) {
	commitInfo, err := s.gitService.GetCommitInfo(repoPath, commitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit info: %v", err)
	}

	diff, err := s.gitService.GetCommitDiffSimple(repoPath, commitHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit diff: %v", err)
	}

	addedLines, deletedLines, fileChanges, languageChanges := s.parseDiff(diff)
	totalChanges := addedLines + deletedLines

	fileChangesJSON, err := json.Marshal(fileChanges)
	if err != nil {
		return nil, err
	}

	languageChangesJSON, err := json.Marshal(languageChanges)
	if err != nil {
		return nil, err
	}

	commitType := s.detectCommitType(commitInfo.Message)

	isMerge := strings.Contains(commitInfo.Message, "Merge") || strings.Contains(commitInfo.Message, "merge")

	analysis := &po.CommitAnalysis{
		RepoKey:         repoKey,
		CommitHash:      commitHash,
		Author:          commitInfo.Author,
		AuthorEmail:     commitInfo.AuthorEmail,
		Committer:       commitInfo.Committer,
		CommitterEmail:  commitInfo.CommitterEmail,
		Message:         commitInfo.Message,
		CommitTime:      commitInfo.CommitTime,
		AnalysisTime:    time.Now(),
		AddedLines:      addedLines,
		DeletedLines:    deletedLines,
		TotalChanges:    totalChanges,
		FileChanges:     string(fileChangesJSON),
		LanguageChanges: string(languageChangesJSON),
		CommitType:      commitType,
		IsMerge:         isMerge,
	}

	if err := s.commitAnalysisDAO.CreateCommitAnalysis(analysis); err != nil {
		return nil, err
	}

	s.updateCommitPatterns(repoKey, analysis)

	return analysis, nil
}

func (s *AnalyzerService) AnalyzeRepo(repoPath, repoKey string) error {
	commits, err := s.gitService.GetRecentCommits(repoPath, 100)
	if err != nil {
		return fmt.Errorf("failed to get recent commits: %v", err)
	}

	for _, commitHash := range commits {
		_, err := s.commitAnalysisDAO.GetCommitAnalysisByHash(repoKey, commitHash)
		if err == nil {
			continue
		}

		_, err = s.AnalyzeCommit(repoPath, repoKey, commitHash)
		if err != nil {
			continue
		}
	}

	return nil
}

func (s *AnalyzerService) GetCommitPatterns(repoKey string, patternType string, limit int) ([]po.CommitPattern, error) {
	return s.commitAnalysisDAO.GetCommitPatternsByRepo(repoKey, patternType, limit)
}

func (s *AnalyzerService) GetCommitStats(repoKey string, days int) (map[string]interface{}, error) {
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -days)

	commitCount, err := s.commitAnalysisDAO.GetCommitCountByTimeRange(repoKey, startTime, endTime)
	if err != nil {
		return nil, err
	}

	totalChanges, err := s.commitAnalysisDAO.GetTotalChangesByTimeRange(repoKey, startTime, endTime)
	if err != nil {
		return nil, err
	}

	analyses, err := s.commitAnalysisDAO.GetCommitAnalysesByTimeRange(repoKey, startTime, endTime)
	if err != nil {
		return nil, err
	}

	commitTypeCount := make(map[string]int)
	for _, analysis := range analyses {
		commitTypeCount[analysis.CommitType]++
	}

	dailyCommits := make(map[string]int)
	for _, analysis := range analyses {
		dateKey := analysis.CommitTime.Format("2006-01-02")
		dailyCommits[dateKey]++
	}

	return map[string]interface{}{
		"commit_count":     commitCount,
		"total_changes":    totalChanges,
		"commit_type_count": commitTypeCount,
		"daily_commits":    dailyCommits,
		"start_time":       startTime,
		"end_time":         endTime,
	}, nil
}

func (s *AnalyzerService) parseDiff(diff string) (int, int, []FileChange, []LanguageChange) {
	addedLines := 0
	deletedLines := 0
	fileChanges := []FileChange{}
	languageChanges := []LanguageChange{}

	lines := strings.Split(diff, "\n")
	currentFile := ""
	currentAdded := 0
	currentDeleted := 0

	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			if currentFile != "" {
				fileChanges = append(fileChanges, FileChange{
					File:       currentFile,
					Added:      currentAdded,
					Deleted:    currentDeleted,
					ChangeType: s.detectChangeType(currentAdded, currentDeleted),
				})
			}
			parts := strings.Split(line, " ")
			if len(parts) >= 3 {
				currentFile = strings.TrimPrefix(parts[2], "a/")
			}
			currentAdded = 0
			currentDeleted = 0
		} else if strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++") {
			addedLines++
			currentAdded++
		} else if strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---") {
			deletedLines++
			currentDeleted++
		}
	}

	if currentFile != "" {
		fileChanges = append(fileChanges, FileChange{
			File:       currentFile,
			Added:      currentAdded,
			Deleted:    currentDeleted,
			ChangeType: s.detectChangeType(currentAdded, currentDeleted),
		})
	}

	languageMap := make(map[string]LanguageChange)
	for _, change := range fileChanges {
		language := s.detectLanguage(change.File)
		if language != "" {
			if langChange, exists := languageMap[language]; exists {
				langChange.Added += change.Added
				langChange.Deleted += change.Deleted
				languageMap[language] = langChange
			} else {
				languageMap[language] = LanguageChange{
					Language: language,
					Added:    change.Added,
					Deleted:  change.Deleted,
				}
			}
		}
	}

	for _, langChange := range languageMap {
		languageChanges = append(languageChanges, langChange)
	}

	return addedLines, deletedLines, fileChanges, languageChanges
}

func (s *AnalyzerService) detectChangeType(added, deleted int) string {
	if added > 0 && deleted == 0 {
		return "added"
	} else if added == 0 && deleted > 0 {
		return "deleted"
	} else {
		return "modified"
	}
}

func (s *AnalyzerService) detectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".go":
		return "Go"
	case ".c", ".h":
		return "C"
	case ".cpp", ".cc", ".cxx", ".hpp":
		return "C++"
	case ".ts", ".tsx":
		return "TypeScript"
	case ".js", ".jsx":
		return "JavaScript"
	case ".rs":
		return "Rust"
	case ".py":
		return "Python"
	case ".java":
		return "Java"
	case ".html":
		return "HTML"
	case ".css":
		return "CSS"
	default:
		return ""
	}
}

func (s *AnalyzerService) detectCommitType(message string) string {
	message = strings.ToLower(message)
	if strings.HasPrefix(message, "feat") {
		return "feat"
	} else if strings.HasPrefix(message, "fix") {
		return "fix"
	} else if strings.HasPrefix(message, "docs") {
		return "docs"
	} else if strings.HasPrefix(message, "style") {
		return "style"
	} else if strings.HasPrefix(message, "refactor") {
		return "refactor"
	} else if strings.HasPrefix(message, "test") {
		return "test"
	} else if strings.HasPrefix(message, "chore") {
		return "chore"
	} else {
		return "other"
	}
}

func (s *AnalyzerService) CleanupOldData(repoKey string, keepDays int) error {
	if err := s.commitAnalysisDAO.DeleteOldCommitAnalyses(repoKey, keepDays); err != nil {
		return err
	}
	if err := s.commitAnalysisDAO.DeleteOldCommitPatterns(repoKey, keepDays); err != nil {
		return err
	}
	return nil
}
