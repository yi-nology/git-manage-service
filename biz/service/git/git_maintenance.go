package git

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	for i := 0; i < len(stats); i++ {
		for j := i + 1; j < len(stats); j++ {
			if stats[j].maxSize > stats[i].maxSize {
				stats[i], stats[j] = stats[j], stats[i]
			}
		}
	}
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
		sizeCmd := exec.Command("git", "cat-file", "-s", ref)
		sizeCmd.Dir = repoPath
		sizeOutput, err := sizeCmd.CombinedOutput()
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
		diffCmd := exec.Command("git", "diff-tree", "--no-commit-id", "-r", ref)
		diffCmd.Dir = repoPath
		diffOutput, err := diffCmd.CombinedOutput()
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
			sizeCmd := exec.Command("git", "cat-file", "-s", blobSHA)
			sizeCmd.Dir = repoPath
			sizeOutput, err := sizeCmd.CombinedOutput()
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

	batchCmd := exec.Command("git", "cat-file", "--batch-check", "--batch-all-objects")
	batchCmd.Dir = repoPath
	batchOutput, err := batchCmd.CombinedOutput()
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

	revCmd := exec.Command("git", "rev-list", "--objects", "--all")
	revCmd.Dir = repoPath
	revOutput, err := revCmd.CombinedOutput()
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

	reflogCmd := exec.Command("git", "reflog", "--format=%H")
	reflogCmd.Dir = repoPath
	reflogOutput, err := reflogCmd.CombinedOutput()
	if err != nil || strings.TrimSpace(string(reflogOutput)) == "" {
		return result
	}

	seen := make(map[string]bool)
	for _, commitSHA := range strings.Split(strings.TrimSpace(string(reflogOutput)), "\n") {
		commitSHA = strings.TrimSpace(commitSHA)
		if commitSHA == "" {
			continue
		}
		treeCmd := exec.Command("git", "ls-tree", "-r", commitSHA)
		treeCmd.Dir = repoPath
		treeOutput, err := treeCmd.CombinedOutput()
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
		exec.Command("git", "add", "-A").Run()
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
			exec.Command("git", "reset", "--soft", "HEAD~1").Run()
		}
		return fmt.Errorf("filter-branch failed: %w, output: %s", err, string(output))
	}
	appendLog("filter-branch 完成")

	if needCleanup {
		appendLog("撤销临时提交...")
		exec.Command("git", "reset", "--soft", "HEAD~1").Run()
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
			exec.Command("git", "tag", "-f", tag, "HEAD").Run()
			appendLog("更新 tag: " + tag)
		}
	}
	appendLog("清理 backup refs...")
	cmd = exec.Command("git", "for-each-ref", "--format=%(refname)", "refs/original/")
	cmd.Dir = repoPath
	refsOutput, _ := cmd.CombinedOutput()
	for _, ref := range strings.Split(strings.TrimSpace(string(refsOutput)), "\n") {
		if ref != "" {
			exec.Command("git", "update-ref", "-d", ref).Run()
		}
	}
	appendLog("清理 reflog...")
	exec.Command("git", "reflog", "expire", "--expire=now", "--all").Run()
	appendLog("执行 gc --prune=now ...")
	cmd = exec.Command("git", "gc", "--prune=now", "--aggressive")
	cmd.Dir = repoPath
	gcOutput, gcErr := cmd.CombinedOutput()
	if gcErr != nil {
		appendLog("gc 警告: " + string(gcOutput))
	} else {
		appendLog("gc 完成")
	}
	afterSnap := s.TakeSnapshot(repoPath)
	afterJSON, _ := json.Marshal(afterSnap)
	dao.UpdateStatus(taskID, "success", "", string(afterJSON))
	appendLog("仓库瘦身完成！")
	return nil
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
