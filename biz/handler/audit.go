package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/query"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

// @Summary List audit logs
// @Description Retrieve a list of system audit logs, ordered by creation time (descending).
// @Tags Audit
// @Produce json
// @Success 200 {object} response.Response{data=[]model.AuditLog}
// @Router /api/audit/logs [get]
func ListAuditLogs(ctx context.Context, c *app.RequestContext) {
	logs, err := query.NewAuditLogDAO().FindLatest(100)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.Success(c, logs)
}
