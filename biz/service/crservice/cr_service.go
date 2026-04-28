package crservice

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/provider"
)

func CreateCR(ctx context.Context, req *api.CreateCRReq) (*api.CRDTO, error) {
	repo, p, owner, repoName, providerCfgID, err := resolveRepoProvider(req.RepoKey)
	if err != nil {
		return nil, err
	}
	cr, err := p.CreateCR(ctx, provider.CreateCROptions{
		Owner: owner, Repo: repoName, Title: req.Title, Description: req.Description,
		SourceBranch: req.SourceBranch, TargetBranch: req.TargetBranch,
		Labels: req.Labels, RemoveSourceBranch: req.RemoveSourceBranch,
	})
	if err != nil {
		return nil, err
	}
	localCR := platformCRToLocal(repo.ID, providerCfgID, cr)
	crDAO := db.NewChangeRequestDAO()
	if err := crDAO.Create(localCR); err != nil {
		log.Printf("Warning: failed to save CR locally: %v", err)
	}
	return toCRDTO(localCR), nil
}

func GetCR(ctx context.Context, repoKey string, crNumber int) (*api.CRDTO, error) {
	repoDAO := db.NewRepoDAO()
	repo, err := repoDAO.FindByKey(repoKey)
	if err != nil {
		return nil, fmt.Errorf("repo not found: %w", err)
	}
	crDAO := db.NewChangeRequestDAO()
	localCR, err := crDAO.FindByRepoAndNumber(repo.ID, crNumber)
	if err == nil {
		return toCRDTO(localCR), nil
	}
	_, p, owner, repoName, _, err := resolveRepoProvider(repoKey)
	if err != nil {
		return nil, err
	}
	cr, err := p.GetCR(ctx, owner, repoName, crNumber)
	if err != nil {
		return nil, err
	}
	return platformCRToAPI(cr), nil
}

func ListCRs(ctx context.Context, repoKey, state, sourceBranch, targetBranch string, page, pageSize int) ([]api.CRDTO, int, error) {
	repoDAO := db.NewRepoDAO()
	repo, err := repoDAO.FindByKey(repoKey)
	if err != nil {
		return nil, 0, fmt.Errorf("repo not found: %w", err)
	}
	crDAO := db.NewChangeRequestDAO()
	localCRs, total, err := crDAO.FindByRepo(repo.ID, state, sourceBranch, targetBranch, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	dtos := make([]api.CRDTO, 0, len(localCRs))
	for i := range localCRs {
		dtos = append(dtos, *toCRDTO(&localCRs[i]))
	}
	return dtos, int(total), nil
}

func MergeCR(ctx context.Context, repoKey string, crNumber int, mergeMsg string, squash, removeBranch bool) (*api.CRDTO, error) {
	repo, p, owner, repoName, _, err := resolveRepoProvider(repoKey)
	if err != nil {
		return nil, err
	}
	cr, err := p.MergeCR(ctx, owner, repoName, crNumber, provider.MergeCROptions{
		MergeCommitMessage: mergeMsg, Squash: squash, RemoveSourceBranch: removeBranch,
	})
	if err != nil {
		return nil, err
	}
	crDAO := db.NewChangeRequestDAO()
	localCR, dbErr := crDAO.FindByRepoAndNumber(repo.ID, crNumber)
	if dbErr == nil {
		now := time.Now()
		localCR.State = "merged"
		localCR.MergedAt = &now
		crDAO.Save(localCR)
	}
	return platformCRToAPI(cr), nil
}

func CloseCR(ctx context.Context, repoKey string, crNumber int) (*api.CRDTO, error) {
	repo, p, owner, repoName, _, err := resolveRepoProvider(repoKey)
	if err != nil {
		return nil, err
	}
	cr, err := p.CloseCR(ctx, owner, repoName, crNumber)
	if err != nil {
		return nil, err
	}
	crDAO := db.NewChangeRequestDAO()
	localCR, dbErr := crDAO.FindByRepoAndNumber(repo.ID, crNumber)
	if dbErr == nil {
		now := time.Now()
		localCR.State = "closed"
		localCR.ClosedAt = &now
		crDAO.Save(localCR)
	}
	return platformCRToAPI(cr), nil
}

func SyncCRs(ctx context.Context, repoKey, state string) (int, error) {
	repo, p, owner, repoName, providerCfgID, err := resolveRepoProvider(repoKey)
	if err != nil {
		return 0, err
	}
	crs, _, err := p.ListCRs(ctx, provider.ListCROptions{
		Owner: owner, Repo: repoName, State: provider.CRState(state), Page: 1, PerPage: 100,
	})
	if err != nil {
		return 0, err
	}
	crDAO := db.NewChangeRequestDAO()
	synced := 0
	for _, cr := range crs {
		localCR := platformCRToLocal(repo.ID, providerCfgID, cr)
		existing, dbErr := crDAO.FindByRepoAndNumber(repo.ID, cr.Number)
		if dbErr != nil {
			if saveErr := crDAO.Create(localCR); saveErr == nil {
				synced++
			}
		} else {
			existing.State = string(cr.State)
			existing.Title = cr.Title
			existing.Description = cr.Description
			existing.MergeStatus = cr.MergeStatus
			existing.Labels = cr.Labels
			existing.WebURL = cr.WebURL
			if cr.State == provider.CRStateMerged {
				now := time.Now()
				existing.MergedAt = &now
			}
			if cr.State == provider.CRStateClosed {
				now := time.Now()
				existing.ClosedAt = &now
			}
			crDAO.Save(existing)
			synced++
		}
	}
	return synced, nil
}

func resolveRepoProvider(repoKey string) (*po.Repo, provider.Provider, string, string, uint, error) {
	repoDAO := db.NewRepoDAO()
	repo, err := repoDAO.FindByKey(repoKey)
	if err != nil {
		return nil, nil, "", "", 0, fmt.Errorf("repo not found: %w", err)
	}

	bindingDAO := db.NewRepoProviderBindingDAO()
	b, err := bindingDAO.FindPrimaryByRepoID(repo.ID)
	if err == nil && b != nil {
		p, err := provider.GetManager().GetProvider(b.ProviderConfigID)
		if err != nil {
			return nil, nil, "", "", 0, err
		}
		return repo, p, b.PlatformOwner, b.PlatformRepo, b.ProviderConfigID, nil
	}

	if repo.ProviderConfigID == 0 {
		return nil, nil, "", "", 0, fmt.Errorf("repo %s has no provider configured", repoKey)
	}
	p, err := provider.GetManager().GetProvider(repo.ProviderConfigID)
	if err != nil {
		return nil, nil, "", "", 0, err
	}
	owner := repo.PlatformOwner
	repoName := repo.PlatformRepo
	if owner == "" || repoName == "" {
		return nil, nil, "", "", 0, fmt.Errorf("repo %s missing platform owner/repo info", repoKey)
	}
	return repo, p, owner, repoName, repo.ProviderConfigID, nil
}

func platformCRToLocal(repoID, providerConfigID uint, cr *provider.ChangeRequest) *po.ChangeRequest {
	var labels []string
	if cr.Labels != nil {
		labels = cr.Labels
	}
	authorName := ""
	authorUsername := ""
	if cr.Author != nil {
		authorName = cr.Author.Name
		authorUsername = cr.Author.Username
	}
	return &po.ChangeRequest{
		RepoID:           repoID,
		ProviderConfigID: providerConfigID,
		PlatformCRID:     cr.ID,
		CRNumber:         cr.Number,
		Title:            cr.Title,
		Description:      cr.Description,
		State:            string(cr.State),
		SourceBranch:     cr.SourceBranch,
		TargetBranch:     cr.TargetBranch,
		AuthorName:       authorName,
		AuthorUsername:   authorUsername,
		WebURL:           cr.WebURL,
		MergeStatus:      cr.MergeStatus,
		Labels:           labels,
	}
}

func platformCRToAPI(cr *provider.ChangeRequest) *api.CRDTO {
	return &api.CRDTO{
		CRNumber:       cr.Number,
		Title:          cr.Title,
		Description:    cr.Description,
		State:          string(cr.State),
		SourceBranch:   cr.SourceBranch,
		TargetBranch:   cr.TargetBranch,
		AuthorName:     cr.Author.Name,
		AuthorUsername: cr.Author.Username,
		WebURL:         cr.WebURL,
		MergeStatus:    cr.MergeStatus,
		Labels:         cr.Labels,
		CreatedAt:      cr.CreatedAt,
		UpdatedAt:      cr.UpdatedAt,
	}
}

func toCRDTO(cr *po.ChangeRequest) *api.CRDTO {
	return &api.CRDTO{
		ID:             cr.ID,
		RepoID:         cr.RepoID,
		ProviderID:     cr.ProviderConfigID,
		CRNumber:       cr.CRNumber,
		Title:          cr.Title,
		Description:    cr.Description,
		State:          cr.State,
		SourceBranch:   cr.SourceBranch,
		TargetBranch:   cr.TargetBranch,
		AuthorName:     cr.AuthorName,
		AuthorUsername: cr.AuthorUsername,
		WebURL:         cr.WebURL,
		MergeStatus:    cr.MergeStatus,
		Labels:         cr.Labels,
		CreatedAt:      cr.CreatedAt,
		UpdatedAt:      cr.UpdatedAt,
		MergedAt:       cr.MergedAt,
	}
}

func resolveProvider(providerID uint) (provider.Provider, error) {
	p, err := provider.GetManager().GetProvider(providerID)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}
	return p, nil
}

func ListCRsByProvider(ctx context.Context, providerIDStr, owner, repoName, state string, page, perPage int) ([]api.CRDTO, int, error) {
	providerID := uint(0)
	fmt.Sscanf(providerIDStr, "%d", &providerID)
	p, err := resolveProvider(providerID)
	if err != nil {
		return nil, 0, err
	}
	crs, total, err := p.ListCRs(ctx, provider.ListCROptions{
		Owner: owner, Repo: repoName, State: provider.CRState(state), Page: page, PerPage: perPage,
	})
	if err != nil {
		return nil, 0, err
	}
	dtos := make([]api.CRDTO, 0, len(crs))
	for _, cr := range crs {
		dtos = append(dtos, *platformCRToAPI(cr))
	}
	return dtos, total, nil
}

func CreateCRByProvider(ctx context.Context, req *api.CreateCRByProviderReq) (*api.CRDTO, error) {
	p, err := resolveProvider(req.ProviderID)
	if err != nil {
		return nil, err
	}
	cr, err := p.CreateCR(ctx, provider.CreateCROptions{
		Owner: req.Owner, Repo: req.Repo, Title: req.Title, Description: req.Description,
		SourceBranch: req.SourceBranch, TargetBranch: req.TargetBranch,
		Labels: req.Labels, RemoveSourceBranch: req.RemoveSourceBranch,
	})
	if err != nil {
		return nil, err
	}
	return platformCRToAPI(cr), nil
}

func MergeCRByProvider(ctx context.Context, providerID uint, owner, repoName string, crNumber int, mergeMsg string, squash, removeBranch bool) (*api.CRDTO, error) {
	p, err := resolveProvider(providerID)
	if err != nil {
		return nil, err
	}
	cr, err := p.MergeCR(ctx, owner, repoName, crNumber, provider.MergeCROptions{
		MergeCommitMessage: mergeMsg, Squash: squash, RemoveSourceBranch: removeBranch,
	})
	if err != nil {
		return nil, err
	}
	return platformCRToAPI(cr), nil
}

func CloseCRByProvider(ctx context.Context, providerID uint, owner, repoName string, crNumber int) (*api.CRDTO, error) {
	p, err := resolveProvider(providerID)
	if err != nil {
		return nil, err
	}
	cr, err := p.CloseCR(ctx, owner, repoName, crNumber)
	if err != nil {
		return nil, err
	}
	return platformCRToAPI(cr), nil
}
