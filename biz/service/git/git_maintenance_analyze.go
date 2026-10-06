package git

import (
	"cmp"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/model/api"
)

func (s *MaintenanceService) FindFilesInfo(repoPath string, filePaths []string) ([]api.LargeFileEntry, error) {
	if len(filePaths) == 0 {
		return nil, nil
	}

	pathSet := make(map[string]bool, len(filePaths))
	for _, p := range filePaths {
		pathSet[p] = true
	}

	args := []string{"rev-list", "--objects", "--all", "--"}
	args = append(args, filePaths...)
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	revOutput, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git rev-list failed: %w", err)
	}

	shaToPath := make(map[string]string)
	for _, line := range strings.Split(string(revOutput), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && pathSet[fields[1]] {
			shaToPath[fields[0]] = fields[1]
		}
	}

	if len(shaToPath) == 0 {
		return nil, nil
	}

	shaList := make([]string, 0, len(shaToPath))
	for sha := range shaToPath {
		shaList = append(shaList, sha)
	}

	input := strings.Join(shaList, "\n") + "\n"
	cmd = exec.Command("git", "cat-file", "--batch-check")
	cmd.Dir = repoPath
	cmd.Stdin = strings.NewReader(input)
	batchOutput, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git cat-file failed: %w", err)
	}

	shaSize := make(map[string]int64)
	for _, line := range strings.Split(string(batchOutput), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[1] == "blob" {
			if size, err := strconv.ParseInt(fields[2], 10, 64); err == nil {
				shaSize[fields[0]] = size
			}
		}
	}

	type pathStat struct {
		maxSize int64
		count   int
	}
	pathStats := make(map[string]*pathStat)
	for sha, path := range shaToPath {
		if size, ok := shaSize[sha]; ok {
			stat, exists := pathStats[path]
			if !exists {
				stat = &pathStat{}
				pathStats[path] = stat
			}
			stat.count++
			if size > stat.maxSize {
				stat.maxSize = size
			}
		}
	}

	result := make([]api.LargeFileEntry, 0, len(pathStats))
	for _, fp := range filePaths {
		stat, ok := pathStats[fp]
		if !ok {
			continue
		}
		_, statErr := os.Stat(filepath.Join(repoPath, fp))
		result = append(result, api.LargeFileEntry{
			Path:        fp,
			Size:        formatSize(stat.maxSize),
			SizeBytes:   stat.maxSize,
			Exists:      statErr == nil,
			CommitCount: stat.count,
			Source:      "history",
		})
	}

	return result, nil
}

func (s *MaintenanceService) AnalyzeHealthForPaths(repoPath string, threshold int64, filePaths []string) (*api.RepoHealthReport, error) {
	if threshold <= 0 {
		threshold = 1 * 1024 * 1024
	}
	snap := s.TakeSnapshot(repoPath)
	report := &api.RepoHealthReport{
		GitDirSize:      snap.GitDirSize,
		GitDirSizeBytes: snap.GitDirSizeBytes,
		LooseObjects:    snap.LooseObjects,
		PackFiles:       snap.PackFiles,
		InPackObjects:   snap.InPackObjects,
		CommitCount:     snap.CommitCount,
		BranchCount:     snap.BranchCount,
		TagCount:        snap.TagCount,
		Threshold:       threshold,
		ThresholdHuman:  formatSize(threshold),
	}

	report.GitDirBreakdown = s.ScanGitDirBreakdown(repoPath)

	largeFiles, err := s.FindFilesInfo(repoPath, filePaths)
	if err != nil {
		largeFiles = []api.LargeFileEntry{}
	}
	report.LargeFiles = largeFiles

	return report, nil
}

func (s *MaintenanceService) AnalyzeHealth(repoPath string, threshold int64, excludes []string) (*api.RepoHealthReport, error) {
	if threshold <= 0 {
		threshold = 1 * 1024 * 1024
	}
	snap := s.TakeSnapshot(repoPath)
	report := &api.RepoHealthReport{
		GitDirSize:      snap.GitDirSize,
		GitDirSizeBytes: snap.GitDirSizeBytes,
		LooseObjects:    snap.LooseObjects,
		PackFiles:       snap.PackFiles,
		InPackObjects:   snap.InPackObjects,
		CommitCount:     snap.CommitCount,
		BranchCount:     snap.BranchCount,
		TagCount:        snap.TagCount,
		Threshold:       threshold,
		ThresholdHuman:  formatSize(threshold),
		Excludes:        excludes,
	}

	report.GitDirBreakdown = s.ScanGitDirBreakdown(repoPath)
	report.StashEntries = s.FindStashEntries(repoPath, threshold)

	allFiles := []api.LargeFileEntry{}

	historyFiles, err := s.FindLargeFiles(repoPath, threshold)
	if err == nil {
		for i := range historyFiles {
			historyFiles[i].Source = "history"
		}
		allFiles = append(allFiles, historyFiles...)
	}

	stashFiles := s.FindStashLargeObjects(repoPath, threshold)
	allFiles = append(allFiles, stashFiles...)

	reflogFiles := s.FindReflogLargeObjects(repoPath, threshold)
	allFiles = append(allFiles, reflogFiles...)

	if allFiles == nil {
		allFiles = []api.LargeFileEntry{}
	}

	if len(excludes) > 0 {
		filtered := make([]api.LargeFileEntry, 0, len(allFiles))
		for _, f := range allFiles {
			if matchExclude(f.Path, excludes) {
				continue
			}
			filtered = append(filtered, f)
		}
		allFiles = filtered
	}

	report.LargeFiles = allFiles
	return report, nil
}

func (s *MaintenanceService) FindLargeFiles(repoPath string, threshold int64) ([]api.LargeFileEntry, error) {
	if threshold <= 0 {
		threshold = 1 * 1024 * 1024
	}
	cmd := exec.Command("git", "rev-list", "--objects", "--all")
	cmd.Dir = repoPath
	revOutput, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git rev-list failed: %w", err)
	}
	cmd = exec.Command("git", "cat-file", "--batch-check", "--batch-all-objects")
	cmd.Dir = repoPath
	batchOutput, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git cat-file failed: %w", err)
	}
	blobSize := make(map[string]int64)
	for _, line := range strings.Split(string(batchOutput), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[1] == "blob" {
			if size, err := strconv.ParseInt(fields[2], 10, 64); err == nil {
				blobSize[fields[0]] = size
			}
		}
	}
	fileBlobs := make(map[string][]string)
	for _, line := range strings.Split(string(revOutput), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 {
			sha := fields[0]
			path := fields[1]
			if _, ok := blobSize[sha]; ok {
				fileBlobs[path] = append(fileBlobs[path], sha)
			}
		}
	}
	type fileStat struct {
		path    string
		maxSize int64
		count   int
	}
	var stats []fileStat
	for path, shas := range fileBlobs {
		var maxSz int64
		for _, sha := range shas {
			if sz, ok := blobSize[sha]; ok {
				maxSz = max(maxSz, sz)
			}
		}
		if maxSz >= threshold {
			stats = append(stats, fileStat{path: path, maxSize: maxSz, count: len(shas)})
		}
	}
	slices.SortFunc(stats, func(a, b fileStat) int { return cmp.Compare(b.maxSize, a.maxSize) })
	if len(stats) > 50 {
		stats = stats[:50]
	}
	var result []api.LargeFileEntry
	for _, st := range stats {
		_, err := os.Stat(filepath.Join(repoPath, st.path))
		result = append(result, api.LargeFileEntry{
			Path:        st.path,
			Size:        formatSize(st.maxSize),
			SizeBytes:   st.maxSize,
			Exists:      err == nil,
			CommitCount: st.count,
		})
	}
	return result, nil
}

func (s *MaintenanceService) ScanGitDirBreakdown(repoPath string) *api.GitDirBreakdown {
	gitDir := filepath.Join(repoPath, ".git")
	breakdown := &api.GitDirBreakdown{}

	packDir := filepath.Join(gitDir, "objects", "pack")
	if info, err := os.Stat(packDir); err == nil && info.IsDir() {
		if size, err := dirSize(packDir); err == nil {
			breakdown.PackDirSize = formatSize(size)
			breakdown.PackDirSizeBytes = size
		}
	}

	looseSize, _ := s.calcLooseObjSize(filepath.Join(gitDir, "objects"))
	breakdown.LooseObjSize = formatSize(looseSize)
	breakdown.LooseObjSizeBytes = looseSize

	logsDir := filepath.Join(gitDir, "logs")
	if info, err := os.Stat(logsDir); err == nil && info.IsDir() {
		if size, err := dirSize(logsDir); err == nil {
			breakdown.ReflogSize = formatSize(size)
			breakdown.ReflogSizeBytes = size
		}
	}

	cmd := exec.Command("git", "stash", "list")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if strings.TrimSpace(string(output)) == "" {
			breakdown.StashCount = 0
		} else {
			breakdown.StashCount = len(lines)
		}
	}

	totalGitSize := int64(0)
	if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
		if size, err := dirSize(gitDir); err == nil {
			totalGitSize = size
		}
	}
	accounted := breakdown.PackDirSizeBytes + breakdown.LooseObjSizeBytes + breakdown.ReflogSizeBytes
	other := totalGitSize - accounted
	if other < 0 {
		other = 0
	}
	breakdown.OtherSize = formatSize(other)
	breakdown.OtherSizeBytes = other

	return breakdown
}

func (s *MaintenanceService) calcLooseObjSize(objectsDir string) (int64, error) {
	var totalSize int64
	entries, err := os.ReadDir(objectsDir)
	if err != nil {
		return 0, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			totalSize += info.Size()
			continue
		}
		name := entry.Name()
		if len(name) == 2 && name != "pa" && name != "in" {
			sub := filepath.Join(objectsDir, name)
			if size, err := dirSize(sub); err == nil {
				totalSize += size
			}
		}
	}
	return totalSize, nil
}

func (s *MaintenanceService) FindStashEntries(repoPath string, threshold int64) []api.StashEntry {
	var entries []api.StashEntry
	cmd := exec.Command("git", "stash", "list")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil || strings.TrimSpace(string(output)) == "" {
		return entries
	}
	for i, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if line == "" {
			continue
		}
		ref := fmt.Sprintf("stash@{%d}", i)
		sizeOutput, err := runGitCmd(repoPath, "cat-file", "-s", ref)
		var sizeBytes int64
		if err == nil {
			sizeBytes, _ = strconv.ParseInt(strings.TrimSpace(string(sizeOutput)), 10, 64)
		}
		msg := line
		if idx := strings.Index(line, ": "); idx >= 0 {
			msg = line[idx+2:]
		}
		entry := api.StashEntry{
			Index:     i,
			Message:   msg,
			Size:      formatSize(sizeBytes),
			SizeBytes: sizeBytes,
		}
		entries = append(entries, entry)
	}
	return entries
}

func (s *MaintenanceService) FindStashLargeObjects(repoPath string, threshold int64) []api.LargeFileEntry {
	var result []api.LargeFileEntry
	cmd := exec.Command("git", "stash", "list")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil || strings.TrimSpace(string(output)) == "" {
		return result
	}
	stashCount := len(strings.Split(strings.TrimSpace(string(output)), "\n"))

	for i := 0; i < stashCount; i++ {
		ref := fmt.Sprintf("stash@{%d}", i)
		diffOutput, err := runGitCmd(repoPath, "diff-tree", "--no-commit-id", "-r", ref)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(diffOutput), "\n") {
			fields := strings.Split(line, "\t")
			if len(fields) < 2 {
				continue
			}
			meta := strings.Fields(fields[0])
			if len(meta) < 4 {
				continue
			}
			blobSHA := meta[3]
			sizeOutput, err := runGitCmd(repoPath, "cat-file", "-s", blobSHA)
			if err != nil {
				continue
			}
			sizeBytes, _ := strconv.ParseInt(strings.TrimSpace(string(sizeOutput)), 10, 64)
			path := strings.Join(fields[1:], "\t")
			if sizeBytes >= threshold {
				_, statErr := os.Stat(filepath.Join(repoPath, path))
				result = append(result, api.LargeFileEntry{
					Path:        fmt.Sprintf("stash@{%d}:%s", i, path),
					Size:        formatSize(sizeBytes),
					SizeBytes:   sizeBytes,
					Exists:      statErr == nil,
					CommitCount: 1,
					Source:      "stash",
				})
			}
		}
	}
	return result
}

func (s *MaintenanceService) FindReflogLargeObjects(repoPath string, threshold int64) []api.LargeFileEntry {
	var result []api.LargeFileEntry
	blobSize := make(map[string]int64)

	batchOutput, err := runGitCmd(repoPath, "cat-file", "--batch-check", "--batch-all-objects")
	if err != nil {
		return result
	}
	for _, line := range strings.Split(string(batchOutput), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[1] == "blob" {
			if size, err := strconv.ParseInt(fields[2], 10, 64); err == nil && size >= threshold {
				blobSize[fields[0]] = size
			}
		}
	}

	revOutput, err := runGitCmd(repoPath, "rev-list", "--objects", "--all")
	if err != nil {
		return result
	}
	knownBlobs := make(map[string]bool)
	for _, line := range strings.Split(string(revOutput), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 {
			knownBlobs[fields[0]] = true
		}
	}

	reflogOutput, err := runGitCmd(repoPath, "reflog", "--format=%H")
	if err != nil || strings.TrimSpace(string(reflogOutput)) == "" {
		return result
	}

	seen := make(map[string]bool)
	for _, commitSHA := range strings.Split(strings.TrimSpace(string(reflogOutput)), "\n") {
		commitSHA = strings.TrimSpace(commitSHA)
		if commitSHA == "" {
			continue
		}
		treeOutput, err := runGitCmd(repoPath, "ls-tree", "-r", commitSHA)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(treeOutput), "\n") {
			fields := strings.SplitN(line, "\t", 2)
			if len(fields) < 2 {
				continue
			}
			meta := strings.Fields(fields[0])
			if len(meta) < 3 {
				continue
			}
			blobSHA := meta[2]
			path := fields[1]
			size, ok := blobSize[blobSHA]
			if !ok || knownBlobs[blobSHA] {
				continue
			}
			key := blobSHA + ":" + path
			if seen[key] {
				continue
			}
			seen[key] = true
			_, statErr := os.Stat(filepath.Join(repoPath, path))
			result = append(result, api.LargeFileEntry{
				Path:        path,
				Size:        formatSize(size),
				SizeBytes:   size,
				Exists:      statErr == nil,
				CommitCount: 1,
				Source:      "reflog",
			})
		}
	}
	return result
}

// validateSlimPath 限制能进入 filter-branch index-filter 的路径。
func (s *MaintenanceService) FindByPrefix(repoPath string, prefixes []string) ([]api.PrefixFileEntry, error) {
	if len(prefixes) == 0 {
		return nil, nil
	}

	cmd := exec.Command("git", "rev-list", "--objects", "--all")
	cmd.Dir = repoPath
	revOutput, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git rev-list failed: %w", err)
	}

	cmd = exec.Command("git", "cat-file", "--batch-check", "--batch-all-objects")
	cmd.Dir = repoPath
	batchOutput, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git cat-file failed: %w", err)
	}

	blobSize := make(map[string]int64)
	for _, line := range strings.Split(string(batchOutput), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[1] == "blob" {
			if size, err := strconv.ParseInt(fields[2], 10, 64); err == nil {
				blobSize[fields[0]] = size
			}
		}
	}

	type pathStat struct {
		maxSize int64
		count   int
	}
	pathStats := make(map[string]*pathStat)
	for _, line := range strings.Split(string(revOutput), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		sha := fields[0]
		path := fields[1]
		if !slices.ContainsFunc(prefixes, func(prefix string) bool {
			return strings.HasPrefix(path, prefix)
		}) {
			continue
		}
		size, ok := blobSize[sha]
		if !ok {
			continue
		}
		stat, exists := pathStats[path]
		if !exists {
			stat = &pathStat{}
			pathStats[path] = stat
		}
		stat.count++
		if size > stat.maxSize {
			stat.maxSize = size
		}
	}

	result := make([]api.PrefixFileEntry, 0, len(pathStats))
	for path, stat := range pathStats {
		_, statErr := os.Stat(filepath.Join(repoPath, path))
		result = append(result, api.PrefixFileEntry{
			Path:        path,
			Size:        formatSize(stat.maxSize),
			SizeBytes:   stat.maxSize,
			Exists:      statErr == nil,
			CommitCount: stat.count,
		})
	}

	slices.SortFunc(result, func(a, b api.PrefixFileEntry) int {
		return cmp.Compare(b.SizeBytes, a.SizeBytes)
	})

	return result, nil
}
