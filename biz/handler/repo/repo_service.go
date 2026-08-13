package repo

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	repoModel "github.com/yi-nology/git-manage-service/biz/model/repo"
	"github.com/yi-nology/git-manage-service/biz/service/auth"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/biz/service/stats"
	syncv2 "github.com/yi-nology/git-manage-service/biz/service/sync/v2"
	"github.com/yi-nology/git-manage-service/pkg/response"
	"github.com/yi-nology/git-platform-sdk/gitbackend"
)

func toProtoRepo(r po.Repo) *repoModel.RepoDTO {
	return &repoModel.RepoDTO{
		Id:                  uint64(r.ID),
		Key:                 r.Key,
		Name:                r.Name,
		Path:                r.Path,
		RemoteUrl:           r.RemoteURL,
		AuthType:            r.AuthType,
		AuthKey:             r.AuthKey,
		AuthSecret:          r.AuthSecret,
		DefaultCredentialId: uint64(r.DefaultCredentialID),
		ProviderConfigId:    uint64(r.ProviderConfigID),
		PlatformRepoId:      r.PlatformRepoID,
		PlatformOwner:       r.PlatformOwner,
		PlatformRepo:        r.PlatformRepo,
		CreatedAt:           r.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:           r.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func repoNameFromURL(remoteURL string) string {
	name := remoteURL
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	} else if idx := strings.LastIndex(name, ":"); idx >= 0 {
		name = name[idx+1:]
	}
	name = strings.TrimSuffix(name, ".git")
	name = strings.TrimSpace(name)
	return name
}

func findRemote(remotes []domain.GitRemote, name string) *domain.GitRemote {
	for i := range remotes {
		if remotes[i].Name == name {
			return &remotes[i]
		}
	}
	return nil
}

// List .
// @router /api/v1/repo/list [GET]
func List(ctx context.Context, c *app.RequestContext) {
	repos, err := db.NewRepoDAO().FindAll()
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	var dtos []*repoModel.RepoDTO
	for _, r := range repos {
		dtos = append(dtos, toProtoRepo(r))
	}
	response.Success(c, dtos)
}

// Get .
// @router /api/v1/repo/detail [GET]
func Get(ctx context.Context, c *app.RequestContext) {
	key := c.Query("key")
	if key == "" {
		response.BadRequest(c, "key is required")
		return
	}
	repo, err := db.NewRepoDAO().FindByKey(key)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}
	response.Success(c, toProtoRepo(*repo))
}

// Create .
// @router /api/v1/repo/create [POST]
func Create(ctx context.Context, c *app.RequestContext) {
	var req api.RegisterRepoReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	gitSvc := git.NewGitService()
	if !gitSvc.IsGitRepo(req.Path) {
		response.BadRequest(c, "path is not a valid git repository")
		return
	}

	if len(req.Remotes) > 0 {
		existingConfig, err := gitSvc.GetRepoConfig(req.Path)
		if err == nil {
			for _, existing := range existingConfig.Remotes {
				found := false
				for _, r := range req.Remotes {
					if r.Name == existing.Name {
						found = true
						break
					}
				}
				if !found {
					gitSvc.RemoveRemote(req.Path, existing.Name)
				}
			}

			for _, r := range req.Remotes {
				gitSvc.RemoveRemote(req.Path, r.Name)
				if err := gitSvc.AddRemote(req.Path, r.Name, r.FetchURL, r.IsMirror); err != nil {
				}
				if r.PushURL != "" && r.PushURL != r.FetchURL {
					gitSvc.SetRemotePushURL(req.Path, r.Name, r.PushURL)
				}
			}
		}
	}

	repo := po.Repo{
		Key:                 uuid.New().String(),
		Name:                req.Name,
		Path:                req.Path,
		RemoteURL:           req.RemoteURL,
		AuthType:            req.AuthType,
		AuthKey:             req.AuthKey,
		AuthSecret:          req.AuthSecret,
		RemoteAuths:         req.RemoteAuths,
		DefaultCredentialID: req.DefaultCredentialID,
		RemoteCredentials:   req.RemoteCredentials,
	}
	if err := db.NewRepoDAO().Create(&repo); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	go func() {
		head, err := gitSvc.GetHeadBranch(repo.Path)
		if err == nil && head != "" {
			stats.StatsSvc.SyncRepoStats(repo.ID, repo.Path, head)
		}
	}()

	response.Success(c, toProtoRepo(repo))
}

// Update .
// @router /api/v1/repo/update [POST]
func Update(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Key string `json:"key"`
		api.RegisterRepoReq
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repoDAO := db.NewRepoDAO()
	repo, err := repoDAO.FindByKey(req.Key)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	// Partial update: only validate/apply a new path when one is provided.
	// Treating an empty path as "no change" avoids accidentally pointing the
	// repo at the process working directory and overwriting fields the caller
	// did not send.
	if req.Path != "" && req.Path != repo.Path {
		gitSvc := git.NewGitService()
		if !gitSvc.IsGitRepo(req.Path) {
			response.BadRequest(c, "path is not a valid git repository")
			return
		}
	}

	repoPath := repo.Path
	if req.Path != "" {
		repoPath = req.Path
	}

	if len(req.Remotes) > 0 {
		gitSvc := git.NewGitService()
		existingConfig, err := gitSvc.GetRepoConfig(repoPath)
		if err == nil {
			for _, existing := range existingConfig.Remotes {
				found := false
				for _, r := range req.Remotes {
					if r.Name == existing.Name {
						found = true
						break
					}
				}
				if !found {
					gitSvc.RemoveRemote(repoPath, existing.Name)
				}
			}

			for _, r := range req.Remotes {
				gitSvc.RemoveRemote(repoPath, r.Name)
				if err := gitSvc.AddRemote(repoPath, r.Name, r.FetchURL, r.IsMirror); err != nil {
				}
				if r.PushURL != "" && r.PushURL != r.FetchURL {
					gitSvc.SetRemotePushURL(repoPath, r.Name, r.PushURL)
				}
			}
		}
	}

	// Apply only the fields that were actually provided.
	if req.Name != "" {
		repo.Name = req.Name
	}
	if req.Path != "" {
		repo.Path = req.Path
	}
	if req.RemoteURL != "" {
		repo.RemoteURL = req.RemoteURL
	}
	if req.AuthType != "" {
		repo.AuthType = req.AuthType
	}
	if req.AuthKey != "" {
		repo.AuthKey = req.AuthKey
	}
	if req.AuthSecret != "" {
		repo.AuthSecret = req.AuthSecret
	}
	if req.RemoteAuths != nil {
		repo.RemoteAuths = req.RemoteAuths
	}
	if req.DefaultCredentialID != 0 {
		repo.DefaultCredentialID = req.DefaultCredentialID
	}
	if req.RemoteCredentials != nil {
		repo.RemoteCredentials = req.RemoteCredentials
	}

	if err := repoDAO.Save(repo); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, toProtoRepo(*repo))
}

// Delete .
// @router /api/v1/repo/delete [POST]
func Delete(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Key string `json:"key"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repoDAO := db.NewRepoDAO()
	repo, err := repoDAO.FindByKey(req.Key)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	// git-sync-service is the source of truth for sync tasks (the local
	// sync_tasks table is an abandoned orphan that's never populated). Only
	// enforce the guard when the sync service is initialized to avoid nil calls.
	if syncSvc := syncv2.GetService(); syncSvc.GetCore() != nil {
		if tasks, _ := syncSvc.ListTasks(ctx, repo.Key); len(tasks) > 0 {
			response.BadRequest(c, "cannot delete repo used in sync tasks")
			return
		}
	}

	if err := repoDAO.DeleteWithBindings(repo); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "deleted"})
}

// Scan .
// @router /api/v1/repo/scan [POST]
func Scan(ctx context.Context, c *app.RequestContext) {
	var req api.ScanRepoReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	gitSvc := git.NewGitService()
	if !gitSvc.IsGitRepo(req.Path) {
		response.BadRequest(c, "path is not a valid git repository")
		return
	}

	config, err := gitSvc.GetRepoConfig(req.Path)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	c.Set("audit_details", map[string]string{"path": req.Path})
	response.Success(c, config)
}

// ScanDirectory .
// @router /api/v1/repo/scan-directory [POST]
func ScanDirectory(ctx context.Context, c *app.RequestContext) {
	var req api.ScanDirectoryReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Path == "" {
		response.BadRequest(c, "path is required")
		return
	}

	if _, err := os.Stat(req.Path); os.IsNotExist(err) {
		response.BadRequest(c, "path does not exist")
		return
	}

	depth := req.Depth
	if depth <= 0 {
		depth = 2
	}

	gitSvc := git.NewGitService()
	var repos []api.ScannedRepo

	if gitSvc.IsGitRepo(req.Path) {
		name := filepath.Base(req.Path)
		repo := api.ScannedRepo{
			Name:    name,
			Path:    req.Path,
			Remotes: []domain.GitRemote{},
		}

		config, err := gitSvc.GetRepoConfig(req.Path)
		if err == nil {
			repo.Remotes = config.Remotes
		}

		branch, _ := gitSvc.GetHeadBranch(req.Path)
		repo.CurrentBranch = branch

		if branch != "" {
			if hash, err := gitSvc.GetCommitHash(req.Path, "", branch); err == nil {
				if len(hash) > 7 {
					repo.LastCommit = hash[:7]
				} else {
					repo.LastCommit = hash
				}
			}
		}

		status, _ := gitSvc.GetStatus(req.Path)
		repo.HasChanges = status != ""

		repos = append(repos, repo)
	}

	var scan func(path string, currentDepth int)
	scan = func(path string, currentDepth int) {
		if currentDepth > depth {
			return
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			return
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			name := entry.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
				continue
			}

			subPath := filepath.Join(path, name)

			if gitSvc.IsGitRepo(subPath) {
				repo := api.ScannedRepo{
					Name:    name,
					Path:    subPath,
					Remotes: []domain.GitRemote{},
				}

				config, err := gitSvc.GetRepoConfig(subPath)
				if err == nil {
					repo.Remotes = config.Remotes
				}

				branch, _ := gitSvc.GetHeadBranch(subPath)
				repo.CurrentBranch = branch

				if branch != "" {
					if hash, err := gitSvc.GetCommitHash(subPath, "", branch); err == nil {
						if len(hash) > 7 {
							repo.LastCommit = hash[:7]
						} else {
							repo.LastCommit = hash
						}
					}
				}

				status, _ := gitSvc.GetStatus(subPath)
				repo.HasChanges = status != ""

				repos = append(repos, repo)
			} else if req.Recursive || currentDepth < depth {
				scan(subPath, currentDepth+1)
			}
		}
	}

	if depth > 0 {
		scan(req.Path, 1)
	}

	c.Set("audit_details", map[string]string{"path": req.Path})
	response.Success(c, api.ScanDirectoryResp{
		Repos: repos,
		Total: len(repos),
	})
}

// BatchCreate .
// @router /api/v1/repo/batch-create [POST]
func BatchCreate(ctx context.Context, c *app.RequestContext) {
	var req api.BatchCreateRepoReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if len(req.Repos) == 0 {
		response.BadRequest(c, "repos is required")
		return
	}

	gitSvc := git.NewGitService()
	repoDAO := db.NewRepoDAO()

	success := make([]api.RepoDTO, 0)
	failed := make([]api.BatchFailedItem, 0)

	for _, item := range req.Repos {
		if !gitSvc.IsGitRepo(item.Path) {
			failed = append(failed, api.BatchFailedItem{
				Name:   item.Name,
				Path:   item.Path,
				Reason: "not a valid git repository",
			})
			continue
		}

		existing, err := repoDAO.FindByPath(item.Path)
		if err == nil && existing != nil && existing.ID > 0 {
			failed = append(failed, api.BatchFailedItem{
				Name:   item.Name,
				Path:   item.Path,
				Reason: "repository already registered",
			})
			continue
		}

		config, err := gitSvc.GetRepoConfig(item.Path)
		remoteURL := ""
		if err == nil && len(config.Remotes) > 0 {
			origin := findRemote(config.Remotes, "origin")
			if origin != nil {
				remoteURL = origin.FetchURL
			} else {
				remoteURL = config.Remotes[0].FetchURL
			}
		}

		repo := po.Repo{
			Key:                 uuid.New().String(),
			Name:                item.Name,
			Path:                item.Path,
			RemoteURL:           remoteURL,
			DefaultCredentialID: item.DefaultCredentialID,
		}

		if err := repoDAO.Create(&repo); err != nil {
			failed = append(failed, api.BatchFailedItem{
				Name:   item.Name,
				Path:   item.Path,
				Reason: "failed to create: " + err.Error(),
			})
			continue
		}

		go func(r po.Repo) {
			head, err := gitSvc.GetHeadBranch(r.Path)
			if err == nil && head != "" {
				stats.StatsSvc.SyncRepoStats(r.ID, r.Path, head)
			}
		}(repo)

		success = append(success, api.NewRepoDTO(repo))
	}

	response.Success(c, api.BatchCreateRepoResp{
		Success: success,
		Failed:  failed,
	})
}

// Clone .
// @router /api/v1/repo/clone [POST]
func Clone(ctx context.Context, c *app.RequestContext) {
	var req api.CloneRepoReq
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	gitSvc := git.NewGitService()
	if info, err := os.Stat(req.LocalPath); err == nil && info.IsDir() {
		repoName := repoNameFromURL(req.RemoteURL)
		if repoName != "" {
			req.LocalPath = filepath.Join(req.LocalPath, repoName)
		}
	}
	if _, err := os.Stat(req.LocalPath); err == nil {
		if gitSvc.IsGitRepo(req.LocalPath) {
			response.BadRequest(c, "directory already contains a git repository")
			return
		}
	}

	taskID := uuid.New().String()
	git.GlobalTaskManager.AddTask(taskID)

	go func() {
		progressChan := make(chan string)

		go func() {
			for msg := range progressChan {
				git.GlobalTaskManager.AppendLog(taskID, msg)
			}
		}()

		var err error
		authSvc := auth.NewAuthService()

		skipTLS := false
		if req.ProviderConfigID > 0 {
			if pc, pcErr := db.NewProviderConfigDAO().FindByID(req.ProviderConfigID); pcErr == nil {
				skipTLS = pc.SkipTLS
			}
		}

		if req.CredentialID > 0 {
			if authSvc.IsCredentialDBKey(req.CredentialID) {
				privateKey, passphrase, keyErr := authSvc.GetCredentialKeyContent(req.CredentialID)
				if keyErr != nil {
					git.GlobalTaskManager.UpdateStatus(taskID, "failed", "Failed to load credential key: "+keyErr.Error())
					close(progressChan)
					return
				}
				git.GlobalTaskManager.AppendLog(taskID, "Using credential (DB SSH key) for clone...")
				err = gitSvc.CloneWithDBKey(req.RemoteURL, req.LocalPath, privateKey, passphrase, progressChan)
			} else {
				authMethod, authErr := authSvc.ResolveCredential(req.CredentialID)
				if authErr != nil {
					git.GlobalTaskManager.AppendLog(taskID, "Warning: credential resolution failed: "+authErr.Error())
				}
				git.GlobalTaskManager.AppendLog(taskID, "Using credential for clone...")
				err = gitSvc.CloneWithAuthMethod(req.RemoteURL, req.LocalPath, authMethod, progressChan, skipTLS)
			}
		} else if req.SSHKeyID > 0 {
			privateKey, passphrase, keyErr := authSvc.GetDBSSHKeyContent(req.SSHKeyID)
			if keyErr != nil {
				git.GlobalTaskManager.UpdateStatus(taskID, "failed", "Failed to load SSH key: "+keyErr.Error())
				close(progressChan)
				return
			}
			git.GlobalTaskManager.AppendLog(taskID, "Using database SSH key for clone...")
			err = gitSvc.CloneWithDBKey(req.RemoteURL, req.LocalPath, privateKey, passphrase, progressChan)
		} else {
			authMethod, authErr := authSvc.ResolveAuthFromParams(req.AuthType, req.AuthKey, req.AuthSecret, 0)
			if authErr != nil {
				git.GlobalTaskManager.AppendLog(taskID, "Warning: auth resolution failed: "+authErr.Error())
			}
			err = gitSvc.CloneWithAuthMethod(req.RemoteURL, req.LocalPath, authMethod, progressChan, skipTLS)
		}
		close(progressChan)

		if err != nil {
			git.GlobalTaskManager.UpdateStatus(taskID, "failed", err.Error())
			return
		}

		git.GlobalTaskManager.UpdateStatus(taskID, "success", "")

		name := filepath.Base(req.LocalPath)
		repoName := req.Name
		if repoName == "" {
			repoName = name
		}
		repo := po.Repo{
			Key:                 uuid.New().String(),
			Name:                repoName,
			Path:                req.LocalPath,
			RemoteURL:           req.RemoteURL,
			AuthType:            req.AuthType,
			AuthKey:             req.AuthKey,
			AuthSecret:          req.AuthSecret,
			DefaultCredentialID: req.CredentialID,
			ProviderConfigID:    req.ProviderConfigID,
			PlatformOwner:       req.PlatformOwner,
			PlatformRepo:        req.PlatformRepo,
		}
		if err := db.NewRepoDAO().Create(&repo); err != nil {
			git.GlobalTaskManager.AppendLog(taskID, "Warning: failed to persist repo record: "+err.Error())
			return
		}

		if repo.ProviderConfigID > 0 && repo.PlatformOwner != "" && repo.PlatformRepo != "" {
			bindingDAO := db.NewRepoProviderBindingDAO()
			binding := &po.RepoProviderBinding{
				RepoID:           repo.ID,
				ProviderConfigID: repo.ProviderConfigID,
				PlatformOwner:    repo.PlatformOwner,
				PlatformRepo:     repo.PlatformRepo,
				RemoteName:       "origin",
				IsPrimary:        true,
				Status:           "active",
			}
			if err := bindingDAO.Create(binding); err != nil {
				git.GlobalTaskManager.AppendLog(taskID, "Warning: failed to create provider binding: "+err.Error())
			}
		}

		go func() {
			head, err := gitSvc.GetHeadBranch(repo.Path)
			if err == nil && head != "" {
				stats.StatsSvc.SyncRepoStats(repo.ID, repo.Path, head)
			}
		}()
	}()

	c.Set("audit_skip", true)
	response.Success(c, map[string]string{"task_id": taskID})
}

// Fetch .
// @router /api/v1/repo/fetch [POST]
func Fetch(ctx context.Context, c *app.RequestContext) {
	var req struct {
		RepoKey string `json:"repo_key"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	repo, err := db.NewRepoDAO().FindByKey(req.RepoKey)
	if err != nil {
		response.NotFound(c, "repo not found")
		return
	}

	gitSvc := git.NewGitService()
	authSvc := auth.NewAuthService()

	remotes, err := gitSvc.GetRemotes(repo.Path)
	if err != nil {
		response.InternalServerError(c, "failed to list remotes: "+err.Error())
		return
	}

	var errors []string
	for _, remoteName := range remotes {
		remoteURL, urlErr := gitSvc.GetRemoteURL(repo.Path, remoteName)
		if urlErr != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to get URL: %v", remoteName, urlErr))
			continue
		}

		authMethod, isDBKey, resolveErr := authSvc.ResolveCredentialForRemote(
			repo.RemoteCredentials,
			repo.DefaultCredentialID,
			repo.RemoteAuths,
			remoteName,
			repo.AuthType, repo.AuthKey, repo.AuthSecret,
		)

		if resolveErr != nil {
			errors = append(errors, fmt.Sprintf("%s: failed to resolve auth: %v", remoteName, resolveErr))
			continue
		}

		var fetchErr error
		if isDBKey {
			credID := auth.GetCredentialIDForRemote(repo.RemoteCredentials, repo.DefaultCredentialID, remoteName)
			if credID > 0 {
				privateKey, passphrase, keyErr := authSvc.GetCredentialKeyContent(credID)
				if keyErr != nil {
					errors = append(errors, fmt.Sprintf("%s: failed to load SSH key: %v", remoteName, keyErr))
					continue
				}
				fetchErr = gitSvc.FetchWithDBKey(repo.Path, remoteURL, privateKey, passphrase, nil)
			} else {
				fetchErr = gitSvc.Fetch(repo.Path, remoteName, nil)
			}
		} else if authMethod.Type != gitbackend.AuthNone {
			fetchErr = gitSvc.Fetch(repo.Path, remoteName, nil)
		} else {
			fetchErr = gitSvc.Fetch(repo.Path, remoteName, nil)
		}

		if fetchErr != nil && fetchErr.Error() != "already up-to-date" {
			errors = append(errors, fmt.Sprintf("%s: %v", remoteName, fetchErr))
		}
	}

	if len(errors) > 0 {
		response.InternalServerError(c, strings.Join(errors, "; "))
		return
	}

	response.Success(c, map[string]string{"message": "fetched"})
}

// GetCloneTask .
// @router /api/v1/repo/task [GET]
func GetCloneTask(ctx context.Context, c *app.RequestContext) {
	id := c.Query("task_id")
	if id == "" {
		response.BadRequest(c, "task_id is required")
		return
	}
	task, ok := git.GlobalTaskManager.GetTask(id)
	if !ok {
		response.NotFound(c, "task not found")
		return
	}
	response.Success(c, task)
}
