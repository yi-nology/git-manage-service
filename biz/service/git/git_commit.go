package git

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
)

func (s *GitService) CheckoutBranch(path, branch string) error {
	return s.backend.Checkout(context.Background(), path, branch)
}

func (s *GitService) GetStatus(path string) (string, error) {
	status, err := s.backend.GetStatus(context.Background(), path)
	if err != nil {
		return "", err
	}

	result := ""
	for path, fileStatus := range status.Staged {
		result += fmt.Sprintf("%c%c %s\n", fileStatus.Staging, fileStatus.Worktree, path)
	}
	for path, fileStatus := range status.Unstaged {
		result += fmt.Sprintf("%c%c %s\n", fileStatus.Staging, fileStatus.Worktree, path)
	}
	for _, path := range status.Untracked {
		result += fmt.Sprintf("?? %s\n", path)
	}
	return result, nil
}

func (s *GitService) AddAll(path string) error {
	return s.backend.Add(context.Background(), path, []string{"."})
}

func (s *GitService) AddFiles(path string, files []string) error {
	return s.backend.Add(context.Background(), path, files)
}

func (s *GitService) Commit(path, message, authorName, authorEmail string) error {
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

	return s.backend.CommitWithIdentity(context.Background(), path, authorName, authorEmail, message)
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
