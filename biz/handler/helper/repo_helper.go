package helper

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/constants"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

// GetRepoFromQuery 从 Query 参数获取仓库信息
// 返回 (*po.Repo, bool) - 第二个返回值表示是否成功获取
func GetRepoFromQuery(c *app.RequestContext) (*po.Repo, bool) {
	repoKey := c.Query("repo_key")
	if repoKey == "" {
		response.BadRequest(c, "repo_key is required")
		return nil, false
	}
	repo, err := db.NewRepoDAO().FindByKey(repoKey)
	if err != nil {
		response.NotFound(c, "repository not found")
		return nil, false
	}
	return repo, true
}

// GetRepoFromContext 从 Context 获取已验证的仓库（由中间件注入）
func GetRepoFromContext(c *app.RequestContext) *po.Repo {
	if repo, exists := c.Get(constants.ContextKeyRepo); exists {
		if r, ok := repo.(*po.Repo); ok {
			return r
		}
	}
	return nil
}

// GetRepoKeyFromQuery 仅获取 repo_key 参数
func GetRepoKeyFromQuery(c *app.RequestContext) (string, bool) {
	repoKey := c.Query("repo_key")
	if repoKey == "" {
		response.BadRequest(c, "repo_key is required")
		return "", false
	}
	return repoKey, true
}

// MustBindAndValidate 绑定并验证请求参数
// 返回 bool 表示是否成功，失败时已自动返回错误响应
func MustBindAndValidate[T any](c *app.RequestContext) (*T, bool) {
	var req T
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return nil, false
	}
	return &req, true
}

// BindJSON 绑定 JSON 请求体
func BindJSON[T any](c *app.RequestContext) (*T, bool) {
	var req T
	if err := c.BindJSON(&req); err != nil {
		response.BadRequest(c, "invalid JSON: "+err.Error())
		return nil, false
	}
	return &req, true
}
