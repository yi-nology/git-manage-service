package mirror

import (
	"context"
	"fmt"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/branchfilter"
	"github.com/yi-nology/git-manage-service/pkg/gitbackend"
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
		Auth:     e.resolveAuth(mirror),
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

func (e *PushExecutor) resolveAuth(mirror *po.Mirror) gitbackend.AuthConfig {
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
