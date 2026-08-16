package credential

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/model/credential"
	creds "github.com/yi-nology/git-manage-service/biz/service/credential"
	"github.com/yi-nology/git-manage-service/pkg/handler"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

var svc = creds.NewCredentialService()

func List(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *credential.ListCredentialsRequest) (any, error) {
		return svc.List(req)
	})
}

func Create(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *credential.CreateCredentialRequest) (any, error) {
		return svc.Create(req)
	})
}

func Match(ctx context.Context, c *app.RequestContext) {
	handler.BindAndDo(c, func(req *credential.MatchCredentialRequest) (any, error) {
		return svc.Match(req.Url)
	})
}

func Get(ctx context.Context, c *app.RequestContext) {
	id, ok := response.ParseIDParam(c, "id")
	if !ok {
		return
	}
	resp, err := svc.Get(uint(id))
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, resp)
}

func Update(ctx context.Context, c *app.RequestContext) {
	id, ok := response.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req credential.UpdateCredentialRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	resp, err := svc.Update(uint(id), &req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, resp)
}

func Delete(ctx context.Context, c *app.RequestContext) {
	id, ok := response.ParseIDParam(c, "id")
	if !ok {
		return
	}
	if err := svc.Delete(uint(id)); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, map[string]string{"message": "Credential deleted successfully"})
}

func Test(ctx context.Context, c *app.RequestContext) {
	id, ok := response.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req credential.TestCredentialRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	success, message, err := svc.TestConnection(uint(id), req.Url)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, map[string]interface{}{"success": success, "message": message})
}

func GetUsages(ctx context.Context, c *app.RequestContext) {
	id, ok := response.ParseIDParam(c, "id")
	if !ok {
		return
	}
	resp, err := svc.GetUsages(uint(id))
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, resp)
}

func Rotate(ctx context.Context, c *app.RequestContext) {
	id, ok := response.ParseIDParam(c, "id")
	if !ok {
		return
	}
	var req credential.RotateCredentialRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := svc.Rotate(uint(id), &req); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, map[string]string{"message": "Credential rotated successfully"})
}
