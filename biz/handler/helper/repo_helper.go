package helper

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

// BindJSON 绑定 JSON 请求体
func BindJSON[T any](c *app.RequestContext) (*T, bool) {
	var req T
	if err := c.BindJSON(&req); err != nil {
		response.BadRequest(c, "invalid JSON: "+err.Error())
		return nil, false
	}
	return &req, true
}
