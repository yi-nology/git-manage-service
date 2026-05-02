package repo

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/biz/service/stats"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

func List(ctx context.Context, c *app.RequestContext) {
	repos, err := db.NewRepoDAO().FindAll()
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	var dtos []api.RepoDTO
	for _, r := range repos {
		dtos = append(dtos, api.NewRepoDTO(r))
	}
	response.Success(c, dtos)
}

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
	response.Success(c, api.NewRepoDTO(*repo))
}

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

	response.Success(c, api.NewRepoDTO(repo))
}

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

	if req.Path != repo.Path {
		gitSvc := git.NewGitService()
		if !gitSvc.IsGitRepo(req.Path) {
			response.BadRequest(c, "path is not a valid git repository")
			return
		}
	}

	if len(req.Remotes) > 0 {
		gitSvc := git.NewGitService()
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

	repo.Name = req.Name
	repo.Path = req.Path
	repo.RemoteURL = req.RemoteURL
	repo.AuthType = req.AuthType
	repo.AuthKey = req.AuthKey
	repo.AuthSecret = req.AuthSecret
	repo.RemoteAuths = req.RemoteAuths
	repo.DefaultCredentialID = req.DefaultCredentialID
	repo.RemoteCredentials = req.RemoteCredentials

	if err := repoDAO.Save(repo); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, api.NewRepoDTO(*repo))
}

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

	count, _ := db.NewSyncTaskDAO().CountByRepoKey(repo.Key)
	if count > 0 {
		response.BadRequest(c, "cannot delete repo used in sync tasks")
		return
	}

	if err := repoDAO.DeleteWithBindings(repo); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, map[string]string{"message": "deleted"})
}
