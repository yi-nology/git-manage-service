package git

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/yi-nology/git-platform-sdk/gitbackend"
)

type DiffStat struct {
	FilesChanged int
	Insertions   int
	Deletions    int
}

type FileDiffStatus struct {
	Path   string `json:"path"`
	Status string `json:"status"` // A, M, D, R, C, U, etc.
}

// GetDiffStat returns the shortstat of diff between base and target
func (s *GitService) GetDiffStat(path, base, target string) (*DiffStat, error) {
	diff, err := s.backend.Diff(context.Background(), path, gitbackend.DiffOptions{
		From: base,
		To:   target,
	})
	if err != nil {
		return nil, err
	}

	ds := &DiffStat{}
	for _, line := range strings.Split(diff, "\n") {
		if len(line) == 0 {
			continue
		}
		if line[0] == '+' && !strings.HasPrefix(line, "+++") {
			ds.Insertions++
		} else if line[0] == '-' && !strings.HasPrefix(line, "---") {
			ds.Deletions++
		}
	}

	// Count files from diff names
	names, err := s.backend.DiffNames(context.Background(), path, base, target)
	if err == nil {
		ds.FilesChanged = len(names)
	}

	return ds, nil
}

// GetDiffFiles returns list of changed files with status
func (s *GitService) GetDiffFiles(path, base, target string) ([]FileDiffStatus, error) {
	names, err := s.backend.DiffNames(context.Background(), path, base, target)
	if err != nil {
		return nil, err
	}

	deleted, err := s.backend.DeletedFiles(context.Background(), path, base, target)
	if err != nil {
		return nil, err
	}

	deletedMap := make(map[string]bool)
	for _, f := range deleted {
		deletedMap[f] = true
	}

	var files []FileDiffStatus
	for _, name := range names {
		status := "M"
		if deletedMap[name] {
			status = "D"
		}
		files = append(files, FileDiffStatus{
			Path:   name,
			Status: status,
		})
	}
	return files, nil
}

// GetRawDiff returns the full diff content
func (s *GitService) GetRawDiff(path, base, target, file string) (string, error) {
	opts := gitbackend.DiffOptions{
		From: base,
		To:   target,
	}
	if file != "" {
		opts.Paths = []string{file}
	}

	return s.backend.Diff(context.Background(), path, opts)
}

type MergeResult struct {
	Success   bool     `json:"success"`
	Conflicts []string `json:"conflicts"`
	Output    string   `json:"output"`
	MergeID   string   `json:"merge_id"`
}

// MergeDryRun checks for conflicts without committing
func (s *GitService) MergeDryRun(path, source, target string) (*MergeResult, error) {
	// Get merge base
	baseHash, err := s.backend.MergeBase(context.Background(), path, source, target)
	if err != nil || baseHash == "" {
		return nil, fmt.Errorf("no merge base found")
	}

	// Get files changed in source
	sourceFiles, err := s.backend.DiffNames(context.Background(), path, baseHash, source)
	if err != nil {
		return nil, err
	}

	// Get files changed in target
	targetFiles, err := s.backend.DiffNames(context.Background(), path, baseHash, target)
	if err != nil {
		return nil, err
	}

	// Check overlap
	sourceMap := make(map[string]bool)
	for _, f := range sourceFiles {
		sourceMap[f] = true
	}

	var conflicts []string
	for _, f := range targetFiles {
		if sourceMap[f] {
			conflicts = append(conflicts, f)
		}
	}

	return &MergeResult{
		Success:   len(conflicts) == 0,
		Conflicts: conflicts,
	}, nil
}

// MergeConflictError 表示真实合并阶段发生冲突（MergeDryRun 的启发式
// 可能漏报）。Files 是 abort 前收集到的未合并文件列表。
type MergeConflictError struct {
	Files []string
}

func (e *MergeConflictError) Error() string {
	return fmt.Sprintf("merge conflicts in %d file(s)", len(e.Files))
}

// unmergedFiles 返回当前处于未合并（conflicted）状态的文件；必须在
// merge --abort 之前调用。
func (s *GitService) unmergedFiles(path string) []string {
	out, err := s.RunCommand(path, "diff", "--name-only", "--diff-filter=U")
	if err != nil || strings.TrimSpace(out) == "" {
		return nil
	}
	return strings.Split(strings.TrimSpace(out), "\n")
}

// Merge performs the actual merge
func (s *GitService) Merge(path, source, target, message string, noFF, squash bool) error {
	// Checkout target
	if err := s.CheckoutBranch(path, target); err != nil {
		return fmt.Errorf("checkout target failed: %v", err)
	}

	// Merge source
	opts := gitbackend.MergeOptions{
		Message: message,
		Squash:  squash,
	}
	if noFF {
		opts.FFOnly = false
	}

	err := s.backend.Merge(context.Background(), path, source, opts)
	if err != nil {
		var conflict *MergeConflictError
		if errors.Is(err, gitbackend.ErrMergeConflict) {
			conflict = &MergeConflictError{Files: s.unmergedFiles(path)}
		}
		// Abort merge on failure
		_, _ = s.RunCommand(path, "merge", "--abort")
		if conflict != nil {
			return conflict
		}
		return fmt.Errorf("merge failed (aborted): %v", err)
	}

	// If squash, commit separately
	if squash {
		commitMsg := message
		if commitMsg == "" {
			commitMsg = fmt.Sprintf("Squash merge %s into %s", source, target)
		}
		err = s.backend.CommitWithIdentity(context.Background(), path, "Git Manage Service", "git-manage@example.com", commitMsg)
		if err != nil {
			return fmt.Errorf("squash commit failed: %v", err)
		}
	}

	return nil
}

// GetPatch generates a patch file content
func (s *GitService) GetPatch(path, base, target string) (string, error) {
	return s.backend.Diff(context.Background(), path, gitbackend.DiffOptions{
		From: base,
		To:   target,
	})
}
