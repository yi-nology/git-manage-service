// Package handler 提供泛型 HTTP handler 包装器，消除每个 endpoint 中
// BindAndValidate → 错误处理 → repo 查找 → 业务调用 → 响应 的重复代码。
//
// 使用示例:
//
//	func ListStash(ctx context.Context, c *app.RequestContext) {
//	    handler.DoWithRepo(c,
//	        func(r *stash.ListStashRequest) string { return r.RepoKey },
//	        func(repo *po.Repo, req *stash.ListStashRequest) (any, error) {
//	            return git.NewGitService().StashList(repo.Path)
//	        },
//	    )
//	}
//
// 注意: Req 使用指针 (*Req) 因为 proto 生成的结构体包含 sync.Mutex，不可值拷贝。
package handler

import (
	"errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/po"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

// BizError 业务错误，携带 HTTP 语义状态码。
type BizError struct {
	code int
	msg  string
}

func (e *BizError) Error() string { return e.msg }

func ErrBadRequest(msg string) *BizError { return &BizError{code: 400, msg: msg} }
func ErrNotFound(msg string) *BizError   { return &BizError{code: 404, msg: msg} }
func ErrInternal(msg string) *BizError   { return &BizError{code: 500, msg: msg} }

func respondError(c *app.RequestContext, err error) {
	var biz *BizError
	if errors.As(err, &biz) {
		switch biz.code {
		case 400:
			response.BadRequest(c, biz.msg)
		case 404:
			response.NotFound(c, biz.msg)
		default:
			response.InternalServerError(c, biz.msg)
		}
		return
	}
	response.InternalError(c, err)
}

// BindAndDo 绑定请求 → 业务逻辑 → 响应。
// Proto 结构体使用指针避免值拷贝 Mutex。
func BindAndDo[Req any, Resp any](c *app.RequestContext, fn func(req *Req) (Resp, error)) {
	var req Req
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := fn(&req)
	if err != nil {
		respondError(c, err)
		return
	}
	response.Success(c, resp)
}

// Do 绑定请求并执行无返回数据的业务操作。
func Do[Req any](c *app.RequestContext, fn func(req *Req) error) {
	var req Req
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := fn(&req); err != nil {
		respondError(c, err)
		return
	}
	response.Success(c, nil)
}

// DoWithRepo 绑定请求 → 提取 repo_key → 查找仓库 → 业务逻辑 → 响应。
func DoWithRepo[Req any, Resp any](
	c *app.RequestContext,
	getRepoKey func(*Req) string,
	fn func(repo *po.Repo, req *Req) (Resp, error),
) {
	BindAndDo(c, func(req *Req) (Resp, error) {
		repoKey := getRepoKey(req)
		if repoKey == "" {
			var zero Resp
			return zero, ErrBadRequest("repo_key is required")
		}
		repo, err := db.NewRepoDAO().FindByKey(repoKey)
		if err != nil {
			var zero Resp
			return zero, ErrNotFound("repo not found")
		}
		return fn(repo, req)
	})
}

// DoWithRepoVoid 与 DoWithRepo 相同但无返回数据。
func DoWithRepoVoid[Req any](
	c *app.RequestContext,
	getRepoKey func(*Req) string,
	fn func(repo *po.Repo, req *Req) error,
) {
	Do(c, func(req *Req) error {
		repoKey := getRepoKey(req)
		if repoKey == "" {
			return ErrBadRequest("repo_key is required")
		}
		repo, err := db.NewRepoDAO().FindByKey(repoKey)
		if err != nil {
			return ErrNotFound("repo not found")
		}
		return fn(repo, req)
	})
}

// DoWithQueryRepo 从 query param "repo_key" 获取仓库 → 业务逻辑 → 响应。
func DoWithQueryRepo[Resp any](c *app.RequestContext, fn func(repo *po.Repo) (Resp, error)) {
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
	resp, err := fn(repo)
	if err != nil {
		respondError(c, err)
		return
	}
	response.Success(c, resp)
}
