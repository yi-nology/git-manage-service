package webhook_event

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	webhook_event "github.com/yi-nology/git-manage-service/biz/handler/webhook_event"
)

// RegisterRules mounts the webhook rule management routes.
//
// These CRUD endpoints are intentionally not part of the generated IDL router;
// they live in this hand-written file (same package) so they are not dropped by
// `hz update`. The rule engine itself (applyRules) and the auto-created
// code_review rules from binding registration are unaffected.
func RegisterRules(h *server.Hertz) {
	g := h.Group("/api/v1/webhook")
	g.GET("/rules", webhook_event.ListRules)
	g.POST("/rules", webhook_event.CreateRule)
	g.PUT("/rules/:id", webhook_event.UpdateRule)
	g.DELETE("/rules/:id", webhook_event.DeleteRule)
}
