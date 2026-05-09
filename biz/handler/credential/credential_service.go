package credential

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/model/credential"
	creds "github.com/yi-nology/git-manage-service/biz/service/credential"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

var svc = creds.NewCredentialService()

func List(ctx context.Context, c *app.RequestContext) {
	var req credential.ListCredentialsRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := svc.List(&req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, resp)
}

func Create(ctx context.Context, c *app.RequestContext) {
	var req credential.CreateCredentialRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := svc.Create(&req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, resp)
}

func Match(ctx context.Context, c *app.RequestContext) {
	var req credential.MatchCredentialRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	resp, err := svc.Match(req.Url)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, resp)
}

func Get(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
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
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
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
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	if err := svc.Delete(uint(id)); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, map[string]string{"message": "Credential deleted successfully"})
}

func Test(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
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
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
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
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid id")
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
