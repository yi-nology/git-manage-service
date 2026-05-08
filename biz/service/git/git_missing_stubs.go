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
	if remote == "" {
		remote = "origin"
	}

	if fetchOnly {
		args := []string{"fetch", remote}
		if branch != "" {
			args = append(args, branch)
		}
		out, err := runGitCmd(repoPath, args...)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", strings.TrimSpace(out), err)
		}
		return map[string]interface{}{"fetched": true}, nil
	}

	args := []string{"pull", remote}
	if branch != "" {
		args = append(args, branch)
	}
	out, err := runGitCmd(repoPath, args...)
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
	args := []string{"log", "--no-merges", "--pretty=format:%H|%h|%aN|%aE|%aI|%cN|%cE|%cI|%P|%s", "--numstat"}
	if opts.Ref != "" {
		args = append(args, opts.Ref)
	}
	if opts.Author != "" {
		args = append(args, "--author="+opts.Author)
	}
	if opts.Keyword != "" {
		args = append(args, "--grep="+opts.Keyword)
	}
	if opts.Since != "" {
		args = append(args, "--since="+opts.Since)
	}
	if opts.Until != "" {
		args = append(args, "--until="+opts.Until)
	}
	if opts.Path != "" {
		args = append(args, "--", opts.Path)
	}

	out, err := runGitCmd(repoPath, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("git log: %s: %w", strings.TrimSpace(out), err)
	}

	all := parseSearchLog(out)
	total := len(all)
	if opts.Page < 1 {
		opts.Page = 1
	}
	if opts.PageSize < 1 {
		opts.PageSize = 20
	}
	start := (opts.Page - 1) * opts.PageSize
	if start > total {
		start = total
	}
	end := start + opts.PageSize
	if end > total {
		end = total
	}
	return all[start:end], total, nil
}

func parseSearchLog(out string) []*CommitDetailResult {
	var results []*CommitDetailResult
	var current *CommitDetailResult
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 10)
		if len(parts) == 10 && len(parts[0]) >= 7 {
			parents := strings.Fields(strings.TrimSpace(parts[8]))
			current = &CommitDetailResult{
				Hash:           parts[0],
				ShortHash:      parts[1],
				AuthorName:     parts[2],
				AuthorEmail:    parts[3],
				AuthorDate:     parts[4],
				CommitterName:  parts[5],
				CommitterEmail: parts[6],
				CommitterDate:  parts[7],
				ParentHashes:   parents,
				Message:        parts[9],
			}
			results = append(results, current)
			continue
		}
		if current == nil {
			continue
		}
		fs := strings.Fields(line)
		if len(fs) >= 3 {
			var a, d int
			if _, e1 := fmt.Sscanf(fs[0], "%d", &a); e1 == nil {
				if _, e2 := fmt.Sscanf(fs[1], "%d", &d); e2 == nil {
					current.FilesChanged++
					current.Additions += a
					current.Deletions += d
				}
			}
		}
	}
	return results
}

func (s *GitService) GetCommitDetail(repoPath, hash string) (*CommitDetailResult, []*FileChangeResult, error) {
	out, err := runGitCmd(repoPath, "log", "-1", "--pretty=format:%H|%h|%aN|%aE|%aI|%cN|%cE|%cI|%P|%s", "--numstat", hash)
	if err != nil {
		return nil, nil, fmt.Errorf("git log: %s: %w", strings.TrimSpace(out), err)
	}
	commits := parseSearchLog(out)
	if len(commits) == 0 {
		return nil, nil, fmt.Errorf("commit %s not found", hash)
	}

	changesOut, _ := runGitCmd(repoPath, "diff-tree", "--no-commit-id", "-r", "--diff-filter=ACDMRT", hash)
	var files []*FileChangeResult
	for _, line := range strings.Split(changesOut, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// format: :old_mode new_mode old_hash new_hash status\tpath
		if !strings.HasPrefix(line, ":") {
			continue
		}
		line = line[1:]
		tabIdx := strings.Index(line, "\t")
		if tabIdx < 0 {
			continue
		}
		meta := strings.Fields(line[:tabIdx])
		paths := line[tabIdx+1:]
		if len(meta) < 5 {
			continue
		}
		status := meta[4]
		var path, oldPath string
		if strings.Contains(paths, "\t") {
			parts := strings.SplitN(paths, "\t", 2)
			oldPath = parts[0]
			path = parts[1]
		} else {
			path = paths
		}

		var a, d int
		numstatOut, _ := runGitCmd(repoPath, "diff", "--numstat", hash+"^.."+hash, "--", path)
		fs := strings.Fields(numstatOut)
		if len(fs) >= 2 {
			fmt.Sscanf(fs[0], "%d", &a)
			fmt.Sscanf(fs[1], "%d", &d)
		}

		statusMap := map[string]string{
			"A": "added", "C": "added", "M": "modified",
			"D": "deleted", "R": "renamed", "T": "modified",
		}
		files = append(files, &FileChangeResult{
			Path:      path,
			Status:    statusMap[status],
			Additions: a,
			Deletions: d,
			OldPath:   oldPath,
		})
	}

	return commits[0], files, nil
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
