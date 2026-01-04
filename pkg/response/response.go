package response

import (
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/yi-nology/git-manage-service/pkg/errno"
)

// Response standard structure
type Response struct {
	Code    int32       `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success response
func Success(c *app.RequestContext, data interface{}) {
	c.JSON(consts.StatusOK, Response{
		Code:    errno.Success.ErrCode,
		Message: errno.Success.ErrMsg,
		Data:    data,
	})
}

// Error response
func Error(c *app.RequestContext, err error) {
	e := errno.ConvertErr(err)
	c.JSON(consts.StatusOK, Response{
		Code:    e.ErrCode,
		Message: e.ErrMsg,
	})
}

// Deprecated: Use Error instead. Kept for backward compatibility during refactor.
func BadRequest(c *app.RequestContext, msg string) {
	Error(c, errno.ParamErr.WithMessage(msg))
}

// Deprecated: Use Error instead. Kept for backward compatibility during refactor.
func InternalServerError(c *app.RequestContext, msg string) {
	Error(c, errno.ServiceErr.WithMessage(msg))
}

// Deprecated: Use Error instead. Kept for backward compatibility during refactor.
func NotFound(c *app.RequestContext, msg string) {
	// 404 is usually a client error, could map to ParamErr or a new NotFoundErr
	Error(c, errno.ParamErr.WithMessage(msg))
}
