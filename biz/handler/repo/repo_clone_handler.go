package repo

import (
	"context"
	"os"
	"path/filepath"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/auth"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/biz/service/stats"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

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
				err = gitSvc.CloneWithAuthMethod(req.RemoteURL, req.LocalPath, authMethod, progressChan)
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
			err = gitSvc.CloneWithAuthMethod(req.RemoteURL, req.LocalPath, authMethod, progressChan)
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

	useDBKey := false
	var dbPrivateKey, dbPassphrase string

	if repo.DefaultCredentialID > 0 && authSvc.IsCredentialDBKey(repo.DefaultCredentialID) {
		pk, pp, err := authSvc.GetCredentialKeyContent(repo.DefaultCredentialID)
		if err != nil {
			response.InternalServerError(c, "failed to resolve credential key: "+err.Error())
			return
		}
		dbPrivateKey = pk
		dbPassphrase = pp
		useDBKey = true
	}

	if !useDBKey && repo.RemoteAuths != nil {
		for _, authInfo := range repo.RemoteAuths {
			if authInfo.Type == "ssh" && authInfo.Source == "database" && authInfo.SSHKeyID > 0 {
				pk, pp, err := authSvc.GetDBSSHKeyContent(authInfo.SSHKeyID)
				if err != nil {
					response.InternalServerError(c, "failed to resolve SSH key: "+err.Error())
					return
				}
				dbPrivateKey = pk
				dbPassphrase = pp
				useDBKey = true
				break
			}
		}
	}

	var fetchErr error
	if useDBKey {
		fetchErr = gitSvc.FetchAllWithDBKey(repo.Path, dbPrivateKey, dbPassphrase)
	} else {
		fetchErr = gitSvc.FetchAll(repo.Path)
	}
	if fetchErr != nil {
		response.InternalServerError(c, fetchErr.Error())
		return
	}

	response.Success(c, map[string]string{"message": "fetched"})
}

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
