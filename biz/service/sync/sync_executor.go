package sync

import (
	"fmt"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	syncModel "github.com/yi-nology/git-manage-service/biz/model/sync"
)

func (s *SyncService) doSyncSingleBranch(path string, task *po.SyncTask, logf func(string, ...interface{})) (string, error) {
	logf("Starting sync for task %s (Repo: %s)", task.Key, path)

	sourceRemote := task.SourceRemote
	if sourceRemote == "" {
		sourceRemote = "origin"
	}

	isLocalSource := (sourceRemote == "local")
	var sourceHash string
	progressWriter := &logWriter{logf: logf}

	if !isLocalSource {
		sourceURL, _ := s.git.GetRemoteURL(path, sourceRemote)
		if sourceURL == "" && sourceRemote == "origin" {
			sourceURL = task.SourceRepo.RemoteURL
		}

		sRefSpec := fmt.Sprintf("+refs/heads/%s:refs/remotes/%s/%s", task.SourceBranch, sourceRemote, task.SourceBranch)
		logf("Command: git fetch %s %s", sourceRemote, sRefSpec)

		if err := s.fetchRemote(path, task.SourceRepo, sourceRemote, sourceURL, sRefSpec, progressWriter, logf); err != nil {
			return "", fmt.Errorf("fetch source failed: %v", err)
		}

		h, err := s.git.GetCommitHash(path, task.SourceRemote, task.SourceBranch)
		if err != nil {
			return "", fmt.Errorf("get source hash failed: %v", err)
		}
		sourceHash = h
	} else {
		logf("Using local branch: %s", task.SourceBranch)
		h, err := s.git.ResolveRevision(path, task.SourceBranch)
		if err != nil {
			return "", fmt.Errorf("get local source hash failed: %v", err)
		}
		sourceHash = h
	}

	logf("Source hash (%s/%s): %s", task.SourceRemote, task.SourceBranch, sourceHash)

	targetRemote := task.TargetRemote
	if targetRemote == "" {
		targetRemote = "origin"
	}

	targetURL, _ := s.git.GetRemoteURL(path, targetRemote)
	if targetURL == "" && targetRemote == "origin" {
		targetURL = task.TargetRepo.RemoteURL
	}

	tRefSpec := fmt.Sprintf("+refs/heads/%s:refs/remotes/%s/%s", task.TargetBranch, targetRemote, task.TargetBranch)
	logf("Command: git fetch %s %s", targetRemote, tRefSpec)

	if err := s.fetchRemote(path, task.TargetRepo, targetRemote, targetURL, tRefSpec, progressWriter, logf); err != nil {
		return "", fmt.Errorf("fetch target failed: %v", err)
	}

	targetHash, err := s.git.GetCommitHash(path, task.TargetRemote, task.TargetBranch)
	targetExists := err == nil

	if targetExists {
		logf("Target hash (%s/%s): %s", task.TargetRemote, task.TargetBranch, targetHash)
	} else {
		logf("Target branch does not exist yet")
	}

	var commitRange string
	if targetExists {
		commitRange = fmt.Sprintf("%s..%s", targetHash, sourceHash)
	} else {
		commitRange = sourceHash
	}

	if targetExists {
		if sourceHash == targetHash {
			logf("Source and Target are at the same commit. No sync needed.")
			return "", nil
		}

		isAncestor, err := s.git.IsAncestor(path, targetHash, sourceHash)
		if err != nil {
			return "", fmt.Errorf("check ancestor failed: %v", err)
		}

		if !isAncestor {
			logf("Not a fast-forward update. Checking divergence...")
			isSourceBehind, _ := s.git.IsAncestor(path, sourceHash, targetHash)
			if isSourceBehind {
				return "", fmt.Errorf("source is behind target")
			}
			return "", fmt.Errorf("conflict")
		}
		logf("Fast-forward check passed.")
	}

	var pushOpts []string
	if task.PushOptions != "" {
		pushOpts = strings.Fields(task.PushOptions)
	}

	cmdStr := fmt.Sprintf("git push %s %s:refs/heads/%s", task.TargetRemote, sourceHash, task.TargetBranch)
	if len(pushOpts) > 0 {
		cmdStr += " " + strings.Join(pushOpts, " ")
	}
	logf("Command: %s", cmdStr)
	logf("Pushing to %s/%s with options: %v", task.TargetRemote, task.TargetBranch, pushOpts)

	if err := s.pushRemote(path, task.TargetRepo, targetRemote, targetURL, sourceHash, task.TargetBranch, pushOpts, progressWriter, logf); err != nil {
		return "", fmt.Errorf("push failed: %v", err)
	}

	return commitRange, nil
}

func (s *SyncService) doSyncAllBranches(path string, task *po.SyncTask, logf func(string, ...interface{})) (string, error) {
	logf("Starting all-branch sync for task %s (Repo: %s)", task.Key, path)

	sourceRemote := task.SourceRemote
	if sourceRemote == "" {
		sourceRemote = "origin"
	}
	targetRemote := task.TargetRemote
	if targetRemote == "" {
		targetRemote = "origin"
	}

	progressWriter := &logWriter{logf: logf}

	isLocalSource := (sourceRemote == "local")
	if !isLocalSource {
		sourceURL, _ := s.git.GetRemoteURL(path, sourceRemote)
		if sourceURL == "" && sourceRemote == "origin" {
			sourceURL = task.SourceRepo.RemoteURL
		}

		allRefSpec := fmt.Sprintf("+refs/heads/*:refs/remotes/%s/*", sourceRemote)
		logf("Command: git fetch %s %s", sourceRemote, allRefSpec)

		if err := s.fetchRemote(path, task.SourceRepo, sourceRemote, sourceURL, allRefSpec, progressWriter, logf); err != nil {
			return "", fmt.Errorf("fetch source (all branches) failed: %v", err)
		}
	}

	var branches []string
	if isLocalSource {
		allBranches, err := s.git.GetBranches(path)
		if err != nil {
			return "", fmt.Errorf("list local branches failed: %v", err)
		}
		for _, b := range allBranches {
			if !strings.Contains(b, "/") {
				branches = append(branches, b)
			}
		}
	} else {
		var err error
		branches, err = s.git.ListRemoteBranches(path, sourceRemote)
		if err != nil {
			return "", fmt.Errorf("list remote branches failed: %v", err)
		}
	}

	if len(branches) == 0 {
		logf("No branches found on source remote %s", sourceRemote)
		return "", nil
	}
	logf("Found %d branches on source: %v", len(branches), branches)

	targetURL, _ := s.git.GetRemoteURL(path, targetRemote)
	if targetURL == "" && targetRemote == "origin" {
		targetURL = task.TargetRepo.RemoteURL
	}

	tAllRefSpec := fmt.Sprintf("+refs/heads/*:refs/remotes/%s/*", targetRemote)
	logf("Command: git fetch %s %s", targetRemote, tAllRefSpec)

	if err := s.fetchRemote(path, task.TargetRepo, targetRemote, targetURL, tAllRefSpec, progressWriter, logf); err != nil {
		return "", fmt.Errorf("fetch target (all branches) failed: %v", err)
	}

	var pushOpts []string
	if task.PushOptions != "" {
		pushOpts = strings.Fields(task.PushOptions)
	}

	successCount := 0
	failedCount := 0
	skippedCount := 0
	var allCommitRanges []string
	var lastErr error

	for _, branch := range branches {
		logf("--- Syncing branch: %s ---", branch)

		var sourceHash string
		if isLocalSource {
			h, err := s.git.ResolveRevision(path, branch)
			if err != nil {
				logf("  Skip branch %s: cannot resolve local ref: %v", branch, err)
				failedCount++
				lastErr = err
				continue
			}
			sourceHash = h
		} else {
			h, err := s.git.GetCommitHash(path, sourceRemote, branch)
			if err != nil {
				logf("  Skip branch %s: cannot get source hash: %v", branch, err)
				failedCount++
				lastErr = err
				continue
			}
			sourceHash = h
		}
		logf("  Source hash: %s", sourceHash)

		targetHash, err := s.git.GetCommitHash(path, targetRemote, branch)
		targetExists := err == nil

		if targetExists {
			logf("  Target hash: %s", targetHash)
			if sourceHash == targetHash {
				logf("  Already in sync, skipping")
				skippedCount++
				continue
			}

			isAncestor, err := s.git.IsAncestor(path, targetHash, sourceHash)
			if err != nil {
				logf("  Branch %s: ancestor check failed: %v", branch, err)
				failedCount++
				lastErr = err
				continue
			}
			if !isAncestor {
				isSourceBehind, _ := s.git.IsAncestor(path, sourceHash, targetHash)
				if isSourceBehind {
					logf("  Branch %s: source is behind target, skipping", branch)
					failedCount++
					lastErr = fmt.Errorf("branch %s: source is behind target", branch)
					continue
				}
				logf("  Branch %s: conflict (not fast-forward)", branch)
				failedCount++
				lastErr = fmt.Errorf("branch %s: conflict", branch)
				continue
			}
			logf("  Fast-forward check passed")
			allCommitRanges = append(allCommitRanges, fmt.Sprintf("%s: %s..%s", branch, targetHash[:8], sourceHash[:8]))
		} else {
			logf("  Target branch does not exist yet (new branch)")
			allCommitRanges = append(allCommitRanges, fmt.Sprintf("%s: (new) %s", branch, sourceHash[:8]))
		}

		logf("  Pushing %s to %s/%s...", sourceHash[:8], targetRemote, branch)
		if err := s.pushRemote(path, task.TargetRepo, targetRemote, targetURL, sourceHash, branch, pushOpts, progressWriter, logf); err != nil {
			logf("  Branch %s: push failed: %v", branch, err)
			failedCount++
			lastErr = err
			continue
		}

		logf("  Branch %s synced successfully", branch)
		successCount++
	}

	logf("=== All-branch sync summary ===")
	logf("Total: %d, Success: %d, Failed: %d, Skipped (up-to-date): %d", len(branches), successCount, failedCount, skippedCount)

	commitRange := strings.Join(allCommitRanges, "; ")

	if failedCount > 0 && successCount == 0 {
		return commitRange, fmt.Errorf("all branches failed, last error: %v", lastErr)
	}
	if failedCount > 0 {
		return commitRange, fmt.Errorf("%d/%d branches failed, last error: %v", failedCount, len(branches), lastErr)
	}
	return commitRange, nil
}

func (s *SyncService) PreviewSync(repo po.Repo, sourceRemote, sourceBranch, targetRemote, targetBranch string, gitTags, gitForce, gitPrune, gitNoVerify bool) (*syncModel.PreviewSyncResponse, error) {
	path := repo.Path

	var opts []string
	if gitTags {
		opts = append(opts, "--tags")
	}
	if gitForce {
		opts = append(opts, "--force")
	}
	if gitPrune {
		opts = append(opts, "--prune")
	}
	if gitNoVerify {
		opts = append(opts, "--no-verify")
	}

	skipTLS := db.NewRepoProviderBindingDAO().GetSkipTLSForRepo(repo.ID)

	cmdParts := []string{"git push", targetRemote}
	if sourceRemote == "local" {
		cmdParts = append(cmdParts, sourceBranch+":refs/heads/"+targetBranch)
	} else {
		cmdParts = append(cmdParts, "HEAD:refs/heads/"+targetBranch)
	}
	if len(opts) > 0 {
		cmdParts = append(cmdParts, strings.Join(opts, " "))
	}

	response := &syncModel.PreviewSyncResponse{
		Command:     strings.Join(cmdParts, " "),
		FastForward: true,
	}

	isLocalSource := (sourceRemote == "local")
	var sourceHash string

	if !isLocalSource {
		if err := s.git.Fetch(path, sourceRemote, nil, skipTLS); err != nil {
			return nil, fmt.Errorf("fetch source failed: %v", err)
		}
		h, err := s.git.GetCommitHash(path, sourceRemote, sourceBranch)
		if err != nil {
			return nil, fmt.Errorf("get source hash failed: %v", err)
		}
		sourceHash = h
	} else {
		h, err := s.git.ResolveRevision(path, sourceBranch)
		if err != nil {
			return nil, fmt.Errorf("get local source hash failed: %v", err)
		}
		sourceHash = h
	}

	if err := s.git.Fetch(path, targetRemote, nil, skipTLS); err != nil {
		_ = err
	}

	targetHash, err := s.git.GetCommitHash(path, targetRemote, targetBranch)
	if err != nil {
		response.Warning = "Target branch does not exist yet. A new branch will be created."
	}

	if targetHash != "" && sourceHash == targetHash {
		response.Warning = "Source and target are already in sync."
		return response, nil
	}

	if targetHash != "" {
		isAncestor, _ := s.git.IsAncestor(path, targetHash, sourceHash)
		if !isAncestor {
			response.FastForward = false
			response.Warning = "Not a fast-forward update. Force push may be required."
		}
	}

	commitRange := sourceHash
	if targetHash != "" {
		commitRange = targetHash + ".." + sourceHash
	}
	commitLog, err := s.git.GetCommits(path, sourceBranch, commitRange, "")
	if err == nil && commitLog != "" {
		lines := strings.Split(strings.TrimSpace(commitLog), "\n")
		validLines := 0
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				validLines++
			}
		}
		if validLines > 0 {
			response.CommitsToPush = int32(validLines)
		}
	}

	if gitTags {
		tags, _ := s.git.GetTags(path)
		if len(tags) > 0 {
			response.TagsToPush = tags
		}
	}

	return response, nil
}
