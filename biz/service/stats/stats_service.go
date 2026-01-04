package stats

import (
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/service/git"

	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type StatsService struct {
	Git *git.GitService
}

var StatsSvc *StatsService

func InitStatsService() {
	StatsSvc = &StatsService{
		Git: git.NewGitService(),
	}
}

// ParseCommits parses raw git log output into Commit structs
func (s *StatsService) ParseCommits(raw string) []domain.Commit {
	var commits []domain.Commit
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}

		t, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[3])

		commits = append(commits, domain.Commit{
			Hash:      parts[0],
			Author:    parts[1],
			Email:     parts[2],
			Date:      t,
			Timestamp: t.Unix(),
			Message:   parts[4],
		})
	}
	return commits
}

type ActivityStat struct {
	Name  string
	Trend map[string]int
}

// CalculateStats computes effective line counts per author
func (s *StatsService) CalculateStats(path, branch, since, until string) (*api.StatsResponse, error) {
	files, err := s.Git.GetRepoFiles(path, branch)
	if err != nil {
		return nil, err
	}

	// Parse dates
	var sinceTime, untilTime time.Time
	if since != "" {
		sinceTime, _ = time.Parse("2006-01-02", since)
	}
	if until != "" {
		untilTime, _ = time.Parse("2006-01-02", until)
		// Set until to end of day
		untilTime = untilTime.Add(24*time.Hour - time.Nanosecond)
	}

	authorStats := make(map[string]*api.AuthorStat)
	var mu sync.Mutex

	// Worker pool to process files
	// Limiting concurrency to avoid overwhelming system
	sem := make(chan struct{}, 10)
	var wg sync.WaitGroup

	for _, file := range files {
		if strings.TrimSpace(file) == "" {
			continue
		}

		wg.Add(1)
		sem <- struct{}{}
		go func(f string) {
			defer wg.Done()
			defer func() { <-sem }()

			rawBlame, err := s.Git.BlameFile(path, branch, f)
			if err != nil {
				return
			}

			lines := s.parseBlame(rawBlame, f)

			mu.Lock()
			defer mu.Unlock()

			for _, line := range lines {
				// Date Filter
				if !sinceTime.IsZero() && line.Date.Before(sinceTime) {
					continue
				}
				if !untilTime.IsZero() && line.Date.After(untilTime) {
					continue
				}

				if _, exists := authorStats[line.Email]; !exists {
					authorStats[line.Email] = &api.AuthorStat{
						Name:      line.Author,
						Email:     line.Email,
						FileTypes: make(map[string]int),
						TimeTrend: make(map[string]int),
					}
				}

				stat := authorStats[line.Email]
				stat.TotalLines++
				stat.FileTypes[line.Extension]++
			}
		}(file)
	}

	wg.Wait()

	// Calculate Activity Trend using git log
	activityTrends, err := s.getContributionStats(path, branch, since, until)
	if err == nil {
		for email, data := range activityTrends {
			if _, exists := authorStats[email]; !exists {
				authorStats[email] = &api.AuthorStat{
					Name:       data.Name,
					Email:      email,
					FileTypes:  make(map[string]int),
					TimeTrend:  data.Trend,
					TotalLines: 0,
				}
			} else {
				authorStats[email].TimeTrend = data.Trend
			}
		}
	}

	// Convert map to slice
	resp := &api.StatsResponse{
		Authors: make([]*api.AuthorStat, 0, len(authorStats)),
	}
	for _, stat := range authorStats {
		resp.Authors = append(resp.Authors, stat)
		resp.TotalLines += stat.TotalLines
	}

	return resp, nil
}

func (s *StatsService) parseBlame(result *gogit.BlameResult, filename string) []domain.LineStat {
	var stats []domain.LineStat

	ext := strings.ToLower(filepath.Ext(filename))
	if len(ext) > 0 {
		ext = ext[1:] // remove dot
	} else {
		ext = "unknown"
	}

	for _, line := range result.Lines {
		if s.isEffectiveLine(line.Text, ext) {
			stats = append(stats, domain.LineStat{
				Author:    line.Author,
				Email:     line.Author, // go-git Line.Author is typically the email
				Date:      line.Date,
				Extension: ext,
			})
		}
	}

	return stats
}

func (s *StatsService) isEffectiveLine(content, ext string) bool {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return false
	}

	// Basic comment filtering
	// This is not perfect but covers common cases
	if strings.HasPrefix(trimmed, "//") ||
		strings.HasPrefix(trimmed, "#") ||
		strings.HasPrefix(trimmed, "--") ||
		strings.HasPrefix(trimmed, "/*") ||
		strings.HasPrefix(trimmed, "*") {
		return false
	}

	return true
}

func (s *StatsService) getContributionStats(path, branch, since, until string) (map[string]*ActivityStat, error) {
	// Parse dates
	var sinceTime, untilTime time.Time
	if since != "" {
		sinceTime, _ = time.Parse("2006-01-02", since)
	}
	if until != "" {
		untilTime, _ = time.Parse("2006-01-02", until)
		untilTime = untilTime.Add(24*time.Hour - time.Nanosecond)
	}

	cIter, err := s.Git.GetLogIterator(path, branch)
	if err != nil {
		return nil, err
	}

	results := make(map[string]*ActivityStat)

	err = cIter.ForEach(func(c *object.Commit) error {
		if !untilTime.IsZero() && c.Author.When.After(untilTime) {
			return nil
		}
		if !sinceTime.IsZero() && c.Author.When.Before(sinceTime) {
			return nil
		}

		// Skip merges
		if len(c.ParentHashes) > 1 {
			return nil
		}

		stats, err := c.Stats()
		if err != nil {
			return nil
		}

		email := c.Author.Email
		name := c.Author.Name
		date := c.Author.When.Format("2006-01-02")

		added := 0
		for _, fs := range stats {
			added += fs.Addition
		}

		if _, ok := results[email]; !ok {
			results[email] = &ActivityStat{
				Name:  name,
				Trend: make(map[string]int),
			}
		}
		results[email].Trend[date] += added

		return nil
	})

	return results, nil
}
