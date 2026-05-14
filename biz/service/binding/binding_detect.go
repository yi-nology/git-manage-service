package binding

import (
	"fmt"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-platform-sdk/provider"
)

func AutoDetect(repoKey string) (*api.AutoDetectResp, error) {
	repoDAO := db.NewRepoDAO()
	repo, err := repoDAO.FindByKey(repoKey)
	if err != nil {
		return nil, fmt.Errorf("repo not found: %w", err)
	}

	gitSvc := git.NewGitService()
	repoConfig, err := gitSvc.GetRepoConfig(repo.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to read repo config: %w", err)
	}

	providerDAO := db.NewProviderConfigDAO()
	providers, err := providerDAO.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to list providers: %w", err)
	}

	bindingDAO := db.NewRepoProviderBindingDAO()
	existingBindings, _ := bindingDAO.FindByRepoID(repo.ID)
	providerSet := make(map[uint]bool)
	for _, b := range existingBindings {
		providerSet[b.ProviderConfigID] = true
	}

	var suggestions []api.BindingSuggestion

	for _, remote := range repoConfig.Remotes {
		if remote.FetchURL == "" {
			continue
		}

		result, err := provider.DetectPlatform(remote.FetchURL)
		if err != nil {
			continue
		}

		for i := range providers {
			pc := &providers[i]
			if providerSet[pc.ID] {
				continue
			}

			if !isURLMatch(result.BaseURL, pc.BaseURL) {
				continue
			}

			if result.Owner == "" || result.Repo == "" {
				continue
			}

			confidence := "medium"
			if remote.Name == "origin" {
				confidence = "high"
			}

			suggestions = append(suggestions, api.BindingSuggestion{
				ProviderConfigID: pc.ID,
				Platform:         string(pc.Platform),
				PlatformOwner:    result.Owner,
				PlatformRepo:     result.Repo,
				RemoteName:       remote.Name,
				RemoteURL:        remote.FetchURL,
				Confidence:       confidence,
				MatchSource:      "remote_url",
			})
		}
	}

	return &api.AutoDetectResp{Suggestions: suggestions}, nil
}

func FindBindingForRepo(repoKey string, providerConfigID uint) (*po.RepoProviderBinding, error) {
	repoDAO := db.NewRepoDAO()
	repo, err := repoDAO.FindByKey(repoKey)
	if err != nil {
		return nil, err
	}

	bindingDAO := db.NewRepoProviderBindingDAO()

	if providerConfigID > 0 {
		return bindingDAO.FindByRepoAndProvider(repo.ID, providerConfigID)
	}

	return bindingDAO.FindPrimaryByRepoID(repo.ID)
}

func FindBindingByPlatformRepo(providerConfigID uint, owner, repoName string) (*api.RepoProviderBindingDTO, error) {
	dao := db.NewRepoProviderBindingDAO()
	b, err := dao.FindByPlatformRepo(providerConfigID, owner, repoName)
	if err != nil {
		return nil, err
	}
	dto := api.NewBindingDTO(*b)
	return &dto, nil
}

func GetRemotesForRepo(repoKey string) ([]domain.GitRemote, error) {
	repoDAO := db.NewRepoDAO()
	repo, err := repoDAO.FindByKey(repoKey)
	if err != nil {
		return nil, err
	}
	gitSvc := git.NewGitService()
	repoConfig, err := gitSvc.GetRepoConfig(repo.Path)
	if err != nil {
		return nil, err
	}
	return repoConfig.Remotes, nil
}
