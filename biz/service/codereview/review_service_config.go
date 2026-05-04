package codereview

import (
	"fmt"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/configs"
)

func GetRemoteRepoConfig(providerConfigID uint, platformOwner, platformRepo string) (*api.ReviewRepoConfigDTO, error) {
	gc := configs.GlobalConfig.CodeReview
	dao := db.NewReviewRepoConfigDAO()
	cfg, err := dao.FindByRemoteRepo(providerConfigID, platformOwner, platformRepo)
	if err != nil {
		defaultCfg := &po.ReviewRepoConfig{
			ProviderConfigID: providerConfigID,
			PlatformOwner:    platformOwner,
			PlatformRepo:     platformRepo,
			Enabled:          true,
			BlockOnHigh:      gc.BlockOnHigh,
			AutoReviewOnMR:   gc.AutoReviewOnMR,
			LLMProvider:      "",
			MaxFiles:         gc.MaxFiles,
			MaxDiffLines:     gc.MaxDiffLines,
		}
		repos := findLinkedRepos(providerConfigID, platformOwner, platformRepo)
		dto := api.NewReviewRepoConfigDTO(*defaultCfg, repos)
		return &dto, nil
	}
	repos := findLinkedRepos(providerConfigID, platformOwner, platformRepo)
	dto := api.NewReviewRepoConfigDTO(*cfg, repos)
	return &dto, nil
}

func UpdateRemoteRepoConfig(providerConfigID uint, platformOwner, platformRepo string, req api.ReviewRepoConfigDTO) (*api.ReviewRepoConfigDTO, error) {
	cfg := &po.ReviewRepoConfig{
		ProviderConfigID:  providerConfigID,
		PlatformOwner:     platformOwner,
		PlatformRepo:      platformRepo,
		Enabled:           req.Enabled,
		BlockOnHigh:       req.BlockOnHigh,
		AutoReviewOnMR:    req.AutoReviewOnMR,
		LLMProvider:       req.LLMProvider,
		MaxFiles:          req.MaxFiles,
		MaxDiffLines:      req.MaxDiffLines,
		RuleOverridesJSON: req.RuleOverrides,
		ScopeNote:         req.ScopeNote,
	}
	dao := db.NewReviewRepoConfigDAO()
	if err := dao.Upsert(cfg); err != nil {
		return nil, fmt.Errorf("failed to save config: %w", err)
	}
	repos := findLinkedRepos(providerConfigID, platformOwner, platformRepo)
	dto := api.NewReviewRepoConfigDTO(*cfg, repos)
	return &dto, nil
}

func findLinkedRepos(providerConfigID uint, platformOwner, platformRepo string) []api.LinkedRepoDTO {
	bindingDAO := db.NewRepoProviderBindingDAO()
	bindings, err := bindingDAO.FindByPlatformRepo(providerConfigID, platformOwner, platformRepo)
	if err != nil {
		return nil
	}
	var repos []api.LinkedRepoDTO
	if bindings.RepoID > 0 {
		repoDAO := db.NewRepoDAO()
		if r, err := repoDAO.FindByID(bindings.RepoID); err == nil {
			repos = append(repos, api.LinkedRepoDTO{ID: r.ID, Key: r.Key, Name: r.Name})
		}
	}
	allBindings, _ := bindingDAO.FindByProviderConfigID(providerConfigID)
	for _, b := range allBindings {
		if b.PlatformOwner == platformOwner && b.PlatformRepo == platformRepo && b.RepoID > 0 {
			exists := false
			for _, r := range repos {
				if r.ID == b.RepoID {
					exists = true
					break
				}
			}
			if !exists {
				repoDAO := db.NewRepoDAO()
				if r, err := repoDAO.FindByID(b.RepoID); err == nil {
					repos = append(repos, api.LinkedRepoDTO{ID: r.ID, Key: r.Key, Name: r.Name})
				}
			}
		}
	}
	return repos
}
