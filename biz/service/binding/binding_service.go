package binding

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/biz/service/provider"
)

func ListBindings(repoKey string, providerConfigID uint) ([]api.RepoProviderBindingDTO, error) {
	dao := db.NewRepoProviderBindingDAO()

	if repoKey != "" {
		bindings, err := dao.FindByRepoKey(repoKey)
		if err != nil {
			return nil, fmt.Errorf("failed to list bindings: %w", err)
		}
		dtos := make([]api.RepoProviderBindingDTO, 0, len(bindings))
		for i := range bindings {
			dtos = append(dtos, api.NewBindingDTO(bindings[i]))
		}
		return dtos, nil
	}

	if providerConfigID > 0 {
		bindings, err := dao.FindByProviderConfigID(providerConfigID)
		if err != nil {
			return nil, fmt.Errorf("failed to list bindings: %w", err)
		}
		dtos := make([]api.RepoProviderBindingDTO, 0, len(bindings))
		for i := range bindings {
			dtos = append(dtos, api.NewBindingDTO(bindings[i]))
		}
		return dtos, nil
	}

	bindings, err := dao.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to list bindings: %w", err)
	}
	dtos := make([]api.RepoProviderBindingDTO, 0, len(bindings))
	for i := range bindings {
		dtos = append(dtos, api.NewBindingDTO(bindings[i]))
	}
	return dtos, nil
}

func GetBinding(id uint) (*api.RepoProviderBindingDTO, error) {
	dao := db.NewRepoProviderBindingDAO()
	b, err := dao.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("binding not found: %w", err)
	}
	dto := api.NewBindingDTO(*b)
	return &dto, nil
}

func CreateBinding(ctx context.Context, req *api.CreateBindingReq) (*api.RepoProviderBindingDTO, error) {
	repoDAO := db.NewRepoDAO()
	repo, err := repoDAO.FindByKey(req.RepoKey)
	if err != nil {
		return nil, fmt.Errorf("repo not found: %w", err)
	}

	providerDAO := db.NewProviderConfigDAO()
	pc, err := providerDAO.FindByID(req.ProviderConfigID)
	if err != nil {
		return nil, fmt.Errorf("provider config not found: %w", err)
	}

	bindingDAO := db.NewRepoProviderBindingDAO()

	exists, _ := bindingDAO.ExistsByRepoAndProvider(repo.ID, req.ProviderConfigID)
	if exists {
		return nil, fmt.Errorf("repo %s already has a binding to provider %d", req.RepoKey, req.ProviderConfigID)
	}

	exists, _ = bindingDAO.ExistsByPlatformRepo(req.ProviderConfigID, req.PlatformOwner, req.PlatformRepo)
	if exists {
		return nil, fmt.Errorf("remote repo %s/%s on provider %d is already linked to another local repo",
			req.PlatformOwner, req.PlatformRepo, req.ProviderConfigID)
	}

	p, err := provider.GetManager().GetProvider(req.ProviderConfigID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	platformRepoID := ""
	platformRepo, err := p.GetRepo(ctx, req.PlatformOwner, req.PlatformRepo)
	if err == nil && platformRepo != nil {
		platformRepoID = fmt.Sprintf("%d", platformRepo.ID)
	}

	if req.IsPrimary {
		if err := bindingDAO.ClearPrimaryByRepoID(repo.ID); err != nil {
			return nil, fmt.Errorf("failed to clear existing primary: %w", err)
		}
	}

	remoteName := req.RemoteName
	if remoteName == "" {
		remoteName = detectRemoteName(repo.Path, pc, req.PlatformOwner, req.PlatformRepo)
	}

	binding := &po.RepoProviderBinding{
		RepoID:           repo.ID,
		ProviderConfigID: req.ProviderConfigID,
		PlatformOwner:    req.PlatformOwner,
		PlatformRepo:     req.PlatformRepo,
		PlatformRepoID:   platformRepoID,
		RemoteName:       remoteName,
		IsPrimary:        req.IsPrimary,
		Status:           "active",
	}

	if req.RegisterWebhook && pc.WebhookEndpoint != "" {
		secret := pc.WebhookSecret
		wh, err := p.CreateWebhook(ctx, provider.CreateWebhookOptions{
			Owner:  req.PlatformOwner,
			Repo:   req.PlatformRepo,
			URL:    pc.WebhookEndpoint,
			Secret: secret,
			Events: []string{"push", "merge_request", "pull_request", "tag_push"},
		})
		if err == nil {
			binding.WebhookID = fmt.Sprintf("%d", wh.ID)
			binding.WebhookURL = wh.URL
		} else {
			log.Printf("Warning: failed to register webhook: %v", err)
		}
	}

	if err := bindingDAO.Create(binding); err != nil {
		return nil, fmt.Errorf("failed to create binding: %w", err)
	}

	if binding.WebhookID != "" {
		ruleDAO := db.NewWebhookRuleDAO()
		repoPattern := req.PlatformOwner + "/" + req.PlatformRepo
		existingRules, _ := ruleDAO.FindByProviderConfigID(req.ProviderConfigID)
		ruleExists := false
		for _, r := range existingRules {
			if r.Action == "code_review" && r.RepoPattern == repoPattern {
				ruleExists = true
				break
			}
		}
		if !ruleExists {
			ruleDAO.Create(&po.WebhookRule{
				Name:             fmt.Sprintf("auto-review-%s-%s", req.PlatformOwner, req.PlatformRepo),
				ProviderConfigID: req.ProviderConfigID,
				EventTypePattern: "cr.*",
				RepoPattern:      repoPattern,
				Action:           "code_review",
				Enabled:          true,
			})
		}
	}

	created, err := bindingDAO.FindByID(binding.ID)
	if err != nil {
		return nil, fmt.Errorf("binding created but failed to reload: %w", err)
	}
	dto := api.NewBindingDTO(*created)
	return &dto, nil
}

func UpdateBinding(id uint, req *api.UpdateBindingReq) (*api.RepoProviderBindingDTO, error) {
	dao := db.NewRepoProviderBindingDAO()
	b, err := dao.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("binding not found: %w", err)
	}

	if req.RemoteName != nil {
		b.RemoteName = *req.RemoteName
	}
	if req.IsPrimary != nil && *req.IsPrimary {
		if err := dao.ClearPrimaryByRepoID(b.RepoID); err != nil {
			return nil, fmt.Errorf("failed to clear existing primary: %w", err)
		}
		b.IsPrimary = true
	}
	if req.PlatformRepoID != nil {
		b.PlatformRepoID = *req.PlatformRepoID
	}

	if err := dao.Save(b); err != nil {
		return nil, fmt.Errorf("failed to update binding: %w", err)
	}

	updated, err := dao.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("binding updated but failed to reload: %w", err)
	}
	dto := api.NewBindingDTO(*updated)
	return &dto, nil
}

func DeleteBinding(ctx context.Context, id uint, cleanupWebhook bool) error {
	dao := db.NewRepoProviderBindingDAO()
	b, err := dao.FindByID(id)
	if err != nil {
		return fmt.Errorf("binding not found: %w", err)
	}

	if cleanupWebhook && b.WebhookID != "" {
		p, err := provider.GetManager().GetProvider(b.ProviderConfigID)
		if err == nil {
			var webhookInt int64
			fmt.Sscanf(b.WebhookID, "%d", &webhookInt)
			_ = p.DeleteWebhook(ctx, b.PlatformOwner, b.PlatformRepo, webhookInt)
		}
	}

	return dao.SoftDelete(id)
}

func SetPrimary(id uint) (*api.RepoProviderBindingDTO, error) {
	dao := db.NewRepoProviderBindingDAO()
	b, err := dao.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("binding not found: %w", err)
	}

	if err := dao.ClearPrimaryByRepoID(b.RepoID); err != nil {
		return nil, fmt.Errorf("failed to clear existing primary: %w", err)
	}
	b.IsPrimary = true
	if err := dao.Save(b); err != nil {
		return nil, fmt.Errorf("failed to set primary: %w", err)
	}

	updated, err := dao.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("primary set but failed to reload: %w", err)
	}
	dto := api.NewBindingDTO(*updated)
	return &dto, nil
}

func detectRemoteName(repoPath string, pc *po.ProviderConfig, owner, repoName string) string {
	gitSvc := git.NewGitService()
	repoConfig, err := gitSvc.GetRepoConfig(repoPath)
	if err != nil {
		return ""
	}

	for _, remote := range repoConfig.Remotes {
		if remote.FetchURL == "" {
			continue
		}
		result, err := provider.DetectPlatform(remote.FetchURL)
		if err != nil {
			continue
		}
		if result.Owner == owner && result.Repo == repoName && isURLMatch(result.BaseURL, pc.BaseURL) {
			return remote.Name
		}
	}
	return ""
}

func isURLMatch(detectedBaseURL, configBaseURL string) bool {
	d1, err1 := url.Parse(detectedBaseURL)
	d2, err2 := url.Parse(configBaseURL)
	if err1 != nil || err2 != nil {
		return strings.EqualFold(detectedBaseURL, configBaseURL)
	}
	return strings.EqualFold(d1.Host, d2.Host)
}
