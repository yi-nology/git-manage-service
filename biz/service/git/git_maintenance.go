package git

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type MaintenanceService struct{}

func NewMaintenanceService() *MaintenanceService {
	return &MaintenanceService{}
}

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
			if sz, ok := blobSize[sha]; ok && sz > maxSz {
				maxSz = sz
			}
		}
		if maxSz >= threshold {
			stats = append(stats, fileStat{path: path, maxSize: maxSz, count: len(shas)})
		}
	}
	sort.Slice(stats, func(i, j int) bool { return stats[i].maxSize > stats[j].maxSize })
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

func (s *MaintenanceService) SlimHistory(repoPath string, paths []string, addGitignore bool, taskID string) error {
	tm := GlobalTaskManager
	dao := db.NewMaintenanceDAO()
	appendLog := func(msg string) {
		tm.AppendLog(taskID, msg)
		record, _ := dao.FindByTaskID(taskID)
		if record != nil {
			t, ok := tm.GetTask(taskID)
			if ok {
				logJSON, _ := json.Marshal(t.Progress)
				dao.UpdateProgress(taskID, string(logJSON))
			}
		}
	}
	appendLog("开始仓库瘦身...")
	if addGitignore {
		appendLog("更新 .gitignore...")
		if err := appendToGitignore(repoPath, paths); err != nil {
			appendLog("警告: 更新 .gitignore 失败: " + err.Error())
		}
	}

	needCleanup := false
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = repoPath
	if statusOut, statusErr := statusCmd.CombinedOutput(); statusErr == nil && strings.TrimSpace(string(statusOut)) != "" {
		needCleanup = true
		appendLog("检测到未提交变更，临时提交...")
		addCmd := exec.Command("git", "add", "-A")
		addCmd.Dir = repoPath
		addCmd.Run()
		commitCmd := exec.Command("git", "commit", "-m", "chore: temp commit for repo slim")
		commitCmd.Dir = repoPath
		commitCmd.Run()
	}

	appendLog("执行 filter-branch 清除历史文件...")
	indexFilter := "git rm --cached --ignore-unmatch " + strings.Join(paths, " ")
	cmdStr := "git filter-branch --force --index-filter '" + strings.ReplaceAll(indexFilter, "'", "'\\''") + "' --prune-empty -- --all"
	cmd := exec.Command("bash", "-c", cmdStr)
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "FILTER_BRANCH_SQUELCH_WARNING=1", "GIT_ASKPASS=", "GIT_TERMINAL_PROMPT=0")
	cmd.Stdin = nil
	output, err := cmd.CombinedOutput()
	if err != nil {
		if needCleanup {
			resetCmd := exec.Command("git", "reset", "--soft", "HEAD~1")
			resetCmd.Dir = repoPath
			resetCmd.Run()
		}
		return fmt.Errorf("filter-branch failed: %w, output: %s", err, string(output))
	}
	appendLog("filter-branch 完成")

	if needCleanup {
		appendLog("撤销临时提交...")
		resetCmd := exec.Command("git", "reset", "--soft", "HEAD~1")
		resetCmd.Dir = repoPath
		resetCmd.Run()
	}
	appendLog("更新指向旧 commit 的 tags...")
	cmd = exec.Command("git", "tag", "-l")
	cmd.Dir = repoPath
	tagOutput, _ := cmd.CombinedOutput()
	for _, tag := range strings.Split(strings.TrimSpace(string(tagOutput)), "\n") {
		if tag == "" {
			continue
		}
		cmd = exec.Command("git", "rev-parse", tag+"^{}")
		cmd.Dir = repoPath
		commitHash, err := cmd.CombinedOutput()
		if err != nil {
			continue
		}
		commit := strings.TrimSpace(string(commitHash))
		cmd = exec.Command("git", "ls-tree", "-r", commit)
		cmd.Dir = repoPath
		treeOut, err := cmd.CombinedOutput()
		if err != nil {
			continue
		}
		needsUpdate := false
		for _, p := range paths {
			if strings.Contains(string(treeOut), "\t"+p+"\n") || strings.Contains(string(treeOut), "\t"+p+" ") {
				needsUpdate = true
				break
			}
		}
		if needsUpdate {
			tagCmd := exec.Command("git", "tag", "-f", tag, "HEAD")
			tagCmd.Dir = repoPath
			tagCmd.Run()
			appendLog("更新 tag: " + tag)
		}
	}
	appendLog("清理 backup refs...")
	cmd = exec.Command("git", "for-each-ref", "--format=%(refname)", "refs/original/")
	cmd.Dir = repoPath
	refsOutput, _ := cmd.CombinedOutput()
	for _, ref := range strings.Split(strings.TrimSpace(string(refsOutput)), "\n") {
		if ref != "" {
			delCmd := exec.Command("git", "update-ref", "-d", ref)
			delCmd.Dir = repoPath
			delCmd.Run()
		}
	}
	appendLog("清理 reflog...")
	cmd = exec.Command("git", "reflog", "expire", "--expire=now", "--all")
	cmd.Dir = repoPath
	if reflogOut, reflogErr := cmd.CombinedOutput(); reflogErr != nil {
		appendLog("reflog expire 警告: " + string(reflogOut))
	}
	appendLog("执行 gc --prune=now ...")
	cmd = exec.Command("git", "gc", "--prune=now", "--force")
	cmd.Dir = repoPath
	gcOutput, gcErr := cmd.CombinedOutput()
	if gcErr != nil {
		appendLog("gc 警告: " + string(gcOutput))
	} else {
		appendLog("gc 完成")
	}
	appendLog("执行 prune 清理不可达对象...")
	cmd = exec.Command("git", "prune", "--expire=now")
	cmd.Dir = repoPath
	if pruneOut, pruneErr := cmd.CombinedOutput(); pruneErr != nil {
		appendLog("prune 警告: " + string(pruneOut))
	} else {
		appendLog("prune 完成")
	}

	appendLog("验证清理结果...")
	if stillExist := s.verifyPathsRemoved(repoPath, paths); len(stillExist) > 0 {
		appendLog("检测到残留文件，执行二次清理: " + strings.Join(stillExist, ", "))
		reflogRetry := exec.Command("git", "reflog", "expire", "--expire=now", "--all")
		reflogRetry.Dir = repoPath
		reflogRetry.Run()
		gcRetry := exec.Command("git", "gc", "--prune=now", "--force")
		gcRetry.Dir = repoPath
		gcRetry.Run()
		pruneRetry := exec.Command("git", "prune", "--expire=now")
		pruneRetry.Dir = repoPath
		pruneRetry.Run()
		if retryExist := s.verifyPathsRemoved(repoPath, paths); len(retryExist) > 0 {
			appendLog("警告: 以下文件仍在对象库中: " + strings.Join(retryExist, ", "))
		} else {
			appendLog("二次清理成功")
		}
	} else {
		appendLog("验证通过，目标文件已从历史中移除")
	}

	afterSnap := s.TakeSnapshot(repoPath)
	afterJSON, _ := json.Marshal(afterSnap)
	dao.UpdateStatus(taskID, "success", "", string(afterJSON))
	appendLog("仓库瘦身完成！")
	return nil
}

func (s *MaintenanceService) verifyPathsRemoved(repoPath string, paths []string) []string {
	var stillExist []string
	batchOutput, err := runGitCmd(repoPath, "cat-file", "--batch-check", "--batch-all-objects")
	if err != nil {
		return stillExist
	}
	blobSizes := make(map[string]int64)
	for _, line := range strings.Split(string(batchOutput), "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 3 && fields[1] == "blob" {
			if size, err := strconv.ParseInt(fields[2], 10, 64); err == nil {
				blobSizes[fields[0]] = size
			}
		}
	}
	revOutput, err := runGitCmd(repoPath, "rev-list", "--objects", "--all")
	if err != nil {
		return stillExist
	}
	for _, line := range strings.Split(string(revOutput), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 {
			if _, ok := blobSizes[fields[0]]; ok {
				for _, p := range paths {
					if fields[1] == p {
						stillExist = append(stillExist, p)
						break
					}
				}
			}
		}
	}
	reflogOutput, err := runGitCmd(repoPath, "reflog", "--format=%H")
	if err != nil || strings.TrimSpace(string(reflogOutput)) == "" {
		return stillExist
	}
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
			parts := strings.SplitN(line, "\t", 2)
			if len(parts) < 2 {
				continue
			}
			meta := strings.Fields(parts[0])
			if len(meta) < 3 {
				continue
			}
			blobSHA := meta[2]
			filePath := parts[1]
			if _, ok := blobSizes[blobSHA]; ok {
				for _, p := range paths {
					if filePath == p {
						stillExist = append(stillExist, p)
						break
					}
				}
			}
		}
	}
	seen := make(map[string]bool)
	var unique []string
	for _, p := range stillExist {
		if !seen[p] {
			seen[p] = true
			unique = append(unique, p)
		}
	}
	return unique
}

func (s *MaintenanceService) GarbageCollect(repoPath string, taskID string) error {
	tm := GlobalTaskManager
	dao := db.NewMaintenanceDAO()
	appendLog := func(msg string) {
		tm.AppendLog(taskID, msg)
		record, _ := dao.FindByTaskID(taskID)
		if record != nil {
			t, ok := tm.GetTask(taskID)
			if ok {
				logJSON, _ := json.Marshal(t.Progress)
				dao.UpdateProgress(taskID, string(logJSON))
			}
		}
	}
	appendLog("开始垃圾回收...")
	appendLog("清理 reflog...")
	exec.Command("git", "reflog", "expire", "--expire=now", "--all").Run()
	appendLog("执行 git gc --aggressive --prune=now ...")
	cmd := exec.Command("git", "gc", "--aggressive", "--prune=now")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git gc failed: %w, output: %s", err, string(output))
	}
	afterSnap := s.TakeSnapshot(repoPath)
	afterJSON, _ := json.Marshal(afterSnap)
	dao.UpdateStatus(taskID, "success", "", string(afterJSON))
	appendLog("垃圾回收完成！")
	return nil
}

func (s *MaintenanceService) AddToGitignore(repoPath string, paths []string) error {
	return appendToGitignore(repoPath, paths)
}

func CreateMaintenanceRecord(repoID uint, opType string, repoPath string) (*po.MaintenanceRecord, error) {
	svc := NewMaintenanceService()
	beforeSnap := svc.TakeSnapshot(repoPath)
	beforeJSON, _ := json.Marshal(beforeSnap)
	now := time.Now()
	record := &po.MaintenanceRecord{
		RepoID:         repoID,
		Type:           opType,
		Status:         "pending",
		TriggerBy:      "manual",
		SnapshotBefore: string(beforeJSON),
		StartedAt:      &now,
	}
	if err := db.NewMaintenanceDAO().Create(record); err != nil {
		return nil, err
	}
	return record, nil
}

func appendToGitignore(repoPath string, paths []string) error {
	gitignorePath := filepath.Join(repoPath, ".gitignore")
	var existing []byte
	existing, err := os.ReadFile(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(existing)
	var newLines []string
	for _, p := range paths {
		line := "/" + p
		if strings.Contains(content, line) {
			continue
		}
		newLines = append(newLines, line)
	}
	if len(newLines) == 0 {
		return nil
	}
	suffix := "\n"
	if len(existing) == 0 || !strings.HasSuffix(content, "\n") {
		suffix = "\n"
	}
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString(suffix + "# Auto-added by repo slim\n")
	for _, line := range newLines {
		f.WriteString(line + "\n")
	}
	return nil
}

func dirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size, err
}

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
		matched := false
		for _, prefix := range prefixes {
			if strings.HasPrefix(path, prefix) {
				matched = true
				break
			}
		}
		if !matched {
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

	sort.Slice(result, func(i, j int) bool {
		return result[i].SizeBytes > result[j].SizeBytes
	})

	return result, nil
}

func (s *MaintenanceService) SlimHistoryByPrefix(repoPath string, prefixes []string, addGitignore bool, taskID string) error {
	tm := GlobalTaskManager
	dao := db.NewMaintenanceDAO()
	appendLog := func(msg string) {
		tm.AppendLog(taskID, msg)
		record, _ := dao.FindByTaskID(taskID)
		if record != nil {
			t, ok := tm.GetTask(taskID)
			if ok {
				logJSON, _ := json.Marshal(t.Progress)
				dao.UpdateProgress(taskID, string(logJSON))
			}
		}
	}

	appendLog("扫描匹配前缀的文件...")
	files, err := s.FindByPrefix(repoPath, prefixes)
	if err != nil {
		return fmt.Errorf("scan prefix files failed: %w", err)
	}
	if len(files) == 0 {
		appendLog("未找到匹配的文件")
		return nil
	}

	paths := make([]string, 0, len(files))
	for _, f := range files {
		paths = append(paths, f.Path)
	}
	appendLog(fmt.Sprintf("找到 %d 个匹配文件，总大小 %s", len(paths), formatSize(sumSizeBytes(files))))

	return s.SlimHistory(repoPath, paths, addGitignore, taskID)
}

func sumSizeBytes(files []api.PrefixFileEntry) int64 {
	var total int64
	for _, f := range files {
		total += f.SizeBytes
	}
	return total
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
