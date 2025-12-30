package main

import (
	"context"
	"git-sync-tool/biz/config"
	"git-sync-tool/biz/dal"
	"git-sync-tool/biz/handler"
	"git-sync-tool/biz/middleware"
	"git-sync-tool/biz/service"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
)

func main() {
	// 0. Init Config
	config.Init()

	// 1. Init DB
	dal.Init()

	// 2. Init Cron
	service.InitCronService()

	// 3. Init Server
	h := server.Default(server.WithHostPorts(":8080"))

	// 4. Register Routes
	h.POST("/api/repos", handler.RegisterRepo)
	h.GET("/api/repos", handler.ListRepos)

	h.POST("/api/sync/tasks", handler.CreateTask)
	h.GET("/api/sync/tasks", handler.ListTasks)
	h.GET("/api/sync/tasks/:id", handler.GetTask)
	h.PUT("/api/sync/tasks/:id", handler.UpdateTask)
	h.POST("/api/sync/run", handler.RunSync)
	h.GET("/api/sync/history", handler.ListHistory)

	// Webhook
	h.POST("/api/webhooks/task-sync", middleware.WebhookAuth(), handler.HandleWebhookTrigger)

	// 5. Static Files (Frontend)
	h.Static("/", "./public")

	// Redirect root to index.html if needed, but Static usually handles index.html
	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		c.File("./public/index.html")
	})

	h.Spin()
}
