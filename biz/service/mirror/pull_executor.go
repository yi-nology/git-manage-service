package mirror

import (
	"context"
	"fmt"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/branchfilter"
	"github.com/yi-nology/git-manage-service/pkg/gitbackend"
)

type PullExecutor struct {
	backend gitbackend.GitBackend
}

func NewPullExecutor(backend gitbackend.GitBackend) *PullExecutor {
	return &PullExecutor{backend: backend}
}

type PullResult struct {
	BranchesSynced int
	CommitsPulled  int
	Log            string
}

func (e *PullExecutor) Execute(ctx context.Context, mirror *po.Mirror, logf func(string, ...interface{})) (*PullResult, error) {
	result := &PullResult{}

	repoPath := mirror.Repo.Path
	remoteName := mirror.RemoteName
	if remoteName == "" {
		remoteName = "origin"
	}

	filter := branchfilter.New(mirror.BranchFilter)

	logf("Starting pull mirror for repo %s from %s", mirror.Repo.Name, mirror.RemoteURL)

	fetchOpts := gitbackend.FetchOptions{
		RepoPath: repoPath,
		Remote:   remoteName,
		Tags:     mirror.GitTags,
		Prune:    mirror.GitPrune,
		Auth:     e.resolveAuth(mirror),
	}

	if !filter.IsEmpty() {
		allBranches, err := e.backend.ListRemoteBranches(ctx, repoPath, remoteName)
		if err != nil {
			return nil, fmt.Errorf("list remote branches: %w", err)
		}
		filtered := filter.FilterBranches(allBranches)
		fetchOpts.Branches = filtered
		logf("Branch filter: %s → matched %d/%d branches", mirror.BranchFilter, len(filtered), len(allBranches))
	}

	fetchResult, err := e.backend.Fetch(ctx, fetchOpts)
	if err != nil {
		return nil, fmt.Errorf("fetch: %w", err)
	}

	if fetchResult != nil {
		result.BranchesSynced = len(fetchResult.FetchedRefs)
		result.CommitsPulled = len(fetchResult.FetchedRefs)
	}

	logf("Pull completed: %d branches synced", result.BranchesSynced)
	return result, nil
}

func (e *PullExecutor) resolveAuth(mirror *po.Mirror) gitbackend.AuthConfig {
	if mirror.Credential == nil {
		return gitbackend.AuthConfig{Type: "none"}
	}

	cred := mirror.Credential
	switch cred.Type {
	case "ssh_key":
		return gitbackend.AuthConfig{
			Type:   "ssh",
			SSHKey: cred.SSHKeyPath,
		}
	case "http_basic":
		return gitbackend.AuthConfig{
			Type:     "http_basic",
			Username: cred.Username,
			Password: cred.Secret,
		}
	case "http_token":
		return gitbackend.AuthConfig{
			Type:  "http_token",
			Token: cred.Secret,
		}
	default:
		return gitbackend.AuthConfig{Type: "none"}
	}
}

func PreviewPull(ctx context.Context, backend gitbackend.GitBackend, mirror *po.Mirror) (string, error) {
	remoteName := mirror.RemoteName
	if remoteName == "" {
		remoteName = "origin"
	}

	var preview strings.Builder
	preview.WriteString(fmt.Sprintf("Pull mirror: %s → local (%s)\n", mirror.RemoteURL, mirror.Repo.Name))
	preview.WriteString(fmt.Sprintf("Remote: %s\n", remoteName))
	preview.WriteString(fmt.Sprintf("Tags: %v\n", mirror.GitTags))
	preview.WriteString(fmt.Sprintf("Prune: %v\n", mirror.GitPrune))

	if mirror.BranchFilter != "" {
		filter := branchfilter.New(mirror.BranchFilter)
		branches, err := backend.ListRemoteBranches(ctx, mirror.Repo.Path, remoteName)
		if err != nil {
			return preview.String(), nil
		}
		filtered := filter.FilterBranches(branches)
		preview.WriteString(fmt.Sprintf("Branch filter: %s\n", mirror.BranchFilter))
		preview.WriteString(fmt.Sprintf("Matched branches (%d/%d):\n", len(filtered), len(branches)))
		for _, b := range filtered {
			preview.WriteString(fmt.Sprintf("  - %s\n", b))
		}
	} else {
		preview.WriteString("Branch filter: none (all branches)\n")
	}

	return preview.String(), nil
}
