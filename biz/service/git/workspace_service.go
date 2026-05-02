package git

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/yi-nology/git-manage-service/biz/model/api"
)

func (s *GitService) GetWorkspaceStatus(repoPath string) (*api.WorkspaceStatus, error) {
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

	result := &api.WorkspaceStatus{
		Branch: branch,
		IsClean: status.IsClean(),
	}

	isMerging, _ := s.isMerging(repoPath)
	isRebasing, _ := s.isRebasing(repoPath)
	result.IsMerging = isMerging
	result.IsRebasing = isRebasing

	ahead, behind := s.getAheadBehind(r, head)
	result.Ahead = ahead
	result.Behind = behind

	for path, fs := range status {
		staging := fs.Staging
		worktree := fs.Worktree

		item := api.WorkspaceFileStatus{
			Path: path,
		}
		if staging == 'R' {
			item.OldPath = fs.Extra
		}

		switch {
		case staging == 'U' || worktree == 'U':
			item.Status = "conflicted"
			result.Conflicted = append(result.Conflicted, item)
		case staging == 'A' || staging == 'M' || staging == 'D' || staging == 'R' || staging == 'C':
			item.Status = string(staging)
			item.Status = s.readableStatus(string(staging))
			result.Staged = append(result.Staged, item)
		case worktree == 'M' || worktree == 'D':
			item.Status = s.readableStatus(string(worktree))
			result.Unstaged = append(result.Unstaged, item)
		case worktree == '?' && staging == '?':
			item.Status = "untracked"
			result.Untracked = append(result.Untracked, item)
		case worktree == 'A':
			item.Status = "added"
			result.Unstaged = append(result.Unstaged, item)
		}
	}

	if result.Staged == nil {
		result.Staged = []api.WorkspaceFileStatus{}
	}
	if result.Unstaged == nil {
		result.Unstaged = []api.WorkspaceFileStatus{}
	}
	if result.Untracked == nil {
		result.Untracked = []api.WorkspaceFileStatus{}
	}
	if result.Conflicted == nil {
		result.Conflicted = []api.WorkspaceFileStatus{}
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

func (s *GitService) PullWithResolve(repoPath, remote, branch string, fetchOnly bool) (*api.PullResult, error) {
	if remote == "" {
		remote = "origin"
	}

	err := s.Fetch(repoPath, remote, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}

	result := &api.PullResult{
		Status: "fetched",
	}

	if fetchOnly {
		return result, nil
	}

	if branch == "" {
		branch, _ = s.GetHeadBranch(repoPath)
	}

	remoteBranch := fmt.Sprintf("%s/%s", remote, branch)
	head, _ := s.GetHeadBranch(repoPath)

	mergeResult, err := s.MergeDryRun(repoPath, remoteBranch, head)
	if err != nil {
		result.Status = "error"
		return result, nil
	}

	if !mergeResult.Success {
		result.Status = "conflicts"
		result.Conflicts = mergeResult.Conflicts
		return result, nil
	}

	err = s.Merge(repoPath, remoteBranch, head, fmt.Sprintf("Merge pull from %s/%s", remote, branch), false, false)
	if err != nil {
		result.Status = "error"
		return result, nil
	}

	result.Status = "success"
	result.BehindPulled = true
	return result, nil
}

func (s *GitService) GetConflictDetail(repoPath, filePath string) (*api.ConflictDetail, error) {
	absPath := filepath.Join(repoPath, filePath)
	content, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	detail := &api.ConflictDetail{
		Path:           filePath,
		ConflictMarker: string(content),
	}

	ours, _ := s.readMergeStage(repoPath, filePath, 2)
	theirs, _ := s.readMergeStage(repoPath, filePath, 3)
	base, _ := s.readMergeBase(repoPath, filePath)

	detail.OursContent = ours
	detail.TheirsContent = theirs
	detail.BaseContent = base

	return detail, nil
}

func (s *GitService) MarkConflictResolved(repoPath, filePath, resolvedContent string, stage bool) error {
	absPath := filepath.Join(repoPath, filePath)
	err := os.WriteFile(absPath, []byte(resolvedContent), 0644)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	if stage {
		return s.StageFiles(repoPath, []string{filePath}, false)
	}
	return nil
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

func (s *GitService) readMergeStage(repoPath, filePath string, stage int) (string, error) {
	r, err := s.openRepo(repoPath)
	if err != nil {
		return "", err
	}
	obj, err := r.Storer.Index()
	if err != nil {
		return "", err
	}
	for _, e := range obj.Entries {
		if e.Name == filePath {
			blob, err := r.BlobObject(e.Hash)
			if err != nil {
				continue
			}
			reader, err := blob.Reader()
			if err != nil {
				continue
			}
			defer reader.Close()
			buf := new(bytes.Buffer)
			buf.ReadFrom(reader)
			return buf.String(), nil
		}
	}
	return "", nil
}

func (s *GitService) readMergeBase(repoPath, filePath string) (string, error) {
	r, err := s.openRepo(repoPath)
	if err != nil {
		return "", err
	}

	headRef, err := r.Head()
	if err != nil {
		return "", err
	}

	mergeHeadBytes, err := os.ReadFile(filepath.Join(repoPath, ".git", "MERGE_HEAD"))
	if err != nil {
		return "", err
	}

	mergeHash := plumbing.NewHash(strings.TrimSpace(string(mergeHeadBytes)))

	headCommit, err := r.CommitObject(headRef.Hash())
	if err != nil {
		return "", err
	}
	mergeCommit, err := r.CommitObject(mergeHash)
	if err != nil {
		return "", err
	}

	baseCommits, err := headCommit.MergeBase(mergeCommit)
	if err != nil || len(baseCommits) == 0 {
		return "", err
	}

	baseFile, err := baseCommits[0].File(filePath)
	if err != nil {
		return "", nil
	}

	return baseFile.Contents()
}

func (s *GitService) generateDiff(from, to, filePath string) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("--- a/%s\n", filePath))
	buf.WriteString(fmt.Sprintf("+++ b/%s\n", filePath))

	fromLines := splitLines(from)
	toLines := splitLines(to)

	ops := s.computeDiffOps(fromLines, toLines)
	lineNum := 1
	for _, op := range ops {
		switch op.kind {
		case ' ':
			buf.WriteString(fmt.Sprintf(" %s\n", op.line))
			lineNum++
		case '+':
			buf.WriteString(fmt.Sprintf("+%s\n", op.line))
		case '-':
			buf.WriteString(fmt.Sprintf("-%s\n", op.line))
			lineNum++
		}
	}

	return buf.String()
}

type diffOp struct {
	kind byte
	line string
}

func (s *GitService) computeDiffOps(from, to []string) []diffOp {
	var ops []diffOp
	fi, ti := 0, 0
	for fi < len(from) && ti < len(to) {
		if from[fi] == to[ti] {
			ops = append(ops, diffOp{' ', from[fi]})
			fi++
			ti++
		} else {
			ops = append(ops, diffOp{'-', from[fi]})
			fi++
		}
	}
	for fi < len(from) {
		ops = append(ops, diffOp{'-', from[fi]})
		fi++
	}
	for ti < len(to) {
		ops = append(ops, diffOp{'+', to[ti]})
		ti++
	}
	return ops
}

func countDiffStats(diff string) (additions, deletions int) {
	scanner := bufio.NewScanner(strings.NewReader(diff))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > 0 {
			if line[0] == '+' && !strings.HasPrefix(line, "+++") {
				additions++
			} else if line[0] == '-' && !strings.HasPrefix(line, "---") {
				deletions++
			}
		}
	}
	return
}

func splitLines(s string) []string {
	if s == "" {
		return nil
	}
	var lines []string
	scanner := bufio.NewScanner(strings.NewReader(s))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

func countLines(f *object.File) int {
	c, err := f.Contents()
	if err != nil {
		return 0
	}
	return len(splitLines(c))
}

func isBinaryFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	binaryExts := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true,
		".ico": true, ".svg": true, ".pdf": true, ".zip": true, ".tar": true,
		".gz": true, ".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".woff": true, ".woff2": true, ".ttf": true, ".eot": true, ".mp3": true,
		".mp4": true, ".avi": true, ".mov": true, ".wasm": true,
	}
	return binaryExts[ext]
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
