package git

import (
	"context"
	"fmt"
	"io"
	"log"
	"os/exec"

	"github.com/yi-nology/git-platform-sdk/gitbackend"
)

type channelWriter struct {
	ch chan string
}

func (w *channelWriter) Write(p []byte) (n int, err error) {
	if w.ch != nil {
		w.ch <- string(p)
	}
	return len(p), nil
}

func (s *GitService) IsGitRepo(path string) bool {
	// An empty path would make `git -C "" ...` resolve to the process working
	// directory, falsely treating it (e.g. the server's own repo) as the target.
	if path == "" {
		return false
	}
	cmd := exec.Command("git", "-C", path, "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func (s *GitService) Fetch(path, remote string, progress io.Writer, skipTLS ...bool) error {
	insecure := len(skipTLS) > 0 && skipTLS[0]

	_, err := s.backend.Fetch(context.Background(), gitbackend.FetchOptions{
		RepoPath:        path,
		Remote:          remote,
		Tags:            true,
		InsecureSkipTLS: insecure,
		Progress:        progress,
	})
	return err
}

func (s *GitService) FetchWithAuth(path, remoteURL, authType, authKey, authSecret string, progress io.Writer, skipTLS bool, extraArgs ...string) error {
	auth := s.buildSDKAuth(authType, authKey, authSecret)

	_, err := s.backend.Fetch(context.Background(), gitbackend.FetchOptions{
		RepoPath:        path,
		Remote:          remoteURL,
		Tags:            true,
		InsecureSkipTLS: skipTLS,
		Auth:            auth,
		Progress:        progress,
	})
	return err
}

func (s *GitService) FetchWithAuthMethod(path, remoteURL string, auth gitbackend.AuthConfig, progress io.Writer, skipTLS bool, extraArgs ...string) error {
	_, err := s.backend.Fetch(context.Background(), gitbackend.FetchOptions{
		RepoPath:        path,
		Remote:          remoteURL,
		Tags:            true,
		InsecureSkipTLS: skipTLS,
		Auth:            auth,
		Progress:        progress,
	})
	return err
}

func (s *GitService) Clone(remoteURL, localPath, authType, authKey, authSecret string, skipTLS ...bool) error {
	return s.CloneWithProgress(remoteURL, localPath, authType, authKey, authSecret, nil, skipTLS...)
}

func (s *GitService) CloneWithProgress(remoteURL, localPath, authType, authKey, authSecret string, progressChan chan string, skipTLS ...bool) error {
	auth := s.buildSDKAuth(authType, authKey, authSecret)

	var progress io.Writer
	if progressChan != nil {
		progress = &channelWriter{ch: progressChan}
	}

	insecure := len(skipTLS) > 0 && skipTLS[0]

	return s.backend.Clone(context.Background(), gitbackend.CloneOptions{
		URL:             remoteURL,
		Path:            localPath,
		Auth:            auth,
		Progress:        progress,
		InsecureSkipTLS: insecure,
	})
}

func (s *GitService) CloneWithAuthMethod(remoteURL, localPath string, auth gitbackend.AuthConfig, progressChan chan string, skipTLS ...bool) error {
	var progress io.Writer
	if progressChan != nil {
		progress = &channelWriter{ch: progressChan}
	}

	insecure := len(skipTLS) > 0 && skipTLS[0]

	return s.backend.Clone(context.Background(), gitbackend.CloneOptions{
		URL:             remoteURL,
		Path:            localPath,
		Auth:            auth,
		Progress:        progress,
		InsecureSkipTLS: insecure,
	})
}

func (s *GitService) GetCommitHash(path, remote, branch string) (string, error) {
	refName := fmt.Sprintf("refs/remotes/%s/%s", remote, branch)
	hash, err := s.backend.RevParse(context.Background(), path, refName)
	if err != nil {
		return "", fmt.Errorf("remote branch %s/%s not found: %v", remote, branch, err)
	}
	return hash, nil
}

func (s *GitService) IsAncestor(path, ancestor, descendant string) (bool, error) {
	return s.backend.IsAncestor(context.Background(), path, ancestor, descendant)
}

func (s *GitService) GetBranches(path string) ([]string, error) {
	return s.backend.ListLocalBranches(context.Background(), path)
}

func (s *GitService) GetCommits(path, branch, since, until string) (string, error) {
	log.Printf("[DEBUG] GetCommits path=%s branch=%s since=%s until=%s", path, branch, since, until)

	args := []string{"-C", path, "log", "--pretty=format:%H|%aN|%aE|%ad|%s", "--date=format:%Y-%m-%d %H:%M:%S %z"}
	if branch != "" {
		args = append(args, branch)
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git log failed: %v", err)
	}
	return string(output), nil
}

func (s *GitService) GetRepoFiles(path, branch string) ([]string, error) {
	entries, err := s.backend.GetTree(context.Background(), path, branch, "", true)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, e := range entries {
		if e.Type == "file" {
			files = append(files, e.Path)
		}
	}
	return files, nil
}

func (s *GitService) ResolveRevision(path, rev string) (string, error) {
	return s.backend.RevParse(context.Background(), path, rev)
}

func (s *GitService) GetHeadBranch(path string) (string, error) {
	return s.backend.GetCurrentBranch(context.Background(), path)
}
