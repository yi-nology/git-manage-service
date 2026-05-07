package binding

import (
	"context"
	"fmt"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/provider"
)

func RegisterWebhook(ctx context.Context, id uint) (*api.RepoProviderBindingDTO, error) {
	dao := db.NewRepoProviderBindingDAO()
	b, err := dao.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("binding not found: %w", err)
	}

	if b.WebhookID != "" {
		return nil, fmt.Errorf("webhook already registered for this binding")
	}

	pcDAO := db.NewProviderConfigDAO()
	pc, err := pcDAO.FindByID(b.ProviderConfigID)
	if err != nil {
		return nil, fmt.Errorf("provider config not found: %w", err)
	}

	p, err := provider.GetManager().GetProvider(b.ProviderConfigID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	wh, err := p.CreateWebhook(ctx, provider.CreateWebhookOptions{
		Owner:  b.PlatformOwner,
		Repo:   b.PlatformRepo,
		URL:    pc.WebhookEndpoint,
		Secret: pc.WebhookSecret,
		Events: []string{"push", "merge_request", "pull_request", "tag_push"},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to register webhook: %w", err)
	}

	b.WebhookID = fmt.Sprintf("%d", wh.ID)
	b.WebhookURL = wh.URL
	if err := dao.Save(b); err != nil {
		return nil, fmt.Errorf("failed to save webhook info: %w", err)
	}

	ruleDAO := db.NewWebhookRuleDAO()
	repoPattern := b.PlatformOwner + "/" + b.PlatformRepo
	existingRules, _ := ruleDAO.FindByProviderConfigID(b.ProviderConfigID)
	ruleExists := false
	for _, r := range existingRules {
		if r.Action == "code_review" && r.RepoPattern == repoPattern {
			ruleExists = true
			break
		}
	}
	if !ruleExists {
		ruleDAO.Create(&po.WebhookRule{
			Name:             fmt.Sprintf("auto-review-%s-%s", b.PlatformOwner, b.PlatformRepo),
			ProviderConfigID: b.ProviderConfigID,
			EventTypePattern: "cr.*",
			RepoPattern:      repoPattern,
			Action:           "code_review",
			Enabled:          true,
		})
	}

	updated, err := dao.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("webhook registered but failed to reload: %w", err)
	}
	dto := api.NewBindingDTO(*updated)
	return &dto, nil
}

func DeleteWebhook(ctx context.Context, id uint) (*api.RepoProviderBindingDTO, error) {
	dao := db.NewRepoProviderBindingDAO()
	b, err := dao.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("binding not found: %w", err)
	}

	if b.WebhookID == "" {
		return nil, fmt.Errorf("no webhook registered for this binding")
	}

	p, err := provider.GetManager().GetProvider(b.ProviderConfigID)
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	var webhookInt int64
	fmt.Sscanf(b.WebhookID, "%d", &webhookInt)
	if err := p.DeleteWebhook(ctx, b.PlatformOwner, b.PlatformRepo, webhookInt); err != nil {
		return nil, fmt.Errorf("failed to delete webhook: %w", err)
	}

	b.WebhookID = ""
	b.WebhookURL = ""
	if err := dao.Save(b); err != nil {
		return nil, fmt.Errorf("failed to update binding: %w", err)
	}

	updated, err := dao.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("webhook deleted but failed to reload: %w", err)
	}
	dto := api.NewBindingDTO(*updated)
	return &dto, nil
}
