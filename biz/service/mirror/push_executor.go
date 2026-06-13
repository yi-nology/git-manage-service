package mirror

import (
	"context"
	"fmt"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-platform-sdk/gitbackend"
	"github.com/yi-nology/git-platform-sdk/pkg/branchfilter"
)

type PushExecutor struct {
	backend gitbackend.GitBackend
}

func NewPushExecutor(backend gitbackend.GitBackend) *PushExecutor {
	return &PushExecutor{backend: backend}
}

type PushResult struct {
	BranchesSynced int
	CommitsPushed  int
	Log            string
}

func (e *PushExecutor) Execute(ctx context.Context, mirror *po.Mirror, logf func(string, ...interface{})) (*PushResult, error) {
	result := &PushResult{}

	repoPath := mirror.Repo.Path
	remoteName := mirror.RemoteName
	if remoteName == "" {
		remoteName = "origin"
	}

	filter := branchfilter.New(mirror.BranchFilter)

	logf("Starting push mirror for repo %s to %s", mirror.Repo.Name, mirror.RemoteURL)

	pushOpts := gitbackend.PushOptions{
		RepoPath: repoPath,
		Remote:   remoteName,
		Force:    mirror.GitForce,
		Mirror:   filter.IsEmpty(),
		Auth:     resolveAuth(mirror),
	}

	if !filter.IsEmpty() {
		branches, err := e.backend.ListRemoteBranches(ctx, repoPath, remoteName)
		if err != nil {
			return nil, fmt.Errorf("list branches: %w", err)
		}
		filtered := filter.FilterBranches(branches)
		refSpecs := make([]string, 0, len(filtered))
		for _, b := range filtered {
			prefix := "+"
			if !mirror.GitForce {
				prefix = ""
			}
			refSpecs = append(refSpecs, fmt.Sprintf("%srefs/heads/%s:refs/heads/%s", prefix, b, b))
		}
		pushOpts.RefSpecs = refSpecs
		pushOpts.Mirror = false
		logf("Branch filter: %s → pushing %d branches", mirror.BranchFilter, len(filtered))
	}

	pushResult, err := e.backend.Push(ctx, pushOpts)
	if err != nil {
		return nil, fmt.Errorf("push: %w", err)
	}

	if pushResult != nil {
		result.BranchesSynced = len(pushResult.PushedRefs)
		result.CommitsPushed = len(pushResult.PushedRefs)
	}

	logf("Push completed: %d refs pushed", result.BranchesSynced)
	return result, nil
}

func PreviewPush(ctx context.Context, backend gitbackend.GitBackend, mirror *po.Mirror) (string, error) {
	remoteName := mirror.RemoteName
	if remoteName == "" {
		remoteName = "origin"
	}

	var preview strings.Builder
	preview.WriteString(fmt.Sprintf("Push mirror: local (%s) → %s\n", mirror.Repo.Name, mirror.RemoteURL))
	preview.WriteString(fmt.Sprintf("Remote: %s\n", remoteName))
	preview.WriteString(fmt.Sprintf("Force: %v\n", mirror.GitForce))
	preview.WriteString(fmt.Sprintf("Tags: %v\n", mirror.GitTags))

	if mirror.BranchFilter != "" {
		filter := branchfilter.New(mirror.BranchFilter)
		branches, err := backend.ListRemoteBranches(ctx, mirror.Repo.Path, remoteName)
		if err != nil {
			return preview.String(), nil
		}
		filtered := filter.FilterBranches(branches)
		preview.WriteString(fmt.Sprintf("Branch filter: %s\n", mirror.BranchFilter))
		preview.WriteString(fmt.Sprintf("Will push %d/%d branches:\n", len(filtered), len(branches)))
		for _, b := range filtered {
			preview.WriteString(fmt.Sprintf("  - %s\n", b))
		}
	} else {
		preview.WriteString("Branch filter: none (--mirror mode)\n")
	}

	return preview.String(), nil
}
