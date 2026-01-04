package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

// @Summary List audit logs
// @Description Retrieve a list of system audit logs, ordered by creation time (descending).
// @Tags Audit
// @Produce json
// @Success 200 {object} response.Response{data=[]api.AuditLogDTO}
// @Router /api/audit/logs [get]
func ListAuditLogs(ctx context.Context, c *app.RequestContext) {
	logs, err := db.NewAuditLogDAO().FindLatest(100)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	dtos := make([]api.AuditLogDTO, len(logs))
	for i, log := range logs {
		dtos[i] = api.NewAuditLogDTO(log)
	}
	response.Success(c, dtos)
}
