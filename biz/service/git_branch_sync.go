package service

import (
	"fmt"
	"strings"
)

// GetBranchSyncStatus returns ahead/behind counts against upstream
func (s *GitService) GetBranchSyncStatus(path, branch, upstream string) (int, int, error) {
	if upstream == "" {
		return 0, 0, nil
	}
	// git rev-list --left-right --count branch...upstream
	out, err := s.RunCommand(path, "rev-list", "--left-right", "--count", fmt.Sprintf("%s...%s", branch, upstream))
	if err != nil {
		// If upstream doesn't exist locally (maybe not fetched), this fails.
		// Return 0,0 is safer than error for list view
		return 0, 0, nil 
	}
	
	var ahead, behind int
	// Output: "3       5" (3 ahead, 5 behind)
	parts := strings.Fields(out)
	if len(parts) >= 2 {
		fmt.Sscanf(parts[0], "%d", &ahead)
		fmt.Sscanf(parts[1], "%d", &behind)
	}
	return ahead, behind, nil
}

// PushBranch pushes local branch to remote
func (s *GitService) PushBranch(path, remote, branch string) error {
	// git push remote branch
	_, err := s.RunCommand(path, "push", remote, branch)
	return err
}

// PullBranch pulls changes from upstream (rebase)
func (s *GitService) PullBranch(path, remote, branch string) error {
	// git pull --rebase remote branch
	_, err := s.RunCommand(path, "pull", "--rebase", remote, branch)
	return err
}

// FetchAll fetches all remotes
func (s *GitService) FetchAll(path string) error {
	_, err := s.RunCommand(path, "fetch", "--all")
	return err
}
