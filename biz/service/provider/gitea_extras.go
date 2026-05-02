package provider

import (
	"context"
	"fmt"
)

func (g *giteaProvider) ListBranches(ctx context.Context, owner, repo string) ([]*PlatformBranch, error) {
	path := fmt.Sprintf("/api/v1/repos/%s/%s/branches?limit=100", owner, repo)
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

func (g *giteaProvider) CreateBranch(ctx context.Context, owner, repo, branch, ref string) (*PlatformBranch, error) {
	body := map[string]string{"new_branch_name": branch, "old_branch_name": ref}
	var res struct {
		Name string `json:"name"`
	}
	if err := g.doRequest(ctx, "POST", fmt.Sprintf("/api/v1/repos/%s/%s/branches", owner, repo), body, &res); err != nil {
		return nil, err
	}
	return &PlatformBranch{Name: res.Name}, nil
}

func (g *giteaProvider) DeleteBranch(ctx context.Context, owner, repo, branch string) error {
	return g.doRequest(ctx, "DELETE", fmt.Sprintf("/api/v1/repos/%s/%s/branches/%s", owner, repo, branch), nil, nil)
}

func (g *giteaProvider) GetCRDiff(ctx context.Context, owner, repo string, number int) (*MergeDiff, error) {
	return nil, fmt.Errorf("GetCRDiff not yet implemented for Gitea")
}

func (g *giteaProvider) GetCRFiles(ctx context.Context, owner, repo string, number int) ([]*ChangedFile, error) {
	return nil, fmt.Errorf("GetCRFiles not yet implemented for Gitea")
}

func (g *giteaProvider) CreateNote(ctx context.Context, owner, repo string, number int, body string) (string, error) {
	return "", fmt.Errorf("CreateNote not yet implemented for Gitea")
}

func (g *giteaProvider) CreateDiscussion(ctx context.Context, owner, repo string, number int, opts DiscussionOptions) (string, error) {
	return "", fmt.Errorf("CreateDiscussion not yet implemented for Gitea")
}

func (g *giteaProvider) CreateCommitStatus(ctx context.Context, owner, repo, sha string, opts CommitStatusOptions) error {
	return fmt.Errorf("CreateCommitStatus not yet implemented for Gitea")
}

func (g *giteaProvider) GetFileContent(ctx context.Context, owner, repo, path, ref string) (string, error) {
	return "", fmt.Errorf("GetFileContent not yet implemented for Gitea")
}

func (g *giteaProvider) UpdateCRLabels(ctx context.Context, owner, repo string, number int, labels []string) error {
	return fmt.Errorf("UpdateCRLabels not yet implemented for Gitea")
}
