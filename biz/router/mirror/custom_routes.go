package mirror

import (
	"github.com/cloudwego/hertz/pkg/app/server"
	handler "github.com/yi-nology/git-manage-service/biz/handler/mirror"
)

func RegisterCustomRoutes(h *server.Hertz) {
	mirror := h.Group("/api/v1/mirror")
	{
		mirror.POST("", handler.CreateMirror)
		mirror.POST("/analyze", handler.AnalyzeRemote)
		mirror.POST("/validate-credential", handler.ValidateCredential)
		mirror.POST("/webhook/:token", handler.HandleWebhook)
	}

	mirrorWithID := h.Group("/api/v1/mirror/:id")
	{
		mirrorWithID.GET("", handler.GetMirror)
		mirrorWithID.POST("/update", handler.UpdateMirror)
		mirrorWithID.POST("/delete", handler.DeleteMirror)
		mirrorWithID.POST("/sync", handler.TriggerSync)
		mirrorWithID.POST("/preview", handler.PreviewSync)
		mirrorWithID.GET("/logs", handler.ListSyncLogs)
		mirrorWithID.POST("/pause", handler.PauseMirror)
		mirrorWithID.POST("/resume", handler.ResumeMirror)
	}

	mirrorLog := h.Group("/api/v1/mirror/log")
	{
		mirrorLog.GET("/:log_id", handler.GetSyncLog)
		mirrorLog.POST("/:log_id/delete", handler.DeleteSyncLog)
	}

	mirrors := h.Group("/api/v1/mirrors")
	{
		mirrors.GET("", handler.ListMirrors)
		mirrors.POST("/sync", handler.BatchTriggerSync)
	}
}
