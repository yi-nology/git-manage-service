package service

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/config"
)

type GitService struct{}

func NewGitService() *GitService {
	return &GitService{}
}

func (s *GitService) RunCommand(dir string, args ...string) (string, error) {
	if config.DebugMode {
		log.Printf("[DEBUG] Executing in %s: git %s", dir, strings.Join(args, " "))
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("git command failed: %s, output: %s", err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func (s *GitService) IsGitRepo(path string) bool {
	_, err := s.RunCommand(path, "rev-parse", "--is-inside-work-tree")
	return err == nil
}

func (s *GitService) Fetch(path, remote string) error {
	_, err := s.RunCommand(path, "fetch", remote)
	return err
}

func (s *GitService) GetCommitHash(path, remote, branch string) (string, error) {
	// ref: refs/remotes/<remote>/<branch>
	ref := fmt.Sprintf("refs/remotes/%s/%s", remote, branch)
	return s.RunCommand(path, "rev-parse", ref)
}

// IsAncestor checks if ancestor is an ancestor of descendant (fast-forward possible)
func (s *GitService) IsAncestor(path, ancestor, descendant string) (bool, error) {
	cmd := exec.Command("git", "merge-base", "--is-ancestor", ancestor, descendant)
	cmd.Dir = path
	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func (s *GitService) Push(path, targetRemote, sourceHash, targetBranch string, options []string) error {
	// git push [options] <remote> <source_hash>:refs/heads/<target_branch>
	args := []string{"push"}
	if len(options) > 0 {
		args = append(args, options...)
	}
	refSpec := fmt.Sprintf("%s:refs/heads/%s", sourceHash, targetBranch)
	args = append(args, targetRemote, refSpec)
	_, err := s.RunCommand(path, args...)
	return err
}

func (s *GitService) GetRemotes(path string) ([]string, error) {
	out, err := s.RunCommand(path, "remote")
	if err != nil {
		return nil, err
	}
	return strings.Split(out, "\n"), nil
}

// GetBranches returns all local and remote branches
func (s *GitService) GetBranches(path string) ([]string, error) {
	// git branch -a --format="%(refname:short)"
	out, err := s.RunCommand(path, "branch", "-a", "--format=%(refname:short)")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(out, "\n")
	var branches []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.Contains(line, "HEAD") {
			branches = append(branches, line)
		}
	}
	return branches, nil
}

// GetCommits returns commits for a specific branch
func (s *GitService) GetCommits(path, branch, since, until string) (string, error) {
	// git log --pretty=format:"%H|%an|%ae|%ad|%s" --date=iso
	args := []string{"log", "--pretty=format:%H|%an|%ae|%ad|%s", "--date=iso", branch}
	if since != "" {
		args = append(args, "--since="+since)
	}
	if until != "" {
		args = append(args, "--until="+until)
	}
	return s.RunCommand(path, args...)
}

// GetRepoFiles returns all files in the current HEAD of the branch
func (s *GitService) GetRepoFiles(path, branch string) ([]string, error) {
	// git ls-tree -r --name-only <branch>
	out, err := s.RunCommand(path, "ls-tree", "-r", "--name-only", branch)
	if err != nil {
		return nil, err
	}
	return strings.Split(out, "\n"), nil
}

// BlameFile returns blame information for a file
func (s *GitService) BlameFile(path, branch, file string) (string, error) {
	// git blame --line-porcelain -w <branch> -- <file>
	// -w ignores whitespace
	return s.RunCommand(path, "blame", "--line-porcelain", "-w", branch, "--", file)
}

// TestRemoteConnection checks if the remote is accessible
func (s *GitService) TestRemoteConnection(url string) error {
	// git ls-remote <url>
	cmd := exec.Command("git", "ls-remote", url)
	// We might need to set timeouts or handle auth prompts (which will fail in non-interactive mode)
	// If it prompts for password, it will hang or fail.
	// Setting GIT_TERMINAL_PROMPT=0 prevents hanging on password prompt
	cmd.Env = append(cmd.Env, "GIT_TERMINAL_PROMPT=0")

	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("connection failed: %v, output: %s", err, string(out))
	}
	return nil
}
