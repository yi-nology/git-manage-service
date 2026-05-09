package git

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	workspaceModel "github.com/yi-nology/git-manage-service/biz/model/workspace"
)

func (s *GitService) GetWorkspaceStatus(repoPath string) (*workspaceModel.GetWorkspaceStatusResp, error) {
	r, err := s.openRepo(repoPath)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}
	w, err := r.Worktree()
	if err != nil {
		return nil, fmt.Errorf("worktree: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return nil, fmt.Errorf("status: %w", err)
	}

	head, err := r.Head()
	var branch string
	if err == nil && head != nil {
		branch = head.Name().Short()
	}

	result := &workspaceModel.GetWorkspaceStatusResp{
		Branch:  branch,
		IsClean: status.IsClean(),
	}

	isMerging, _ := s.isMerging(repoPath)
	isRebasing, _ := s.isRebasing(repoPath)
	result.IsMerging = isMerging
	result.IsRebasing = isRebasing

	ahead, behind := s.getAheadBehind(r, head)
	result.Ahead = int32(ahead)
	result.Behind = int32(behind)

	for path, fs := range status {
		staging := fs.Staging
		worktree := fs.Worktree

		item := &workspaceModel.FileStatus{
			Path: path,
		}
		if staging == 'R' {
			item.OldPath = fs.Extra
		}

		switch {
		case staging == 'U' || worktree == 'U':
			statusStr := "conflicted"
			item.Status = statusStr
			result.Conflicted = append(result.Conflicted, item)
		case staging == 'A' || staging == 'M' || staging == 'D' || staging == 'R' || staging == 'C':
			statusStr := s.readableStatus(string(staging))
			item.Status = statusStr
			result.Staged = append(result.Staged, item)
		case worktree == 'M' || worktree == 'D':
			statusStr := s.readableStatus(string(worktree))
			item.Status = statusStr
			result.Unstaged = append(result.Unstaged, item)
		case worktree == '?' && staging == '?':
			statusStr := "untracked"
			item.Status = statusStr
			result.Untracked = append(result.Untracked, item)
		case worktree == 'A':
			statusStr := "added"
			item.Status = statusStr
			result.Unstaged = append(result.Unstaged, item)
		}
	}

	if result.Staged == nil {
		result.Staged = []*workspaceModel.FileStatus{}
	}
	if result.Unstaged == nil {
		result.Unstaged = []*workspaceModel.FileStatus{}
	}
	if result.Untracked == nil {
		result.Untracked = []*workspaceModel.FileStatus{}
	}
	if result.Conflicted == nil {
		result.Conflicted = []*workspaceModel.FileStatus{}
	}

	return result, nil
}

func (s *GitService) GetWorkspaceDiff(repoPath, file string, stagedOnly bool) (*api.WorkspaceDiff, error) {
	r, err := s.openRepo(repoPath)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}

	head, err := r.Head()
	if err != nil {
		return nil, fmt.Errorf("head: %w", err)
	}

	commitObj, err := r.CommitObject(head.Hash())
	if err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	result := &api.WorkspaceDiff{}
	result.Files = []api.WorkspaceDiffFile{}

	if file != "" {
		df, err := s.diffSingleFile(r, repoPath, commitObj, file, stagedOnly)
		if err != nil {
			return nil, err
		}
		result.Files = append(result.Files, *df)
		result.TotalAdditions = df.Additions
		result.TotalDeletions = df.Deletions
		return result, nil
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}
	status, err := w.Status()
	if err != nil {
		return nil, err
	}

	for path, fs := range status {
		if stagedOnly && (fs.Staging == ' ' || fs.Staging == '?' || fs.Staging == 0) {
			continue
		}
		df, err := s.diffSingleFile(r, repoPath, commitObj, path, stagedOnly)
		if err != nil {
			continue
		}
		result.Files = append(result.Files, *df)
		result.TotalAdditions += df.Additions
		result.TotalDeletions += df.Deletions
	}

	return result, nil
}

func (s *GitService) diffSingleFile(r *git.Repository, repoPath string, commitObj *object.Commit, filePath string, stagedOnly bool) (*api.WorkspaceDiffFile, error) {
	df := &api.WorkspaceDiffFile{File: filePath}

	fromFile, _ := commitObj.File(filePath)

	absPath := filepath.Join(repoPath, filePath)
	_, err := os.Stat(absPath)
	if err != nil {
		df.Diff = ""
		if fromFile != nil {
			df.Deletions = countLines(fromFile)
		}
		return df, nil
	}
	df.IsBinary = isBinaryFile(filePath)

	if df.IsBinary {
		df.Diff = "Binary file"
		return df, nil
	}

	var fromContent string
	if fromFile != nil {
		c, err := fromFile.Contents()
		if err == nil {
			fromContent = c
		}
	}

	toContent, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", filePath, err)
	}

	diff := s.generateDiff(fromContent, string(toContent), filePath)
	df.Diff = diff
	df.Additions, df.Deletions = countDiffStats(diff)

	return df, nil
}

func (s *GitService) StageFiles(repoPath string, files []string, stageAll bool) error {
	r, err := s.openRepo(repoPath)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	if stageAll {
		_, err = w.Add(".")
		return err
	}

	for _, f := range files {
		if _, err := w.Add(f); err != nil {
			return fmt.Errorf("stage %s: %w", f, err)
		}
	}
	return nil
}

func (s *GitService) UnstageFiles(repoPath string, files []string, unstageAll bool) error {
	r, err := s.openRepo(repoPath)
	if err != nil {
		return err
	}
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	if unstageAll {
		head, err := r.Head()
		if err != nil {
			return err
		}
		return w.Reset(&git.ResetOptions{Commit: head.Hash(), Mode: git.MixedReset})
	}

	head, err := r.Head()
	if err != nil {
		return err
	}
	for _, f := range files {
		err := w.Reset(&git.ResetOptions{
			Commit: head.Hash(),
			Mode:   git.MixedReset,
			Files:  []string{f},
		})
		if err != nil {
			return fmt.Errorf("unstage %s: %w", f, err)
		}
	}
	return nil
}

func (s *GitService) CommitChanges(repoPath string, files []string, stageAll bool, message, authorName, authorEmail string, push bool, pushRemote string) (*api.CommitResult, error) {
	if stageAll || len(files) > 0 {
		if err := s.StageFiles(repoPath, files, stageAll); err != nil {
			return nil, fmt.Errorf("stage: %w", err)
		}
	}

	if err := s.Commit(repoPath, message, authorName, authorEmail); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	headHash, _ := s.ResolveRevision(repoPath, "HEAD")
	result := &api.CommitResult{CommitHash: headHash}

	if push {
		remote := pushRemote
		if remote == "" {
			remote = "origin"
		}
		branch, _ := s.GetHeadBranch(repoPath)
		pushErr := s.Push(repoPath, remote, "", branch, nil, nil)
		result.Pushed = pushErr == nil
	}

	return result, nil
}

func (s *GitService) readableStatus(code string) string {
	switch code {
	case "A":
		return "added"
	case "M":
		return "modified"
	case "D":
		return "deleted"
	case "R":
		return "renamed"
	case "C":
		return "copied"
	case "?":
		return "untracked"
	case "U":
		return "conflicted"
	default:
		return code
	}
}

func (s *GitService) isMerging(repoPath string) (bool, error) {
	_, err := os.Stat(filepath.Join(repoPath, ".git", "MERGE_HEAD"))
	return err == nil, nil
}

func (s *GitService) isRebasing(repoPath string) (bool, error) {
	_, err1 := os.Stat(filepath.Join(repoPath, ".git", "rebase-merge"))
	_, err2 := os.Stat(filepath.Join(repoPath, ".git", "rebase-apply"))
	return err1 == nil || err2 == nil, nil
}

func (s *GitService) getAheadBehind(r *git.Repository, head *plumbing.Reference) (int, int) {
	if head == nil {
		return 0, 0
	}

	localHash := head.Hash()
	remoteRef, err := r.Reference(plumbing.ReferenceName("refs/remotes/origin/"+head.Name().Short()), true)
	if err != nil {
		return 0, 0
	}
	remoteHash := remoteRef.Hash()

	ahead := 0
	behind := 0

	localIter, err := r.Log(&git.LogOptions{From: localHash})
	if err == nil {
		localIter.ForEach(func(c *object.Commit) error {
			if c.Hash == remoteHash {
				return fmt.Errorf("stop")
			}
			ahead++
			return nil
		})
		localIter.Close()
	}

	remoteIter, err := r.Log(&git.LogOptions{From: remoteHash})
	if err == nil {
		remoteIter.ForEach(func(c *object.Commit) error {
			if c.Hash == localHash {
				return fmt.Errorf("stop")
			}
			behind++
			return nil
		})
		remoteIter.Close()
	}

	return ahead, behind
}

func (s *GitService) RemoveTracking(repoPath string, files []string) error {
	for _, f := range files {
		_, err := s.RunCommand(repoPath, "rm", "--cached", f)
		if err != nil {
			return fmt.Errorf("rm --cached %s: %w", f, err)
		}
	}
	return nil
}

func (s *GitService) AddToGitignore(repoPath string, patterns []string) error {
	gitignorePath := filepath.Join(repoPath, ".gitignore")
	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open .gitignore: %w", err)
	}
	defer f.Close()

	for _, p := range patterns {
		if _, err := f.WriteString(p + "\n"); err != nil {
			return fmt.Errorf("write .gitignore: %w", err)
		}
	}
	return nil
}

func (s *GitService) GetWorkspaceDiffRaw(repoPath string, file string) (string, error) {
	args := []string{"diff", "--stat", "--patch"}
	if file != "" {
		args = append(args, "--", file)
	}
	out, err := s.RunCommand(repoPath, args...)
	if err != nil {
		return "", fmt.Errorf("git diff: %w", err)
	}
	return out, nil
}
