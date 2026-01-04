package handler

import (
	"context"
	"strconv"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/model/api"
	"github.com/yi-nology/git-manage-service/pkg/response"
)

// @Summary List audit logs
// @Description Retrieve a list of system audit logs with pagination
// @Tags Audit
// @Param page query int false "Page number (default 1)"
// @Param page_size query int false "Page size (default 20)"
// @Produce json
// @Success 200 {object} response.Response{data=map[string]interface{}}
// @Router /api/audit/logs [get]
func ListAuditLogs(ctx context.Context, c *app.RequestContext) {
	page, _ := strconv.Atoi(c.Query("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize < 1 {
		pageSize = 20
	}

	dao := db.NewAuditLogDAO()
	logs, err := dao.FindPage(page, pageSize)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	total, _ := dao.Count()

	dtos := make([]api.AuditLogDTO, len(logs))
	for i, log := range logs {
		dtos[i] = api.NewAuditLogDTO(log)
		// Details are explicitly excluded in FindPage, so they are empty here
	}

	c.JSON(consts.StatusOK, map[string]interface{}{
		"code": 0,
		"msg":  "success",
		"data": map[string]interface{}{
			"items": dtos,
			"total": total,
			"page":  page,
			"size":  pageSize,
		},
	})
}

// @Summary Get audit log details
// @Description Retrieve details for a specific audit log entry
// @Tags Audit
// @Param id path int true "Audit Log ID"
// @Produce json
// @Success 200 {object} response.Response{data=api.AuditLogDTO}
// @Router /api/audit/logs/{id} [get]
func GetAuditLog(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		response.BadRequest(c, "invalid id")
		return
	}

	log, err := db.NewAuditLogDAO().FindByID(uint(id))
	if err != nil {
		response.NotFound(c, "audit log not found")
		return
	}

	response.Success(c, api.NewAuditLogDTO(*log))
}
