package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/model/api"
)

type MaintenanceService struct{}

func NewMaintenanceService() *MaintenanceService {
	return &MaintenanceService{}
}

func (s *MaintenanceService) AnalyzeHealth(repoPath string) (*api.RepoHealthReport, error) {
	report := &api.RepoHealthReport{}

	gitDir := filepath.Join(repoPath, ".git")
	if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
		if size, err := dirSize(gitDir); err == nil {
			report.GitDirSize = formatSize(size)
			report.GitDirSizeBytes = size
		}
	}

	cmd := exec.Command("git", "count-objects", "-v")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err == nil {
		for _, line := range strings.Split(string(output), "\n") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			switch key {
			case "count":
				report.LooseObjects, _ = strconv.ParseInt(val, 10, 64)
			case "packs":
				report.PackFiles, _ = strconv.Atoi(val)
			case "in-pack":
				if n, err := strconv.ParseInt(val, 10, 64); err == nil {
					report.InPackObjects = n
				}
			}
		}
	}

	cmd = exec.Command("git", "rev-list", "--count", "HEAD")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err == nil {
		report.CommitCount, _ = strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	}

	cmd = exec.Command("git", "branch", "--list")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err == nil {
		report.BranchCount = len(strings.Split(strings.TrimSpace(string(output)), "\n"))
		if strings.TrimSpace(string(output)) == "" {
			report.BranchCount = 0
		}
	}

	cmd = exec.Command("git", "tag", "--list")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err == nil {
		report.TagCount = len(strings.Split(strings.TrimSpace(string(output)), "\n"))
		if strings.TrimSpace(string(output)) == "" {
			report.TagCount = 0
		}
	}

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
		if len(fields) >= 4 && fields[1] == "blob" {
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

func (s *MaintenanceService) SlimHistory(repoPath string, paths []string, addGitignore bool, taskID string) error {
	tm := GlobalTaskManager
	tm.AppendLog(taskID, "开始仓库瘦身...")

	if addGitignore {
		tm.AppendLog(taskID, "更新 .gitignore...")
		if err := appendToGitignore(repoPath, paths); err != nil {
			tm.AppendLog(taskID, "警告: 更新 .gitignore 失败: "+err.Error())
		}
	}

	tm.AppendLog(taskID, "执行 filter-branch 清除历史文件...")
	args := []string{"filter-branch", "--force", "--index-filter"}
	indexFilter := "git rm --cached --ignore-unmatch " + strings.Join(paths, " ")
	args = append(args, indexFilter, "--prune-empty", "--", "--all")

	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("filter-branch failed: %w, output: %s", err, string(output))
	}
	tm.AppendLog(taskID, "filter-branch 完成")

	tm.AppendLog(taskID, "更新指向旧 commit 的 tags...")
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
			tm.AppendLog(taskID, "更新 tag: "+tag)
		}
	}

	tm.AppendLog(taskID, "清理 backup refs...")
	cmd = exec.Command("git", "for-each-ref", "--format=%(refname)", "refs/original/")
	cmd.Dir = repoPath
	refsOutput, _ := cmd.CombinedOutput()
	for _, ref := range strings.Split(strings.TrimSpace(string(refsOutput)), "\n") {
		if ref != "" {
			exec.Command("git", "update-ref", "-d", ref).Run()
		}
	}

	tm.AppendLog(taskID, "清理 reflog...")
	exec.Command("git", "reflog", "expire", "--expire=now", "--all").Run()

	tm.AppendLog(taskID, "执行 gc --prune=now ...")
	cmd = exec.Command("git", "gc", "--prune=now", "--aggressive")
	cmd.Dir = repoPath
	gcOutput, gcErr := cmd.CombinedOutput()
	if gcErr != nil {
		tm.AppendLog(taskID, "gc 警告: "+string(gcOutput))
	} else {
		tm.AppendLog(taskID, "gc 完成")
	}

	tm.AppendLog(taskID, "仓库瘦身完成！")
	return nil
}

func (s *MaintenanceService) GarbageCollect(repoPath string, taskID string) error {
	tm := GlobalTaskManager
	tm.AppendLog(taskID, "开始垃圾回收...")

	tm.AppendLog(taskID, "清理 reflog...")
	exec.Command("git", "reflog", "expire", "--expire=now", "--all").Run()

	tm.AppendLog(taskID, "执行 git gc --aggressive --prune=now ...")
	cmd := exec.Command("git", "gc", "--aggressive", "--prune=now")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git gc failed: %w, output: %s", err, string(output))
	}

	tm.AppendLog(taskID, "垃圾回收完成！")
	return nil
}

func (s *MaintenanceService) AddToGitignore(repoPath string, paths []string) error {
	return appendToGitignore(repoPath, paths)
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
