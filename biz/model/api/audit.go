package api

import (
	"time"

	"github.com/yi-nology/git-manage-service/biz/model/po"
)

type AuditLogDTO struct {
	ID        uint      `json:"id"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	Operator  string    `json:"operator"`
	Details   string    `json:"details"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	CreatedAt time.Time `json:"created_at"`
}

func NewAuditLogDTO(l po.AuditLog) AuditLogDTO {
	return AuditLogDTO{
		ID:        l.ID,
		Action:    l.Action,
		Target:    l.Target,
		Operator:  l.Operator,
		Details:   l.Details,
		IPAddress: l.IPAddress,
		UserAgent: l.UserAgent,
		CreatedAt: l.CreatedAt,
	}
}
