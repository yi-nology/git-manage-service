package git

import (
	"context"
	"fmt"

	"github.com/yi-nology/git-platform-sdk/gitbackend"
)

func (s *GitService) GetBranchSyncStatus(path, branch, upstream string) (int, int, error) {
	return s.backend.GetBranchSyncInfo(context.Background(), path, branch, upstream)
}

func (s *GitService) PushBranch(path, remote, branch string, skipTLS ...bool) error {
	return s.PushBranchWithAuth(path, remote, branch, gitbackend.AuthConfig{Type: gitbackend.AuthNone}, skipTLS...)
}

func (s *GitService) PushBranchWithAuth(path, remote, branch string, auth gitbackend.AuthConfig, skipTLS ...bool) error {
	insecure := len(skipTLS) > 0 && skipTLS[0]

	refSpec := fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch)

	_, err := s.backend.Push(context.Background(), gitbackend.PushOptions{
		RepoPath:        path,
		Remote:          remote,
		RefSpecs:        []string{refSpec},
		InsecureSkipTLS: insecure,
		Auth:            auth,
	})
	return err
}

func (s *GitService) PullBranch(path, remote, branch string, skipTLS ...bool) error {
	return s.PullBranchWithAuth(path, remote, branch, gitbackend.AuthConfig{Type: gitbackend.AuthNone}, skipTLS...)
}

func (s *GitService) PullBranchWithAuth(path, remote, branch string, auth gitbackend.AuthConfig, skipTLS ...bool) error {
	insecure := len(skipTLS) > 0 && skipTLS[0]

	_, err := s.backend.Fetch(context.Background(), gitbackend.FetchOptions{
		RepoPath:        path,
		Remote:          remote,
		Branches:        []string{branch},
		InsecureSkipTLS: insecure,
		Auth:            auth,
	})
	return err
}

func (s *GitService) UpdateBranchFastForward(path, remote, branch, remoteBranch string, skipTLS ...bool) error {
	return s.UpdateBranchFastForwardWithAuth(path, remote, branch, remoteBranch, gitbackend.AuthConfig{Type: gitbackend.AuthNone}, skipTLS...)
}

func (s *GitService) UpdateBranchFastForwardWithAuth(path, remote, branch, remoteBranch string, auth gitbackend.AuthConfig, skipTLS ...bool) error {
	insecure := len(skipTLS) > 0 && skipTLS[0]

	_, err := s.backend.Fetch(context.Background(), gitbackend.FetchOptions{
		RepoPath:        path,
		Remote:          remote,
		Branches:        []string{remoteBranch},
		InsecureSkipTLS: insecure,
		Auth:            auth,
	})
	return err
}

func (s *GitService) FetchAll(path string, skipTLS ...bool) error {
	insecure := len(skipTLS) > 0 && skipTLS[0]
	auth := gitbackend.AuthConfig{Type: gitbackend.AuthNone}
	_ = insecure // TODO: pass insecure to auth config
	return s.backend.FetchAll(context.Background(), path, auth)
}
