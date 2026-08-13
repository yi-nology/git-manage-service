package git

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/model/api"
	workspaceModel "github.com/yi-nology/git-manage-service/biz/model/workspace"
)

func (s *GitService) GetWorkspaceStatus(repoPath string) (*workspaceModel.GetWorkspaceStatusResp, error) {
	// 使用 SDK 获取状态
	status, err := s.backend.GetStatus(context.Background(), repoPath)
	if err != nil {
		return nil, fmt.Errorf("get status: %w", err)
	}

	branch, _ := s.backend.GetCurrentBranch(context.Background(), repoPath)

	result := &workspaceModel.GetWorkspaceStatusResp{
		Branch:  branch,
		IsClean: status.IsClean,
	}

	isMerging, _ := s.isMerging(repoPath)
	isRebasing, _ := s.isRebasing(repoPath)
	result.IsMerging = isMerging
	result.IsRebasing = isRebasing

	result.Ahead = int32(status.Ahead)
	result.Behind = int32(status.Behind)

	// 处理暂存文件
	for _, fs := range status.Staged {
		item := &workspaceModel.FileStatus{
			Path:   fs.Path,
			Status: s.readableStatus(string(fs.Staging)),
		}
		result.Staged = append(result.Staged, item)
	}

	// 处理未暂存文件
	for _, fs := range status.Unstaged {
		item := &workspaceModel.FileStatus{
			Path:   fs.Path,
			Status: s.readableStatus(string(fs.Worktree)),
		}
		result.Unstaged = append(result.Unstaged, item)
	}

	// 处理未跟踪文件
	for _, path := range status.Untracked {
		item := &workspaceModel.FileStatus{
			Path:   path,
			Status: "untracked",
		}
		result.Untracked = append(result.Untracked, item)
	}

	if result.Staged == nil {
		result.Staged = []*workspaceModel.FileStatus{}
	}
	if result.Unstaged == nil {
		result.Unstaged = []*workspaceModel.FileStatus{}
	}
	if result.Untracked == nil {
		result.Untracked = []*workspaceModel.FileStatus{}
	}
	if result.Conflicted == nil {
		result.Conflicted = []*workspaceModel.FileStatus{}
	}

	return result, nil
}

func (s *GitService) GetWorkspaceDiff(repoPath, file string, stagedOnly bool) (*api.WorkspaceDiff, error) {
	// Compute the diff of uncommitted changes against HEAD (or just the staged
	// subset). The previous implementation stuffed the entire `git diff` output
	// into a single file entry and never populated per-file counts, so
	// totalAdditions was always 0 and "no changes" still reported one file.
	// Use --numstat for accurate per-file addition/deletion counts.
	rangeArgs := []string{}
	if stagedOnly {
		rangeArgs = append(rangeArgs, "--cached")
	} else {
		rangeArgs = append(rangeArgs, "HEAD")
	}
	if file != "" {
		rangeArgs = append(rangeArgs, "--", file)
	}

	numstatOut, err := s.RunCommand(repoPath, append([]string{"diff", "--numstat"}, rangeArgs...)...)
	if err != nil {
		// No HEAD yet (empty repo) or other git error: treat as no diff.
		return &api.WorkspaceDiff{Files: []api.WorkspaceDiffFile{}}, nil
	}

	patches := s.workspaceDiffPatches(repoPath, rangeArgs)

	result := &api.WorkspaceDiff{Files: []api.WorkspaceDiffFile{}}
	for _, line := range strings.Split(numstatOut, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) < 3 {
			continue
		}
		df := api.WorkspaceDiffFile{File: parts[2]}
		if parts[0] == "-" || parts[1] == "-" {
			df.IsBinary = true
		} else {
			df.Additions, _ = strconv.Atoi(parts[0])
			df.Deletions, _ = strconv.Atoi(parts[1])
		}
		df.Diff = patches[df.File]
		result.TotalAdditions += df.Additions
		result.TotalDeletions += df.Deletions
		result.Files = append(result.Files, df)
	}

	return result, nil
}

// workspaceDiffPatches returns the raw diff split into per-file patches keyed
// by file path. rangeArgs are the diff range/path arguments (without the
// leading "diff" subcommand and without "--numstat").
func (s *GitService) workspaceDiffPatches(repoPath string, rangeArgs []string) map[string]string {
	patches := make(map[string]string)
	out, err := s.RunCommand(repoPath, append([]string{"diff"}, rangeArgs...)...)
	if err != nil {
		return patches
	}

	var curFile string
	var cur strings.Builder
	flush := func() {
		if curFile != "" {
			patches[curFile] = cur.String()
		}
		cur.Reset()
		curFile = ""
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "diff --git a/") {
			flush()
			curFile = extractDiffPath(line)
		}
		if curFile != "" {
			cur.WriteString(line)
			cur.WriteString("\n")
		}
	}
	flush()
	return patches
}

// extractDiffPath parses the file path from a "diff --git a/<p> b/<p>" header.
func extractDiffPath(header string) string {
	rest := strings.TrimPrefix(header, "diff --git a/")
	if idx := strings.Index(rest, " b/"); idx >= 0 {
		return rest[idx+3:]
	}
	return rest
}

func (s *GitService) StageFiles(repoPath string, files []string, stageAll bool) error {
	if stageAll {
		return s.backend.Add(context.Background(), repoPath, []string{"."})
	}

	return s.backend.Add(context.Background(), repoPath, files)
}

func (s *GitService) UnstageFiles(repoPath string, files []string, unstageAll bool) error {
	// SDK 没有直接的 unstage 方法，使用 RunCommand
	if unstageAll {
		_, err := s.RunCommand(repoPath, "reset", "HEAD")
		return err
	}

	args := []string{"reset", "HEAD", "--"}
	args = append(args, files...)
	_, err := s.RunCommand(repoPath, args...)
	return err
}

func (s *GitService) CommitChanges(repoPath string, files []string, stageAll bool, message, authorName, authorEmail string, push bool, pushRemote string) (*api.CommitResult, error) {
	if stageAll || len(files) > 0 {
		if err := s.StageFiles(repoPath, files, stageAll); err != nil {
			return nil, fmt.Errorf("stage: %w", err)
		}
	}

	if err := s.Commit(repoPath, message, authorName, authorEmail); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	headHash, _ := s.ResolveRevision(repoPath, "HEAD")
	result := &api.CommitResult{CommitHash: headHash}

	if push {
		remote := pushRemote
		if remote == "" {
			remote = "origin"
		}
		branch, _ := s.GetHeadBranch(repoPath)
		pushErr := s.Push(repoPath, remote, "", branch, nil, nil)
		result.Pushed = pushErr == nil
	}

	return result, nil
}

func (s *GitService) readableStatus(code string) string {
	switch code {
	case "A":
		return "added"
	case "M":
		return "modified"
	case "D":
		return "deleted"
	case "R":
		return "renamed"
	case "C":
		return "copied"
	case "?":
		return "untracked"
	case "U":
		return "conflicted"
	default:
		return code
	}
}

func (s *GitService) isMerging(repoPath string) (bool, error) {
	_, err := os.Stat(filepath.Join(repoPath, ".git", "MERGE_HEAD"))
	return err == nil, nil
}

func (s *GitService) isRebasing(repoPath string) (bool, error) {
	_, err1 := os.Stat(filepath.Join(repoPath, ".git", "rebase-merge"))
	_, err2 := os.Stat(filepath.Join(repoPath, ".git", "rebase-apply"))
	return err1 == nil || err2 == nil, nil
}

func (s *GitService) RemoveTracking(repoPath string, files []string) error {
	for _, f := range files {
		_, err := s.RunCommand(repoPath, "rm", "--cached", f)
		if err != nil {
			return fmt.Errorf("rm --cached %s: %w", f, err)
		}
	}
	return nil
}

func (s *GitService) AddToGitignore(repoPath string, patterns []string) error {
	gitignorePath := filepath.Join(repoPath, ".gitignore")
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open .gitignore: %w", err)
	}
	defer f.Close()

	for _, p := range patterns {
		if _, err := f.WriteString(p + "\n"); err != nil {
			return fmt.Errorf("write .gitignore: %w", err)
		}
	}
	return nil
}

func (s *GitService) GetWorkspaceDiffRaw(repoPath string, file string) (string, error) {
	args := []string{"diff", "--stat", "--patch"}
	if file != "" {
		args = append(args, "--", file)
	}
	out, err := s.RunCommand(repoPath, args...)
	if err != nil {
		return "", fmt.Errorf("git diff: %w", err)
	}
	return out, nil
}
