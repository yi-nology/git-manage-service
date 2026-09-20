package git

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/domain"
)

// ListBranchesWithInfo returns detailed information for all branches
func (s *GitService) ListBranchesWithInfo(path string) ([]domain.BranchInfo, error) {
	details, err := s.backend.ListBranches(context.Background(), path)
	if err != nil {
		return nil, err
	}

	var branches []domain.BranchInfo
	for _, d := range details {
		b := domain.BranchInfo{
			Name:        d.Name,
			Hash:        d.Hash,
			IsCurrent:   d.IsCurrent,
			Author:      d.Author,
			AuthorEmail: d.Email,
			Message:     d.Message,
			Upstream:    d.Upstream,
		}
		if d.IsRemote {
			b.Type = "remote"
		} else {
			b.Type = "local"
		}
		if d.Date != "" {
			if t, err := time.Parse(time.RFC3339, d.Date); err == nil {
				b.Date = t
			}
		}
		branches = append(branches, b)
	}
	return branches, nil
}

func (s *GitService) CreateBranch(path, name, base string) error {
	// The SDK gitbackend passes the ref straight to `git branch <name> <ref>`,
	// which fails with "not a valid object name: ''" when base is empty. Restore
	// the historical go-git behavior of defaulting to HEAD when no base is given.
	if base == "" {
		base = "HEAD"
	}
	return s.backend.CreateBranch(context.Background(), path, name, base)
}

func (s *GitService) DeleteBranch(path, name string, force bool) error {
	return s.backend.DeleteBranch(context.Background(), path, name)
}

func (s *GitService) RenameBranch(path, oldName, newName string) error {
	return s.backend.RenameBranch(context.Background(), path, oldName, newName)
}

// validateBranchRefName 拒绝会破坏 git config 位置参数解析的分支名。
func validateBranchRefName(branch string) error {
	if branch == "" || strings.HasPrefix(branch, "-") {
		return fmt.Errorf("invalid branch name: %q", branch)
	}
	return nil
}

func (s *GitService) SetBranchDescription(path, branch, desc string) error {
	if err := validateBranchRefName(branch); err != nil {
		return err
	}
	// SDK v0.62 起 SetConfig 只接受 2/3 段 key，分支名含点号（如 release/1.2.3）
	// 会构成 4+ 段被拒；git CLI 本身支持多段 subsection，走 RunRaw。
	key := "branch." + branch + ".description"
	_, _, err := s.backend.RunRaw(context.Background(), path, []string{"config", "--local", "--", key, desc})
	return err
}

func (s *GitService) GetBranchDescription(path, branch string) (string, error) {
	if err := validateBranchRefName(branch); err != nil {
		return "", err
	}
	key := "branch." + branch + ".description"
	// --default "" 让「未设置」不再是 exit 1，从而能与真实读取错误区分
	val, _, err := s.backend.RunRaw(context.Background(), path, []string{"config", "--local", "--default", "", "--get", key})
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(val), nil
}

// GetBranchMetrics returns simple metrics: commit count
func (s *GitService) GetBranchMetrics(path, branch string) (map[string]int, error) {
	commits, err := s.backend.GetCommitsBetween(context.Background(), path, "", branch)
	if err != nil {
		return nil, err
	}

	return map[string]int{
		"commit_count": len(commits),
	}, nil
}
