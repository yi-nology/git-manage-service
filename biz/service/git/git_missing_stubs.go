package git

import (
	"fmt"
	"os/exec"
	"strings"
)

type SearchCommitsOptions struct {
	Ref      string
	Author   string
	Keyword  string
	Since    string
	Until    string
	Path     string
	Page     int
	PageSize int
}

type CommitDetailResult struct {
	Hash           string
	ShortHash      string
	Message        string
	AuthorName     string
	AuthorEmail    string
	AuthorDate     string
	CommitterName  string
	CommitterEmail string
	CommitterDate  string
	ParentHashes   []string
	FilesChanged   int
	Additions      int
	Deletions      int
}

type FileChangeResult struct {
	Path      string
	Status    string
	Additions int
	Deletions int
	OldPath   string
}

func runGitCmd(repoPath string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func (s *GitService) PullWithResolve(repoPath, remote, branch string, fetchOnly bool) (map[string]interface{}, error) {
	if fetchOnly {
		out, err := runGitCmd(repoPath, "fetch", remote)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", strings.TrimSpace(out), err)
		}
		return map[string]interface{}{"fetched": true}, nil
	}
	out, err := runGitCmd(repoPath, "pull", remote, branch)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", strings.TrimSpace(out), err)
	}
	return map[string]interface{}{"pulled": true}, nil
}

func (s *GitService) GetConflictDetail(repoPath, file string) (map[string]interface{}, error) {
	return map[string]interface{}{
		"file":    file,
		"content": "",
	}, nil
}

func (s *GitService) MarkConflictResolved(repoPath, file, resolvedContent string, stage bool) error {
	return nil
}

func (s *GitService) SearchCommits(repoPath string, opts SearchCommitsOptions) ([]*CommitDetailResult, int, error) {
	return []*CommitDetailResult{}, 0, nil
}

func (s *GitService) GetCommitDetail(repoPath, hash string) (*CommitDetailResult, []*FileChangeResult, error) {
	return &CommitDetailResult{}, []*FileChangeResult{}, nil
}

func (s *GitService) GetCommitDiff(repoPath, hash, file string) (string, error) {
	args := []string{"diff", hash + "^.." + hash}
	if file != "" {
		args = append(args, "--", file)
	}
	out, err := runGitCmd(repoPath, args...)
	if err != nil {
		return "", err
	}
	return out, nil
}
