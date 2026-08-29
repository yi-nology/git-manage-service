package stats

import (
	"bufio"
	"cmp"
	"fmt"
	"log"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/timefmt"
)

type StatsStatus string

const (
	StatusProcessing StatsStatus = "processing"
	StatusReady      StatsStatus = "ready"
	StatusFailed     StatsStatus = "failed"
)

type StatsCacheItem struct {
	mu        sync.RWMutex
	Status    StatsStatus
	Data      *api.StatsResponse
	Error     error
	CreatedAt time.Time
	Progress  string // e.g. "Processed 100 commits..."
}

// snapshot returns a consistent copy of the cache item's readable fields.
// Fields are mutated by updateCache under mu, so readers must hold the RLock.
func (it *StatsCacheItem) snapshot() (*api.StatsResponse, StatsStatus, error, string, time.Time) {
	it.mu.RLock()
	defer it.mu.RUnlock()
	return it.Data, it.Status, it.Error, it.Progress, it.CreatedAt
}

type StatsService struct {
	Git   *git.GitService
	cache sync.Map // map[string]*StatsCacheItem
	wg    sync.WaitGroup
}

var StatsSvc *StatsService

func InitStatsService() {
	StatsSvc = &StatsService{
		Git: git.NewGitService(),
	}
}

func (s *StatsService) SyncRepoHeadStatsAsync(repoID uint, path string) {
	if s == nil {
		return
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		head, err := s.Git.GetHeadBranch(path)
		if err == nil && head != "" {
			s.SyncRepoStats(repoID, path, head)
		}
	}()
}

func (s *StatsService) Wait() {
	if s == nil {
		return
	}
	s.wg.Wait()
}

// SyncRepoStats performs background synchronization of repository statistics
// It fetches new commits since the last checkpoint and saves them to DB.
func (s *StatsService) SyncRepoStats(repoID uint, path, branch string) {
	log.Printf("[StatsSync] Starting sync for repo %d (%s)...", repoID, branch)

	// 1. Get latest commit time from DB (Checkpoint)
	commitStatDAO := db.NewCommitStatDAO()
	lastTime, err := commitStatDAO.FindLatestCommitTime(repoID)
	if err != nil {
		log.Printf("[StatsSync] Failed to get latest commit time: %v", err)
		return
	}

	log.Printf("[StatsSync] Resuming from %v", lastTime)

	// 2. Get git log numstat stream (CLI-based, no go-git dependency)
	stream, err := s.Git.GetLogStatsStream(path, branch)
	if err != nil {
		log.Printf("[StatsSync] Failed to get git log stream: %v", err)
		return
	}
	defer stream.Close()

	var batch []*po.CommitStat
	batchSize := 50

	// 3. Parse the stream: COMMIT|hash|name|email|timestamp lines interleaved
	// with numstat "<added>\t<deleted>\t<file>" lines.
	scanner := bufio.NewScanner(stream)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var curHash, curName, curEmail string
	var curTime time.Time
	var curAdd, curDel int
	commitStarted := false

	flushCommit := func() {
		if !commitStarted {
			return
		}
		// Skip commits older than the checkpoint
		if !lastTime.IsZero() && curTime.Before(lastTime) {
			return
		}
		batch = append(batch, &po.CommitStat{
			RepoID:      repoID,
			CommitHash:  curHash,
			AuthorName:  curName,
			AuthorEmail: curEmail,
			CommitTime:  curTime,
			Additions:   curAdd,
			Deletions:   curDel,
		})
		if len(batch) >= batchSize {
			if err := commitStatDAO.BatchSave(batch); err != nil {
				log.Printf("[StatsSync] Failed to save batch: %v", err)
			}
			batch = nil
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		if strings.HasPrefix(line, "COMMIT|") {
			// Flush the previous commit
			flushCommit()

			parts := strings.SplitN(line, "|", 5)
			if len(parts) >= 5 {
				curHash = parts[1]
				curName = parts[2]
				curEmail = parts[3]
				ts, _ := strconv.ParseInt(parts[4], 10, 64)
				curTime = time.Unix(ts, 0)
			}
			curAdd, curDel = 0, 0
			commitStarted = true
			continue
		}

		// numstat line: "<added>\t<deleted>\t<file>"
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		added, err1 := strconv.Atoi(fields[0])
		deleted, err2 := strconv.Atoi(fields[1])
		if err1 != nil || err2 != nil {
			continue // binary files show "-"
		}
		curAdd += added
		curDel += deleted
	}

	// Flush the last commit
	flushCommit()

	// Flush remaining batch
	if len(batch) > 0 {
		if err := commitStatDAO.BatchSave(batch); err != nil {
			log.Printf("[StatsSync] Failed to save final batch: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("[StatsSync] Error during stream scan: %v", err)
	}

	log.Printf("[StatsSync] Completed sync for repo %d", repoID)
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

	// Sort by Timestamp Descending (Newest First)
	slices.SortFunc(commits, func(a, b domain.Commit) int {
		return cmp.Compare(b.Timestamp, a.Timestamp)
	})

	return commits
}

type ActivityStat struct {
	Name  string
	Trend map[string]int
}

// GetStats retrieves stats from cache or triggers calculation
func (s *StatsService) GetStats(path, branch, since, until string) (*api.StatsResponse, StatsStatus, error, string) {
	key := fmt.Sprintf("%s:%s:%s:%s", path, branch, since, until)

	// 1. Check cache
	if val, ok := s.cache.Load(key); ok {
		item := val.(*StatsCacheItem)
		data, status, err, progress, created := item.snapshot()
		// Simple TTL: 1 hour
		if time.Since(created) < time.Hour {
			return data, status, err, progress
		}
	}

	// 2. Initialize cache item (Processing)
	newItem := &StatsCacheItem{
		Status:    StatusProcessing,
		CreatedAt: time.Now(),
		Progress:  "Initializing...",
	}
	// Use LoadOrStore to prevent duplicate concurrent calculations
	actual, loaded := s.cache.LoadOrStore(key, newItem)

	if loaded {
		item := actual.(*StatsCacheItem)
		data, status, err, progress, _ := item.snapshot()
		return data, status, err, progress
	}

	// 3. Start async calculation
	go func() {
		data, err := s.calculateStatsFast(path, branch, since, until, key)
		if err != nil {
			s.updateCache(key, func(item *StatsCacheItem) {
				item.Status = StatusFailed
				item.Error = err
			})
		} else {
			s.updateCache(key, func(item *StatsCacheItem) {
				item.Status = StatusReady
				item.Data = data
				item.Progress = "Completed"
			})
		}
	}()

	return nil, StatusProcessing, nil, "Initializing..."
}

func (s *StatsService) updateCache(key string, update func(*StatsCacheItem)) {
	if val, ok := s.cache.Load(key); ok {
		item := val.(*StatsCacheItem)
		item.mu.Lock()
		defer item.mu.Unlock()
		update(item)
	}
}

// calculateStatsFast computes stats using git log --numstat (Fast, No Blame)
func (s *StatsService) calculateStatsFast(path, branch, since, until, cacheKey string) (*api.StatsResponse, error) {
	// Parse dates
	var sinceTime, untilTime time.Time
	if since != "" {
		sinceTime, _ = time.Parse(timefmt.LayoutDate, since)
	}
	if until != "" {
		untilTime, _ = time.Parse(timefmt.LayoutDate, until)
		// Set until to end of day
		untilTime = untilTime.Add(24*time.Hour - time.Nanosecond)
	}

	// 1. Get raw log stats stream
	stream, err := s.Git.GetLogStatsStream(path, branch)
	if err != nil {
		return nil, err
	}
	defer stream.Close()

	authorStats := make(map[string]*api.AuthorStat)

	scanner := bufio.NewScanner(stream)
	// Increase buffer size for long lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var currentEmail, currentName string
	var currentDate time.Time

	commitCount := 0
	lastUpdate := time.Now()

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		if strings.HasPrefix(line, "COMMIT|") {
			commitCount++

			// Update progress every 100 commits or 1 second
			if commitCount%100 == 0 || time.Since(lastUpdate) > time.Second {
				s.updateCache(cacheKey, func(item *StatsCacheItem) {
					item.Progress = fmt.Sprintf("Processed %d commits...", commitCount)
				})
				lastUpdate = time.Now()
			}

			parts := strings.Split(line, "|")
			if len(parts) >= 5 {
				// COMMIT|Hash|Name|Email|Timestamp
				currentName = parts[2]
				currentEmail = parts[3]
				ts, _ := strconv.ParseInt(parts[4], 10, 64)
				currentDate = time.Unix(ts, 0)
			}
			continue
		}

		// ... (rest of the loop)

		// Date Filter
		if !sinceTime.IsZero() && currentDate.Before(sinceTime) {
			continue
		}
		if !untilTime.IsZero() && currentDate.After(untilTime) {
			continue
		}

		// Parse numstat: "added deleted filename"
		// Note: binary files might show "-"
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		added, err1 := strconv.Atoi(parts[0])
		deleted, err2 := strconv.Atoi(parts[1])

		// Skip binary files or parse errors
		if err1 != nil || err2 != nil {
			continue
		}

		filename := parts[2]
		ext := strings.ToLower(filepath.Ext(filename))
		if len(ext) > 0 {
			ext = ext[1:] // remove dot
		} else {
			ext = "unknown"
		}

		if _, exists := authorStats[currentEmail]; !exists {
			authorStats[currentEmail] = &api.AuthorStat{
				Name:      currentName,
				Email:     currentEmail,
				FileTypes: make(map[string]int),
				TimeTrend: make(map[string]int),
			}
		}

		stat := authorStats[currentEmail]
		// Use Net Contribution as "Total Lines" approximation
		stat.TotalLines += (added - deleted)
		stat.FileTypes[ext] += (added - deleted)

		dateStr := currentDate.Format(timefmt.LayoutDate)
		stat.TimeTrend[dateStr] += (added - deleted)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
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
