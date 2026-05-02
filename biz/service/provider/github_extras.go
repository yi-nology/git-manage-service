package provider

import (
	"context"
	"fmt"
	"strings"
)

func (g *githubProvider) ListBranches(ctx context.Context, owner, repo string) ([]*PlatformBranch, error) {
	path := fmt.Sprintf("/repos/%s/%s/branches?per_page=100", owner, repo)
	var branches []struct {
		Name string `json:"name"`
	}
	if err := g.doRequest(ctx, "GET", path, nil, &branches); err != nil {
		return nil, err
	}
	result := make([]*PlatformBranch, 0, len(branches))
	for _, b := range branches {
		result = append(result, &PlatformBranch{Name: b.Name})
	}
	return result, nil
}

func (g *githubProvider) CreateBranch(ctx context.Context, owner, repo, branch, ref string) (*PlatformBranch, error) {
	body := map[string]string{"ref": ref, "branch_name": branch}
	var res struct {
		Ref string `json:"ref"`
	}
	if err := g.doRequest(ctx, "POST", fmt.Sprintf("/repos/%s/%s/git/refs", owner, repo), body, &res); err != nil {
		return nil, err
	}
	name := strings.TrimPrefix(res.Ref, "refs/heads/")
	return &PlatformBranch{Name: name}, nil
}

func (g *githubProvider) DeleteBranch(ctx context.Context, owner, repo, branch string) error {
	return g.doRequest(ctx, "DELETE", fmt.Sprintf("/repos/%s/%s/git/refs/heads/%s", owner, repo, branch), nil, nil)
}

func (g *githubProvider) GetCRDiff(ctx context.Context, owner, repo string, number int) (*MergeDiff, error) {
	return nil, fmt.Errorf("GetCRDiff not yet implemented for GitHub")
}

func (g *githubProvider) GetCRFiles(ctx context.Context, owner, repo string, number int) ([]*ChangedFile, error) {
	return nil, fmt.Errorf("GetCRFiles not yet implemented for GitHub")
}

func (g *githubProvider) CreateNote(ctx context.Context, owner, repo string, number int, body string) (string, error) {
	return "", fmt.Errorf("CreateNote not yet implemented for GitHub")
}

func (g *githubProvider) CreateDiscussion(ctx context.Context, owner, repo string, number int, opts DiscussionOptions) (string, error) {
	return "", fmt.Errorf("CreateDiscussion not yet implemented for GitHub")
}

func (g *githubProvider) CreateCommitStatus(ctx context.Context, owner, repo, sha string, opts CommitStatusOptions) error {
	return fmt.Errorf("CreateCommitStatus not yet implemented for GitHub")
}

func (g *githubProvider) GetFileContent(ctx context.Context, owner, repo, path, ref string) (string, error) {
	return "", fmt.Errorf("GetFileContent not yet implemented for GitHub")
}

func (g *githubProvider) UpdateCRLabels(ctx context.Context, owner, repo string, number int, labels []string) error {
	return fmt.Errorf("UpdateCRLabels not yet implemented for GitHub")
}
