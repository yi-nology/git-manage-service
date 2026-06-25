package git

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/yi-nology/git-platform-sdk/gitbackend"
)

func (s *GitService) GetBranchSyncStatus(path, branch, upstream string) (int, int, error) {
	if upstream == "" {
		return 0, 0, nil
	}

	baseHash, err := s.backend.MergeBase(context.Background(), path, branch, upstream)
	if err != nil {
		return 0, 0, nil
	}

	commitsAhead, err := s.backend.GetCommitsBetween(context.Background(), path, baseHash, branch)
	if err != nil {
		return 0, 0, nil
	}

	commitsBehind, err := s.backend.GetCommitsBetween(context.Background(), path, baseHash, upstream)
	if err != nil {
		return 0, 0, nil
	}

	return len(commitsAhead), len(commitsBehind), nil
}

func (s *GitService) PushBranch(path, remote, branch string, skipTLS ...bool) error {
	return s.PushBranchWithAuth(path, remote, branch, nil, skipTLS...)
}

func (s *GitService) PushBranchWithAuth(path, remote, branch string, auth transport.AuthMethod, skipTLS ...bool) error {
	insecure := len(skipTLS) > 0 && skipTLS[0]

	sdkAuth := s.ConvertTransportAuth(auth)

	refSpec := fmt.Sprintf("refs/heads/%s:refs/heads/%s", branch, branch)

	_, err := s.backend.Push(context.Background(), gitbackend.PushOptions{
		RepoPath:        path,
		Remote:          remote,
		RefSpecs:        []string{refSpec},
		InsecureSkipTLS: insecure,
		Auth:            sdkAuth,
	})
	return err
}

func (s *GitService) pushBranchCLI(path, remote, branch string) error {
	cmd := exec.Command("git", "-C", path, "-c", "http.sslVerify=false", "push", remote, branch)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git push %s %s: %s: %w", remote, branch, string(output), err)
	}
	return nil
}

func (s *GitService) PullBranch(path, remote, branch string, skipTLS ...bool) error {
	return s.PullBranchWithAuth(path, remote, branch, nil, skipTLS...)
}

func (s *GitService) PullBranchWithAuth(path, remote, branch string, auth transport.AuthMethod, skipTLS ...bool) error {
	insecure := len(skipTLS) > 0 && skipTLS[0]

	sdkAuth := s.ConvertTransportAuth(auth)

	_, err := s.backend.Fetch(context.Background(), gitbackend.FetchOptions{
		RepoPath:        path,
		Remote:          remote,
		Branches:        []string{branch},
		InsecureSkipTLS: insecure,
		Auth:            sdkAuth,
	})
	return err
}

func (s *GitService) UpdateBranchFastForward(path, remote, branch, remoteBranch string, skipTLS ...bool) error {
	return s.UpdateBranchFastForwardWithAuth(path, remote, branch, remoteBranch, nil, skipTLS...)
}

func (s *GitService) UpdateBranchFastForwardWithAuth(path, remote, branch, remoteBranch string, auth transport.AuthMethod, skipTLS ...bool) error {
	insecure := len(skipTLS) > 0 && skipTLS[0]

	sdkAuth := s.ConvertTransportAuth(auth)

	refSpec := fmt.Sprintf("refs/heads/%s:refs/heads/%s", remoteBranch, branch)

	_, err := s.backend.Fetch(context.Background(), gitbackend.FetchOptions{
		RepoPath:        path,
		Remote:          remote,
		InsecureSkipTLS: insecure,
		Auth:            sdkAuth,
	})
	_ = refSpec
	return err
}

func (s *GitService) FetchAll(path string, skipTLS ...bool) error {
	insecure := len(skipTLS) > 0 && skipTLS[0]

	_, err := s.backend.Fetch(context.Background(), gitbackend.FetchOptions{
		RepoPath:        path,
		Remote:          "origin",
		Tags:            true,
		InsecureSkipTLS: insecure,
	})
	return err
}
