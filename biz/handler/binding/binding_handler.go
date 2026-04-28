package binding

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	bindingsvc "github.com/yi-nology/git-manage-service/biz/service/binding"
	pkgresponse "github.com/yi-nology/git-manage-service/pkg/response"
)

func List(ctx context.Context, c *app.RequestContext) {
	var req api.ListBindingsReq
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	result, err := bindingsvc.ListBindings(req.RepoKey, req.ProviderConfigID)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to list bindings: "+err.Error())
		return
	}
	if result == nil {
		result = []api.RepoProviderBindingDTO{}
	}
	pkgresponse.Success(c, result)
}

func Get(ctx context.Context, c *app.RequestContext) {
	id, err := parseID(c)
	if err != nil {
		pkgresponse.BadRequest(c, "Invalid ID")
		return
	}

	result, err := bindingsvc.GetBinding(id)
	if err != nil {
		pkgresponse.NotFound(c, err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func Create(ctx context.Context, c *app.RequestContext) {
	var req api.CreateBindingReq
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	result, err := bindingsvc.CreateBinding(ctx, &req)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to create binding: "+err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func Update(ctx context.Context, c *app.RequestContext) {
	id, err := parseID(c)
	if err != nil {
		pkgresponse.BadRequest(c, "Invalid ID")
		return
	}

	var req api.UpdateBindingReq
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	result, err := bindingsvc.UpdateBinding(id, &req)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to update binding: "+err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func Delete(ctx context.Context, c *app.RequestContext) {
	id, err := parseID(c)
	if err != nil {
		pkgresponse.BadRequest(c, "Invalid ID")
		return
	}

	cleanup := c.Query("cleanup_webhook") == "true"
	if err := bindingsvc.DeleteBinding(ctx, id, cleanup); err != nil {
		pkgresponse.InternalServerError(c, "Failed to delete binding: "+err.Error())
		return
	}
	pkgresponse.Success(c, nil)
}

func SetPrimary(ctx context.Context, c *app.RequestContext) {
	id, err := parseID(c)
	if err != nil {
		pkgresponse.BadRequest(c, "Invalid ID")
		return
	}

	result, err := bindingsvc.SetPrimary(id)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to set primary: "+err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func AutoDetect(ctx context.Context, c *app.RequestContext) {
	var req api.AutoDetectReq
	if err := c.BindAndValidate(&req); err != nil {
		pkgresponse.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	result, err := bindingsvc.AutoDetect(req.RepoKey)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to auto-detect: "+err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func RegisterWebhook(ctx context.Context, c *app.RequestContext) {
	id, err := parseID(c)
	if err != nil {
		pkgresponse.BadRequest(c, "Invalid ID")
		return
	}

	result, err := bindingsvc.RegisterWebhook(ctx, id)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to register webhook: "+err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func DeleteWebhook(ctx context.Context, c *app.RequestContext) {
	id, err := parseID(c)
	if err != nil {
		pkgresponse.BadRequest(c, "Invalid ID")
		return
	}

	result, err := bindingsvc.DeleteWebhook(ctx, id)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to delete webhook: "+err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

func parseID(c *app.RequestContext) (uint, error) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
