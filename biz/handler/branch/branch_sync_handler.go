package branch

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/branch"
	"github.com/yi-nology/git-manage-service/biz/service/auth"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func Push(ctx context.Context, c *app.RequestContext) {
	var req branch.PushBranchRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	authSvc := auth.NewAuthService()

	var errors []string
	for _, remote := range req.GetRemotes() {
		authMethod, isDBKey, resolveErr := authSvc.ResolveCredentialForRemote(
			repo.RemoteCredentials,
			repo.DefaultCredentialID,
			repo.RemoteAuths,
			remote,
			repo.AuthType, repo.AuthKey, repo.AuthSecret,
		)

		if resolveErr != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to resolve auth: %v", remote, resolveErr))
			continue
		}

		if isDBKey {
			credID := auth.GetCredentialIDForRemote(repo.RemoteCredentials, repo.DefaultCredentialID, remote)
			if credID > 0 {
				privateKey, passphrase, keyErr := authSvc.GetCredentialKeyContent(credID)
				if keyErr != nil {
					errors = append(errors, fmt.Sprintf("%s: failed to load SSH key: %v", remote, keyErr))
					continue
				}
				if err := gitSvc.PushBranchWithDBKey(repo.Path, remote, req.GetName(), privateKey, passphrase); err != nil {
					errors = append(errors, fmt.Sprintf("%s: %v", remote, err))
				}
			} else {
				errors = append(errors, fmt.Sprintf("%s: no credential configured", remote))
			}
		} else if authMethod != nil {
			if err := gitSvc.PushBranchWithAuth(repo.Path, remote, req.GetName(), authMethod); err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", remote, err))
			}
		} else {
			if err := gitSvc.PushBranch(repo.Path, remote, req.GetName()); err != nil {
				errors = append(errors, fmt.Sprintf("%s: %v", remote, err))
			}
		}
	}

	if len(errors) > 0 {
		response.InternalServerError(c, strings.Join(errors, "; "))
		return
	}

	response.Success(c, map[string]string{"message": "pushed"})
}

func Pull(ctx context.Context, c *app.RequestContext) {
	var req branch.PullBranchRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.GetRepoKey())
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	authSvc := auth.NewAuthService()
	branches, _ := gitSvc.ListBranchesWithInfo(repo.Path)

	var isCurrent bool
	var upstreamRemote string
	var remoteBranch string

	for _, b := range branches {
		if b.Name == req.GetName() {
			isCurrent = b.IsCurrent
			if b.Upstream != "" {
				parts := strings.Split(b.Upstream, "/")
				if len(parts) > 0 {
					upstreamRemote = parts[0]
					if len(parts) > 1 {
						remoteBranch = strings.Join(parts[1:], "/")
					}
				}
			}
			break
		}
	}

	if upstreamRemote == "" {
		response.BadRequest(c, "No upstream configured for this branch")
		return
	}

	if remoteBranch == "" {
		remoteBranch = req.GetName()
	}

	authMethod, isDBKey, resolveErr := authSvc.ResolveCredentialForRemote(
		repo.RemoteCredentials,
		repo.DefaultCredentialID,
		repo.RemoteAuths,
		upstreamRemote,
		repo.AuthType, repo.AuthKey, repo.AuthSecret,
	)
	if resolveErr != nil {
		response.InternalServerError(c, fmt.Sprintf("failed to resolve auth: %v", resolveErr))
		return
	}

	var privateKey, passphrase string
	if isDBKey {
		credID := auth.GetCredentialIDForRemote(repo.RemoteCredentials, repo.DefaultCredentialID, upstreamRemote)
		if credID > 0 {
			var keyErr error
			privateKey, passphrase, keyErr = authSvc.GetCredentialKeyContent(credID)
			if keyErr != nil {
				response.InternalServerError(c, fmt.Sprintf("failed to load SSH key: %v", keyErr))
				return
			}
		}
	}

	if !isCurrent {
		var updateErr error
		if isDBKey {
			updateErr = gitSvc.FetchBranchWithDBKey(repo.Path, upstreamRemote, remoteBranch, privateKey, passphrase)
		} else {
			updateErr = gitSvc.UpdateBranchFastForward(repo.Path, upstreamRemote, req.GetName(), remoteBranch)
		}
		if updateErr != nil {
			response.InternalServerError(c, fmt.Sprintf("Update failed (must be fast-forward): %v", updateErr))
			return
		}

		response.Success(c, map[string]string{"message": "updated (fast-forward)"})
		return
	}

	var pullErr error
	if isDBKey {
		pullErr = gitSvc.PullBranchWithDBKey(repo.Path, upstreamRemote, req.GetName(), privateKey, passphrase)
	} else if authMethod != nil {
		pullErr = gitSvc.PullBranchWithAuth(repo.Path, upstreamRemote, req.GetName(), authMethod)
	} else {
		pullErr = gitSvc.PullBranch(repo.Path, upstreamRemote, req.GetName())
	}
	if pullErr != nil {
		response.InternalServerError(c, pullErr.Error())
		return
	}

	response.Success(c, map[string]string{"message": "synced"})
}
