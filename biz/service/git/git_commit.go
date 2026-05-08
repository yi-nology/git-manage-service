package git

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
)

func (s *GitService) CheckoutBranch(path, branch string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	refName := plumbing.ReferenceName("refs/heads/" + branch)
	_, err = r.Reference(refName, true)
	if err != nil {
		remoteRefName := plumbing.ReferenceName("refs/remotes/origin/" + branch)
		remoteRef, err := r.Reference(remoteRefName, true)
		if err != nil {
			return fmt.Errorf("branch %s not found (local or remote)", branch)
		}

		return w.Checkout(&git.CheckoutOptions{
			Hash:   remoteRef.Hash(),
			Branch: refName,
			Create: true,
			Force:  true,
		})
	}

	return w.Checkout(&git.CheckoutOptions{
		Branch: refName,
		Force:  true,
	})
}

func (s *GitService) GetStatus(path string) (string, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return "", err
	}
	w, err := r.Worktree()
	if err != nil {
		return "", err
	}
	status, err := w.Status()
	if err != nil {
		return "", err
	}
	return status.String(), nil
}

func (s *GitService) AddAll(path string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	_, err = w.Add(".")
	return err
}

func (s *GitService) AddFiles(path string, files []string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	for _, f := range files {
		if _, err := w.Add(f); err != nil {
			return fmt.Errorf("failed to add %s: %w", f, err)
		}
	}
	return nil
}

func (s *GitService) Commit(path, message, authorName, authorEmail string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	if authorName == "" || authorEmail == "" {
		if r, _ := db.NewRepoDAO().FindByPath(path); r != nil {
			authorSvc := NewAuthorService()
			if name, email, _ := authorSvc.GetEffectiveAuthor(r.ID); name != "" && email != "" {
				if authorName == "" {
					authorName = name
				}
				if authorEmail == "" {
					authorEmail = email
				}
			}
		}
	}
	if authorName == "" {
		authorName = "Git Manage Service"
	}
	if authorEmail == "" {
		authorEmail = "git-manage@example.com"
	}

	_, err = w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  time.Now(),
		},
	})
	return err
}

func (s *GitService) Reset(path string) error {
	r, err := s.openRepo(path)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}
	return w.Reset(&git.ResetOptions{Mode: git.MixedReset})
}

func (s *GitService) GetLogIterator(path, branch string) (object.CommitIter, error) {
	r, err := s.openRepo(path)
	if err != nil {
		return nil, err
	}
	hash, err := r.ResolveRevision(plumbing.Revision(branch))
	if err != nil {
		return nil, err
	}
	return r.Log(&git.LogOptions{From: *hash})
}

func (s *GitService) GetLogStats(path, branch string) (string, error) {
	return s.RunCommand(path, "log", "--numstat", "--no-merges", "--pretty=format:COMMIT|%H|%aN|%aE|%at", branch)
}

func (s *GitService) GetLogStatsStream(path, branch string) (io.ReadCloser, error) {
	args := []string{"log", "--numstat", "--no-merges", "--pretty=format:COMMIT|%H|%aN|%aE|%at"}
	if branch != "" {
		args = append(args, branch)
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = path
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "LC_ALL=C")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &cmdStream{
		cmd:    cmd,
		stdout: stdout,
	}, nil
}

type cmdStream struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
}

func (c *cmdStream) Read(p []byte) (n int, err error) {
	return c.stdout.Read(p)
}

func (c *cmdStream) Close() error {
	_ = c.stdout.Close()
	return c.cmd.Wait()
}
