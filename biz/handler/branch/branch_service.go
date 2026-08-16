package branch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/branch"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/auth"
	"github.com/yi-nology/git-manage-service/biz/service/branchrule"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/pkg/handler"
	"github.com/yi-nology/git-manage-service/pkg/response"
	"github.com/yi-nology/git-platform-sdk/gitbackend"
)

// List .
// @router /api/v1/branch/list [GET]
func List(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepo(c,
		func(req *branch.ListBranchRequest) string { return req.GetRepoKey() },
		func(repo *po.Repo, req *branch.ListBranchRequest) (map[string]interface{}, error) {
			gitSvc := git.NewGitService()
			branches, err := gitSvc.ListBranchesWithInfo(repo.Path)
			if err != nil {
				return nil, handler.ErrInternal(err.Error())
			}

			branchTypeFilter := c.Query("type")
			if branchTypeFilter != "" && (branchTypeFilter == "local" || branchTypeFilter == "remote") {
				var filtered []domain.BranchInfo
				for _, b := range branches {
					if b.Type == branchTypeFilter {
						filtered = append(filtered, b)
					}
				}
				branches = filtered
			}

			if req.GetKeyword() != "" {
				var filtered []domain.BranchInfo
				keyword := strings.ToLower(req.GetKeyword())
				for _, b := range branches {
					if strings.Contains(strings.ToLower(b.Name), keyword) ||
						strings.Contains(strings.ToLower(b.Author), keyword) {
						filtered = append(filtered, b)
					}
				}
				branches = filtered
			}

			page := int(req.GetPage())
			if page < 1 {
				page = 1
			}
			pageSize := int(req.GetPageSize())
			if pageSize < 1 {
				pageSize = 100
			}

			start := (page - 1) * pageSize
			end := start + pageSize
			if start > len(branches) {
				start = len(branches)
			}
			if end > len(branches) {
				end = len(branches)
			}

			paged := branches[start:end]

			for i := range paged {
				b := &paged[i]
				if b.Upstream != "" {
					ahead, behind, _ := gitSvc.GetBranchSyncStatus(repo.Path, b.Name, b.Upstream)
					b.Ahead = ahead
					b.Behind = behind
				}
			}

			return map[string]interface{}{
				"total": len(branches),
				"list":  paged,
			}, nil
		},
	)
}

// Create .
// @router /api/v1/branch/create [POST]
func Create(ctx context.Context, c *app.RequestContext) {
	skipRules := string(c.GetHeader("X-Skip-Branch-Rules")) == "true"
	handler.DoWithRepoVoid(c,
		func(req *branch.CreateBranchRequest) string { return req.RepoKey },
		func(repo *po.Repo, req *branch.CreateBranchRequest) error {
			validation, err := branchrule.ValidateBranchName(req.RepoKey, req.GetName(), req.GetBaseRef(), skipRules)
			if err == nil && !validation.Valid {
				return handler.ErrBadRequest(fmt.Sprintf("分支规则校验失败: %s", validation.Message))
			}

			gitSvc := git.NewGitService()
			if err := gitSvc.CreateBranch(repo.Path, req.GetName(), req.GetBaseRef()); err != nil {
				return handler.ErrInternal(err.Error())
			}
			return nil
		},
	)
}

// Delete .
// @router /api/v1/branch/delete [POST]
func Delete(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepoVoid(c,
		func(req *branch.DeleteBranchRequest) string { return req.RepoKey },
		func(repo *po.Repo, req *branch.DeleteBranchRequest) error {
			gitSvc := git.NewGitService()
			if err := gitSvc.DeleteBranch(repo.Path, req.GetName(), req.GetForce()); err != nil {
				return handler.ErrInternal(err.Error())
			}
			return nil
		},
	)
}

// Update .
// @router /api/v1/branch/update [POST]
func Update(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepoVoid(c,
		func(req *branch.UpdateBranchRequest) string { return req.RepoKey },
		func(repo *po.Repo, req *branch.UpdateBranchRequest) error {
			gitSvc := git.NewGitService()
			currentName := req.GetName()

			if req.GetNewName() != "" && req.GetNewName() != currentName {
				if err := gitSvc.RenameBranch(repo.Path, currentName, req.GetNewName()); err != nil {
					return handler.ErrInternal(err.Error())
				}
				currentName = req.GetNewName()
			}

			if req.GetDesc() != "" {
				if err := gitSvc.SetBranchDescription(repo.Path, currentName, req.GetDesc()); err != nil {
				}
			}
			return nil
		},
	)
}

// Checkout .
// @router /api/v1/branch/checkout [POST]
func Checkout(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepoVoid(c,
		func(req *branch.CheckoutBranchRequest) string { return req.RepoKey },
		func(repo *po.Repo, req *branch.CheckoutBranchRequest) error {
			gitSvc := git.NewGitService()
			if err := gitSvc.CheckoutBranch(repo.Path, req.GetName()); err != nil {
				return handler.ErrBadRequest("Checkout failed: " + err.Error())
			}
			return nil
		},
	)
}

// Push .
// @router /api/v1/branch/push [POST]
func Push(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepo(c,
		func(req *branch.PushBranchRequest) string { return req.GetRepoKey() },
		func(repo *po.Repo, req *branch.PushBranchRequest) (map[string]string, error) {
			gitSvc := git.NewGitService()
			authSvc := auth.NewAuthService()

			var errs []string
			for _, remote := range req.GetRemotes() {
				authMethod, isDBKey, resolveErr := authSvc.ResolveCredentialForRemote(
					repo.RemoteCredentials,
					repo.DefaultCredentialID,
					nil,
					remote,
					"", "", "",
				)

				if resolveErr != nil {
					errs = append(errs, fmt.Sprintf("%s: failed to resolve auth: %v", remote, resolveErr))
					continue
				}

				if isDBKey {
					credID := auth.GetCredentialIDForRemote(repo.RemoteCredentials, repo.DefaultCredentialID, remote)
					if credID > 0 {
						privateKey, passphrase, keyErr := authSvc.GetCredentialKeyContent(credID)
						if keyErr != nil {
							errs = append(errs, fmt.Sprintf("%s: failed to load SSH key: %v", remote, keyErr))
							continue
						}
						if err := gitSvc.PushBranchWithDBKey(repo.Path, remote, req.GetName(), privateKey, passphrase); err != nil {
							errs = append(errs, fmt.Sprintf("%s: %v", remote, err))
						}
					} else {
						errs = append(errs, fmt.Sprintf("%s: no credential configured", remote))
					}
				} else if authMethod.Type != gitbackend.AuthNone {
					if err := gitSvc.PushBranchWithAuth(repo.Path, remote, req.GetName(), authMethod); err != nil {
						errs = append(errs, fmt.Sprintf("%s: %v", remote, err))
					}
				} else {
					if err := gitSvc.PushBranch(repo.Path, remote, req.GetName()); err != nil {
						errs = append(errs, fmt.Sprintf("%s: %v", remote, err))
					}
				}
			}

			if len(errs) > 0 {
				return nil, handler.ErrInternal(strings.Join(errs, "; "))
			}

			return map[string]string{"message": "pushed"}, nil
		},
	)
}

// Pull .
// @router /api/v1/branch/pull [POST]
func Pull(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepo(c,
		func(req *branch.PullBranchRequest) string { return req.GetRepoKey() },
		func(repo *po.Repo, req *branch.PullBranchRequest) (map[string]string, error) {
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
				return nil, handler.ErrBadRequest("No upstream configured for this branch")
			}

			if remoteBranch == "" {
				remoteBranch = req.GetName()
			}

			authMethod, isDBKey, resolveErr := authSvc.ResolveCredentialForRemote(
				repo.RemoteCredentials,
				repo.DefaultCredentialID,
				nil,
				upstreamRemote,
				"", "", "",
			)
			if resolveErr != nil {
				return nil, handler.ErrInternal(fmt.Sprintf("failed to resolve auth: %v", resolveErr))
			}

			var privateKey, passphrase string
			if isDBKey {
				credID := auth.GetCredentialIDForRemote(repo.RemoteCredentials, repo.DefaultCredentialID, upstreamRemote)
				if credID > 0 {
					var keyErr error
					privateKey, passphrase, keyErr = authSvc.GetCredentialKeyContent(credID)
					if keyErr != nil {
						return nil, handler.ErrInternal(fmt.Sprintf("failed to load SSH key: %v", keyErr))
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
					return nil, handler.ErrInternal(fmt.Sprintf("Update failed (must be fast-forward): %v", updateErr))
				}

				return map[string]string{"message": "updated (fast-forward)"}, nil
			}

			var pullErr error
			if isDBKey {
				pullErr = gitSvc.PullBranchWithDBKey(repo.Path, upstreamRemote, req.GetName(), privateKey, passphrase)
			} else if authMethod.Type != gitbackend.AuthNone {
				pullErr = gitSvc.PullBranchWithAuth(repo.Path, upstreamRemote, req.GetName(), authMethod)
			} else {
				pullErr = gitSvc.PullBranch(repo.Path, upstreamRemote, req.GetName())
			}
			if pullErr != nil {
				return nil, handler.ErrInternal(pullErr.Error())
			}

			return map[string]string{"message": "synced"}, nil
		},
	)
}

// Compare .
// @router /api/v1/branch/compare [GET]
func Compare(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepo(c,
		func(req *branch.CompareBranchRequest) string { return req.RepoKey },
		func(repo *po.Repo, req *branch.CompareBranchRequest) (map[string]interface{}, error) {
			gitSvc := git.NewGitService()

			stat, err := gitSvc.GetDiffStat(repo.Path, req.GetBase(), req.GetTarget())
			if err != nil {
				return nil, handler.ErrInternal(err.Error())
			}

			files, err := gitSvc.GetDiffFiles(repo.Path, req.GetBase(), req.GetTarget())
			if err != nil {
				return nil, handler.ErrInternal(err.Error())
			}

			return map[string]interface{}{
				"stat":  stat,
				"files": files,
			}, nil
		},
	)
}

// GetDiff .
// @router /api/v1/branch/diff [GET]
func GetDiff(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepo(c,
		func(req *branch.GetDiffRequest) string { return req.RepoKey },
		func(repo *po.Repo, req *branch.GetDiffRequest) (map[string]string, error) {
			gitSvc := git.NewGitService()
			content, err := gitSvc.GetRawDiff(repo.Path, req.GetBase(), req.GetTarget(), req.GetFile())
			if err != nil {
				return nil, handler.ErrInternal(err.Error())
			}
			return map[string]string{"diff": content}, nil
		},
	)
}

// GetPatch .
// @router /api/v1/branch/patch [GET]
func GetPatch(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepoRaw(c,
		func(req *branch.GetPatchRequest) string { return req.GetRepoKey() },
		func(repo *po.Repo, req *branch.GetPatchRequest) error {
			gitSvc := git.NewGitService()
			patch, err := gitSvc.GetPatch(repo.Path, req.GetBase(), req.GetTarget())
			if err != nil {
				return handler.ErrInternal(err.Error())
			}

			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s-%s-%s.patch", repo.Name, req.GetBase(), time.Now().Format("20060102")))
			c.Write([]byte(patch))
			return nil
		},
	)
}

// CherryPick .
// @router /api/v1/branch/cherry-pick [POST]
func CherryPick(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepoRaw(c,
		func(req *branch.CherryPickRequest) string { return req.GetRepoKey() },
		func(repo *po.Repo, req *branch.CherryPickRequest) error {
			gitSvc := git.NewGitService()
			newCommit, conflicts, err := gitSvc.CherryPick(repo.Path, req.GetCommitHash(), false)
			if err != nil {
				if len(conflicts) > 0 {
					c.JSON(200, response.Response{
						Code: 409,
						Msg:  "Cherry-pick conflict",
						Data: map[string]interface{}{
							"conflicts": conflicts,
						},
					})
					return nil
				}
				return handler.ErrInternal(err.Error())
			}

			response.Success(c, map[string]interface{}{
				"success":    true,
				"new_commit": newCommit,
			})
			return nil
		},
	)
}

// Rebase .
// @router /api/v1/branch/rebase [POST]
func Rebase(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepoRaw(c,
		func(req *branch.RebaseRequest) string { return req.GetRepoKey() },
		func(repo *po.Repo, req *branch.RebaseRequest) error {
			gitSvc := git.NewGitService()
			success, conflicts, err := gitSvc.Rebase(repo.Path, req.GetUpstream(), "")
			if err != nil {
				return handler.ErrInternal(err.Error())
			}

			if !success {
				c.JSON(200, response.Response{
					Code: 409,
					Msg:  "Rebase conflict",
					Data: map[string]interface{}{
						"success":     false,
						"in_progress": true,
						"conflicts":   conflicts,
					},
				})
				return nil
			}

			response.Success(c, map[string]interface{}{
				"success":     true,
				"in_progress": false,
			})
			return nil
		},
	)
}

// RebaseAbort .
// @router /api/v1/branch/rebase/abort [POST]
func RebaseAbort(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	if repoKey == "" {
		response.BadRequest(c, "repo_key is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	if err := gitSvc.RebaseAbort(repo.Path); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "rebase aborted"})
}

// RebaseContinue .
// @router /api/v1/branch/rebase/continue [POST]
func RebaseContinue(ctx context.Context, c *app.RequestContext) {
	repoKey := c.Query("repo_key")
	if repoKey == "" {
		response.BadRequest(c, "repo_key is required")
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	success, conflicts, err := gitSvc.RebaseContinue(repo.Path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	if !success {
		c.JSON(200, response.Response{
			Code: 409,
			Msg:  "Rebase conflict",
			Data: map[string]interface{}{
				"success":     false,
				"in_progress": true,
				"conflicts":   conflicts,
			},
		})
		return
	}

	response.Success(c, map[string]interface{}{
		"success":     true,
		"in_progress": false,
	})
}

// Merge .
// @router /api/v1/branch/merge [POST]
func Merge(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepoRaw(c,
		func(req *branch.MergeBranchRequest) string { return req.GetRepoKey() },
		func(repo *po.Repo, req *branch.MergeBranchRequest) error {
			gitSvc := git.NewGitService()

			check, err := gitSvc.MergeDryRun(repo.Path, req.GetSource(), req.GetTarget())
			if err != nil {
				return handler.ErrInternal("Pre-merge check failed: " + err.Error())
			}
			if !check.Success {
				mergeID := uuid.New().String()
				reportURL := fmt.Sprintf("/merge_report.html?repo_key=%s&source=%s&target=%s&merge_id=%s", repo.Key, req.GetSource(), req.GetTarget(), mergeID)

				c.JSON(200, response.Response{
					Code: 409,
					Msg:  "Merge conflict detected",
					Data: map[string]interface{}{
						"conflicts":  check.Conflicts,
						"report_url": reportURL,
						"merge_id":   mergeID,
					},
				})
				return nil
			}

			if err := gitSvc.Merge(repo.Path, req.GetSource(), req.GetTarget(), req.GetMessage(), req.GetNoFf(), req.GetSquash()); err != nil {
				return handler.ErrInternal("Merge execution failed: " + err.Error())
			}

			c.Set("audit_target", "repo:"+repo.Key)
			c.Set("audit_details", map[string]string{"source": req.GetSource(), "target": req.GetTarget()})

			response.Success(c, map[string]string{"status": "merged"})
			return nil
		},
	)
}

// MergeCheck .
// @router /api/v1/branch/merge/check [GET]
func MergeCheck(ctx context.Context, c *app.RequestContext) {
	handler.DoWithRepo(c,
		func(req *branch.MergeCheckRequest) string { return req.RepoKey },
		func(repo *po.Repo, req *branch.MergeCheckRequest) (*git.MergeResult, error) {
			gitSvc := git.NewGitService()
			result, err := gitSvc.MergeDryRun(repo.Path, req.GetBase(), req.GetTarget())
			if err != nil {
				return nil, handler.ErrInternal(err.Error())
			}
			return result, nil
		},
	)
}
