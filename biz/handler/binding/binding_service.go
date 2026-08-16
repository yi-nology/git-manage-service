package binding

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	bindingModel "github.com/yi-nology/git-manage-service/biz/model/binding"
	bindingsvc "github.com/yi-nology/git-manage-service/biz/service/binding"
	"github.com/yi-nology/git-manage-service/pkg/handler"
	pkgresponse "github.com/yi-nology/git-manage-service/pkg/response"
)

func convertToProtoBinding(dto api.RepoProviderBindingDTO) *bindingModel.BindingInfo {
	return &bindingModel.BindingInfo{
		Id:               uint64(dto.ID),
		RepoKey:          dto.RepoKey,
		ProviderConfigId: uint64(dto.ProviderConfigID),
		Platform:         dto.Platform,
		PlatformOwner:    dto.PlatformOwner,
		PlatformRepo:     dto.PlatformRepo,
		RemoteName:       dto.RemoteName,
		IsPrimary:        dto.IsPrimary,
		PlatformRepoId:   dto.PlatformRepoID,
		CreatedAt:        dto.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        dto.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// List .
// @router /api/v1/bindings [GET]
func List(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.ListBindingsReq) ([]*bindingModel.BindingInfo, error) {
		result, err := bindingsvc.ListBindings(req.RepoKey, req.ProviderConfigID)
		if err != nil {
			return nil, handler.ErrInternal("Failed to list bindings: " + err.Error())
		}
		if result == nil {
			result = []api.RepoProviderBindingDTO{}
		}
		protos := make([]*bindingModel.BindingInfo, 0, len(result))
		for _, dto := range result {
			protos = append(protos, convertToProtoBinding(dto))
		}
		return protos, nil
	})
}

// Get .
// @router /api/v1/bindings/:id [GET]
func Get(ctx context.Context, c *app.RequestContext) {
	id, ok := pkgresponse.ParseIDParam(c, "id")
	if !ok {
		return
	}

	result, err := bindingsvc.GetBinding(id)
	if err != nil {
		pkgresponse.NotFound(c, err.Error())
		return
	}
	pkgresponse.Success(c, convertToProtoBinding(*result))
}

// Create .
// @router /api/v1/bindings [POST]
func Create(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.CreateBindingReq) (any, error) {
		result, err := bindingsvc.CreateBinding(ctx, req)
		if err != nil {
			return nil, handler.ErrInternal("Failed to create binding: " + err.Error())
		}
		c.Set("audit_details", map[string]interface{}{"repo_key": req.RepoKey, "provider_config_id": req.ProviderConfigID})
		return result, nil
	})
}

// Update .
// @router /api/v1/bindings/:id [PUT]
func Update(ctx context.Context, c *app.RequestContext) {
	id, ok := pkgresponse.ParseIDParam(c, "id")
	if !ok {
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

// Delete .
// @router /api/v1/bindings/:id [DELETE]
func Delete(ctx context.Context, c *app.RequestContext) {
	id, ok := pkgresponse.ParseIDParam(c, "id")
	if !ok {
		return
	}

	cleanup := c.Query("cleanup_webhook") == "true"
	if err := bindingsvc.DeleteBinding(ctx, id, cleanup); err != nil {
		pkgresponse.InternalServerError(c, "Failed to delete binding: "+err.Error())
		return
	}
	pkgresponse.Success(c, nil)
}

// SetPrimary .
// @router /api/v1/bindings/:id/set-primary [POST]
func SetPrimary(ctx context.Context, c *app.RequestContext) {
	id, ok := pkgresponse.ParseIDParam(c, "id")
	if !ok {
		return
	}

	result, err := bindingsvc.SetPrimary(id)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to set primary: "+err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

// AutoDetect .
// @router /api/v1/bindings/auto-detect [POST]
func AutoDetect(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *api.AutoDetectReq) (any, error) {
		result, err := bindingsvc.AutoDetect(req.RepoKey)
		if err != nil {
			return nil, handler.ErrInternal("Failed to auto-detect: " + err.Error())
		}
		return result, nil
	})
}

// RegisterWebhook .
// @router /api/v1/bindings/:id/webhook [POST]
func RegisterWebhook(ctx context.Context, c *app.RequestContext) {
	id, ok := pkgresponse.ParseIDParam(c, "id")
	if !ok {
		return
	}

	result, err := bindingsvc.RegisterWebhook(ctx, id)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to register webhook: "+err.Error())
		return
	}
	pkgresponse.Success(c, result)
}

// DeleteWebhook .
// @router /api/v1/bindings/:id/webhook [DELETE]
func DeleteWebhook(ctx context.Context, c *app.RequestContext) {
	id, ok := pkgresponse.ParseIDParam(c, "id")
	if !ok {
		return
	}

	result, err := bindingsvc.DeleteWebhook(ctx, id)
	if err != nil {
		pkgresponse.InternalServerError(c, "Failed to delete webhook: "+err.Error())
		return
	}
	pkgresponse.Success(c, result)
}
