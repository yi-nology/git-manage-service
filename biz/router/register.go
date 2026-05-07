package router

import (
	"context"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	handler_maintenance "github.com/yi-nology/git-manage-service/biz/handler/maintenance"
	handler_review "github.com/yi-nology/git-manage-service/biz/handler/review"
	handler_settings "github.com/yi-nology/git-manage-service/biz/handler/settings"
	"github.com/yi-nology/git-manage-service/biz/middleware"
	"github.com/yi-nology/git-manage-service/biz/router/ai"
	"github.com/yi-nology/git-manage-service/biz/router/audit"
	author "github.com/yi-nology/git-manage-service/biz/router/author"
	"github.com/yi-nology/git-manage-service/biz/router/binding"
	"github.com/yi-nology/git-manage-service/biz/router/branch"
	"github.com/yi-nology/git-manage-service/biz/router/commit"
	"github.com/yi-nology/git-manage-service/biz/router/cr"
	"github.com/yi-nology/git-manage-service/biz/router/credential"
	"github.com/yi-nology/git-manage-service/biz/router/file"
	maintenance "github.com/yi-nology/git-manage-service/biz/router/maintenance"
	"github.com/yi-nology/git-manage-service/biz/router/notification"
	"github.com/yi-nology/git-manage-service/biz/router/patch"
	"github.com/yi-nology/git-manage-service/biz/router/provider"
	"github.com/yi-nology/git-manage-service/biz/router/repo"
	review "github.com/yi-nology/git-manage-service/biz/router/review"
	"github.com/yi-nology/git-manage-service/biz/router/settings"
	"github.com/yi-nology/git-manage-service/biz/router/spec"
	"github.com/yi-nology/git-manage-service/biz/router/sshkey"
	"github.com/yi-nology/git-manage-service/biz/router/stash"
	"github.com/yi-nology/git-manage-service/biz/router/stats"
	"github.com/yi-nology/git-manage-service/biz/router/submodule"
	"github.com/yi-nology/git-manage-service/biz/router/sync"
	"github.com/yi-nology/git-manage-service/biz/router/system"
	"github.com/yi-nology/git-manage-service/biz/router/tag"
	"github.com/yi-nology/git-manage-service/biz/router/version"
	"github.com/yi-nology/git-manage-service/biz/router/webhook"
	"github.com/yi-nology/git-manage-service/biz/router/webhook_event"
	workspace "github.com/yi-nology/git-manage-service/biz/router/workspace"
)

// 嵌入的文件系统变量
var (
	embeddedPublicFS fs.FS
	embeddedDocsFS   fs.FS
)

// SetEmbedFS 设置嵌入的文件系统（由 main.go 调用）
func SetEmbedFS(public, docs fs.FS) {
	embeddedPublicFS = public
	embeddedDocsFS = docs
}

// GeneratedRegister registers all routes
func GeneratedRegister(h *server.Hertz) {
	h.Use(middleware.CORS())
	h.Use(middleware.AuditMiddleware())

	//INSERT_POINT: DO NOT DELETE THIS LINE!
	workspace.Register(h)

	author.Register(h)

	maintenance.Register(h)
	h.POST("/api/v1/repo/:repo_key/maintenance/slim-prefix/preview", handler_maintenance.PreviewPrefixSlim)
	h.POST("/api/v1/repo/:repo_key/maintenance/slim-prefix", handler_maintenance.SlimByPrefix)
	h.POST("/api/v1/repo/:repo_key/maintenance/force-push", handler_maintenance.ForcePushRemotes)

	settings.Register(h)
	h.GET("/api/v1/settings/ollama-models", handler_settings.FetchOllamaModels)
	ai.Register(h)
	webhook_event.Register(h)
	provider.Register(h)
	cr.Register(h)

	review.Register(h)
	h.POST("/api/v1/reviews/tasks/provider", handler_review.CreateTaskByProvider)
	h.GET("/api/v1/reviews/tasks/provider", handler_review.ListTasksByProvider)
	audit.Register(h)

	branch.Register(h)
	commit.Register(h)
	credential.Register(h)
	file.Register(h)
	notification.Register(h)
	repo.Register(h)
	sshkey.Register(h)
	stash.Register(h)
	submodule.Register(h)
	sync.Register(h)
	system.Register(h)
	tag.Register(h)
	version.Register(h)
	webhook.Register(h)
	patch.Register(h)
	spec.Register(h)
	stats.Register(h)
	binding.Register(h)

	// 根路径
	h.GET("/", func(ctx context.Context, c *app.RequestContext) {
		if embeddedPublicFS != nil {
			serveEmbedFile(c, embeddedPublicFS, "index.html", "text/html; charset=utf-8")
		} else {
			c.String(http.StatusInternalServerError, "Embedded file system not initialized")
		}
	})

	// 前端 SPA - 使用 NoRoute 处理未匹配的路由
	// 避免使用 GET /*filepath 通配符路由，因为它会与所有 API POST 路由冲突导致 405
	h.NoRoute(func(ctx context.Context, c *app.RequestContext) {
		fp := string(c.Path())
		fp = strings.TrimPrefix(fp, "/")

		// API 路径返回 404
		if strings.HasPrefix(fp, "api/") {
			c.JSON(http.StatusNotFound, map[string]interface{}{
				"code": 404,
				"msg":  "not found",
			})
			return
		}

		// 只处理 GET 和 HEAD 请求的静态文件
		method := string(c.Method())
		if method != "GET" && method != "HEAD" {
			c.Status(http.StatusMethodNotAllowed)
			return
		}

		if fp == "" {
			fp = "index.html"
		}
		// 尝试读取文件，失败则回退到 index.html (SPA)
		if embeddedPublicFS != nil {
			if !serveEmbedFile(c, embeddedPublicFS, fp, "") {
				serveEmbedFile(c, embeddedPublicFS, "index.html", "text/html; charset=utf-8")
			}
		} else {
			c.String(http.StatusInternalServerError, "Embedded file system not initialized")
		}
	})
}

// serveEmbedFile 从嵌入的文件系统提供文件
func serveEmbedFile(c *app.RequestContext, fsys fs.FS, path string, contentType string) bool {
	// 清理路径
	path = strings.TrimPrefix(path, "/")
	if path == "" {
		path = "index.html"
	}

	// 检查文件系统是否为 nil
	if fsys == nil {
		return false
	}

	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return false
	}

	if contentType == "" {
		contentType = getContentType(path)
	}
	c.Data(http.StatusOK, contentType, data)
	return true
}

// getContentType 根据文件扩展名返回 MIME 类型
func getContentType(path string) string {
	ext := filepath.Ext(path)
	if mimeType := mime.TypeByExtension(ext); mimeType != "" {
		return mimeType
	}
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".json":
		return "application/json; charset=utf-8"
	case ".svg":
		return "image/svg+xml"
	default:
		return "application/octet-stream"
	}
}
