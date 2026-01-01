package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/yi-nology/git-manage-service/biz/model"
)

// ListBranchesWithInfo returns detailed information for all branches
func (s *GitService) ListBranchesWithInfo(path string) ([]model.BranchInfo, error) {
	// Format: refname:short | objectname | authorname | authoremail | committerdate:iso | subject | HEAD | upstream:short
	// HEAD will be "*" if current, " " otherwise (but for-each-ref doesn't show * like branch -v)
	// We can check HEAD ref separately or use %(HEAD) which prints * in for-each-ref since 2.7?
	// Actually %(HEAD) prints * if it matches HEAD.

	args := []string{
		"for-each-ref",
		"--format=%(refname:short)|%(objectname)|%(authorname)|%(authoremail)|%(committerdate:iso)|%(subject)|%(HEAD)|%(upstream:short)",
		"refs/heads",
		"refs/remotes",
	}

	out, err := s.RunCommand(path, args...)
	if err != nil {
		return nil, err
	}

	var branches []model.BranchInfo
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 6 {
			continue
		}

		b := model.BranchInfo{
			Name:        parts[0],
			Hash:        parts[1],
			Author:      parts[2],
			AuthorEmail: parts[3],
			Message:     parts[5],
		}

		if len(parts) >= 8 {
			b.Upstream = parts[7]
		}

		// Parse Date
		// Git ISO format: 2025-01-01 14:00:00 +0800
		t, _ := time.Parse("2006-01-02 15:04:05 -0700", parts[4])
		b.Date = t

		// Check if current
		if len(parts) >= 7 && strings.TrimSpace(parts[6]) == "*" {
			b.IsCurrent = true
		}

		branches = append(branches, b)
	}
	return branches, nil
}

func (s *GitService) CreateBranch(path, name, base string) error {
	args := []string{"branch", name}
	if base != "" {
		args = append(args, base)
	}
	_, err := s.RunCommand(path, args...)
	return err
}

func (s *GitService) DeleteBranch(path, name string, force bool) error {
	flag := "-d"
	if force {
		flag = "-D"
	}
	_, err := s.RunCommand(path, "branch", flag, name)
	return err
}

func (s *GitService) RenameBranch(path, oldName, newName string) error {
	_, err := s.RunCommand(path, "branch", "-m", oldName, newName)
	return err
}

func (s *GitService) GetBranchDescription(path, branch string) (string, error) {
	out, err := s.RunCommand(path, "config", fmt.Sprintf("branch.%s.description", branch))
	if err != nil {
		// Config key might not exist, which is not a fatal error for the flow
		return "", nil
	}
	return out, nil
}

func (s *GitService) SetBranchDescription(path, branch, desc string) error {
	_, err := s.RunCommand(path, "config", fmt.Sprintf("branch.%s.description", branch), desc)
	return err
}

// GetBranchMetrics returns simple metrics: commit count, lines of code (approx)
// This is expensive, use sparingly
func (s *GitService) GetBranchMetrics(path, branch string) (map[string]int, error) {
	// Commit count
	out, err := s.RunCommand(path, "rev-list", "--count", branch)
	if err != nil {
		return nil, err
	}
	count, _ := strconv.Atoi(strings.TrimSpace(out))

	return map[string]int{
		"commit_count": count,
	}, nil
}
