package git

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport"
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

func (s *GitService) openRepo(path string) (*git.Repository, error) {
	log.Printf("[DEBUG] Opening repository at: %s", path)
	r, err := git.PlainOpen(path)
	if err != nil {
		log.Printf("[ERROR] Failed to open repository at %s: %v", path, err)
		return nil, fmt.Errorf("failed to open repository at %s: %v", path, err)
	}
	log.Printf("[DEBUG] Repository opened successfully: %s", path)
	return r, nil
}

func (s *GitService) IsGitRepo(path string) bool {
	_, err := git.PlainOpen(path)
	return err == nil
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

func (s *GitService) FetchWithAuthMethod(path, remoteURL string, auth transport.AuthMethod, progress io.Writer, skipTLS bool, extraArgs ...string) error {
	sdkAuth := s.ConvertTransportAuth(auth)

	_, err := s.backend.Fetch(context.Background(), gitbackend.FetchOptions{
		RepoPath:        path,
		Remote:          remoteURL,
		Tags:            true,
		InsecureSkipTLS: skipTLS,
		Auth:            sdkAuth,
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

func (s *GitService) CloneWithAuthMethod(remoteURL, localPath string, auth transport.AuthMethod, progressChan chan string, skipTLS ...bool) error {
	sdkAuth := s.ConvertTransportAuth(auth)

	var progress io.Writer
	if progressChan != nil {
		progress = &channelWriter{ch: progressChan}
	}

	insecure := len(skipTLS) > 0 && skipTLS[0]

	return s.backend.Clone(context.Background(), gitbackend.CloneOptions{
		URL:             remoteURL,
		Path:            localPath,
		Auth:            sdkAuth,
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
	r, err := s.openRepo(path)
	if err != nil {
		return "", err
	}

	if branch == "" {
		head, err := r.Head()
		if err != nil {
			return "", fmt.Errorf("failed to resolve HEAD: %w", err)
		}
		branch = head.Hash().String()
	}

	commit, err := s.resolveCommit(r, branch)
	if err != nil {
		return "", err
	}

	cIter, err := r.Log(&git.LogOptions{From: commit.Hash})
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	forEachErr := cIter.ForEach(func(c *object.Commit) error {
		line := fmt.Sprintf("%s|%s|%s|%s|%s\n",
			c.Hash.String(),
			c.Author.Name,
			c.Author.Email,
			c.Author.When.Format("2006-01-02 15:04:05 -0700"),
			strings.TrimSpace(strings.Split(c.Message, "\n")[0]),
		)
		sb.WriteString(line)
		return nil
	})
	if forEachErr != nil {
		return "", forEachErr
	}

	return sb.String(), nil
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

func (s *GitService) BlameFile(path, branch, file string) (*git.BlameResult, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}

	commit, err := s.resolveCommit(r, branch)
	if err != nil {
		return nil, err
	}

	return git.Blame(commit, file)
}

func (s *GitService) GetCommit(path, hashStr string) (*object.Commit, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}
	return r.CommitObject(plumbing.NewHash(hashStr))
}

func (s *GitService) resolveCommit(r *git.Repository, rev string) (*object.Commit, error) {
	hash, err := r.ResolveRevision(plumbing.Revision(rev))
	if err != nil {
		if !strings.HasPrefix(rev, "refs/") {
			h, err2 := r.ResolveRevision(plumbing.Revision("refs/heads/" + rev))
			if err2 == nil {
				hash = h
				err = nil
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return r.CommitObject(*hash)
}

func (s *GitService) resolveCommitPair(r *git.Repository, base, target string) (*object.Commit, *object.Commit, error) {
	cBase, err := s.resolveCommit(r, base)
	if err != nil {
		return nil, nil, err
	}
	cTarget, err := s.resolveCommit(r, target)
	if err != nil {
		return nil, nil, err
	}
	return cBase, cTarget, nil
}

func (s *GitService) ResolveRevision(path, rev string) (string, error) {
	return s.backend.RevParse(context.Background(), path, rev)
}

func (s *GitService) GetHeadBranch(path string) (string, error) {
	return s.backend.GetCurrentBranch(context.Background(), path)
}
