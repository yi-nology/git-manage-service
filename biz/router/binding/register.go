package binding

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	bindinghandler "github.com/yi-nology/git-manage-service/biz/handler/binding"
)

func Register(h *server.Hertz) {
	g := h.Group("/api/v1/bindings")
	{
		g.GET("", bindinghandler.List)
		g.GET("/:id", bindinghandler.Get)
		g.POST("", bindinghandler.Create)
		g.PUT("/:id", bindinghandler.Update)
		g.DELETE("/:id", bindinghandler.Delete)
		g.POST("/auto-detect", bindinghandler.AutoDetect)
		g.POST("/:id/set-primary", bindinghandler.SetPrimary)
		g.POST("/:id/webhook", bindinghandler.RegisterWebhook)
		g.DELETE("/:id/webhook", bindinghandler.DeleteWebhook)
	}
}
