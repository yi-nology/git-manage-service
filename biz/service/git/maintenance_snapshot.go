package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/model/api"
)

func (s *MaintenanceService) TakeSnapshot(repoPath string) *api.MaintenanceSnapshotDTO {
	snap := &api.MaintenanceSnapshotDTO{}
	gitDir := filepath.Join(repoPath, ".git")
	if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
		if size, err := dirSize(gitDir); err == nil {
			snap.GitDirSize = formatSize(size)
			snap.GitDirSizeBytes = size
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
				snap.LooseObjects, _ = strconv.ParseInt(val, 10, 64)
			case "packs":
				snap.PackFiles, _ = strconv.Atoi(val)
			case "in-pack":
				snap.InPackObjects, _ = strconv.ParseInt(val, 10, 64)
			}
		}
	}
	cmd = exec.Command("git", "rev-list", "--count", "HEAD")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err == nil {
		snap.CommitCount, _ = strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	}
	cmd = exec.Command("git", "branch", "--list")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err == nil {
		snap.BranchCount = len(strings.Split(strings.TrimSpace(string(output)), "\n"))
		if strings.TrimSpace(string(output)) == "" {
			snap.BranchCount = 0
		}
	}
	cmd = exec.Command("git", "tag", "--list")
	cmd.Dir = repoPath
	if output, err := cmd.CombinedOutput(); err == nil {
		snap.TagCount = len(strings.Split(strings.TrimSpace(string(output)), "\n"))
		if strings.TrimSpace(string(output)) == "" {
			snap.TagCount = 0
		}
	}
	return snap
}

func matchExclude(path string, excludes []string) bool {
	for _, pat := range excludes {
		clean := strings.TrimSpace(pat)
		if clean == "" {
			continue
		}
		if strings.HasPrefix(clean, ".") || strings.HasPrefix(clean, "*") {
			if strings.HasSuffix(path, clean) || strings.HasSuffix(path, strings.TrimPrefix(clean, "*")) {
				return true
			}
		}
		if strings.HasSuffix(clean, "/") {
			if strings.HasPrefix(path, clean) || strings.HasPrefix(path, strings.TrimSuffix(clean, "/")+"/") {
				return true
			}
		}
		if path == clean || strings.HasPrefix(path, clean+"/") {
			return true
		}
	}
	return false
}
