package repo

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/biz/model/domain"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/biz/service/stats"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

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

func findRemote(remotes []domain.GitRemote, name string) *domain.GitRemote {
	for i := range remotes {
		if remotes[i].Name == name {
			return &remotes[i]
		}
	}
	return nil
}
