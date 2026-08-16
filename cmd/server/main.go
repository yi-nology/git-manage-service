//go:build !desktop

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	hserver "github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	kserver "github.com/cloudwego/kitex/server"
	prometheus "github.com/hertz-contrib/monitor-prometheus"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/kitex_gen/git/gitservice"
	"github.com/yi-nology/git-manage-service/biz/router"
	"github.com/yi-nology/git-manage-service/biz/rpc_handler"
	"github.com/yi-nology/git-manage-service/biz/service/audit"
	"github.com/yi-nology/git-manage-service/biz/service/git"
	"github.com/yi-nology/git-manage-service/biz/service/llm"
	mirrorSvc "github.com/yi-nology/git-manage-service/biz/service/mirror"
	settingssvc "github.com/yi-nology/git-manage-service/biz/service/settings"
	"github.com/yi-nology/git-manage-service/biz/service/stats"
	syncv2 "github.com/yi-nology/git-manage-service/biz/service/sync/v2"
	"github.com/yi-nology/git-manage-service/biz/utils"
	"github.com/yi-nology/git-manage-service/pkg/appinfo"
	"github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-manage-service/pkg/embed"
	"github.com/yi-nology/git-manage-service/pkg/lock"
	"github.com/yi-nology/git-manage-service/pkg/metrics"
	pkgqueue "github.com/yi-nology/git-manage-service/pkg/queue"
	"github.com/yi-nology/git-platform-sdk/gitbackend"
	_ "github.com/yi-nology/git-platform-sdk/backends/all"
)

// @title Git Manage Service API
// @version 2.0
// @description 轻量级多仓库、多分支自动化同步管理系统 API 文档

// @contact.name API Support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url `http://www.apache.org/licenses/LICENSE-2.0.html`

// @host localhost:12345
// @BasePath /api

var (
	mode    = flag.String("mode", "all", "启动模式: http, rpc, all")
	version = flag.Bool("version", false, "显示版本信息")
)

func main() {
	flag.Parse()

	if *version {
		fmt.Printf("%s version %s\n", appinfo.AppName, appinfo.Version)
		fmt.Printf("Build time: %s\n", appinfo.BuildTime)
		fmt.Printf("Git commit: %s\n", appinfo.GitCommit)
		return
	}

	log.Printf("[%s] Starting in '%s' mode...\n", appinfo.AppName, *mode)

	// 初始化共享资源
	initResources()

	// 创建全局上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 信号处理
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	var httpServer *hserver.Hertz
	var rpcServer kserver.Server

	switch *mode {
	case "http":
		httpServer = startHTTPServer()
	case "rpc":
		rpcServer = startRPCServer()
	case "all":
		rpcServer = startRPCServer()
		httpServer = startHTTPServer()
	default:
		log.Fatalf("Unknown mode: %s. Available modes: http, rpc, all", *mode)
	}

	// 等待退出信号
	<-quit
	log.Println("Shutdown signal received, shutting down servers...")

	// 优雅关闭
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	if rpcServer != nil {
		if err := rpcServer.Stop(); err != nil {
			log.Printf("RPC Server shutdown error: %v\n", err)
		} else {
			log.Println("RPC Server stopped")
		}
	}

	if httpServer != nil {
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP Server shutdown error: %v\n", err)
		} else {
			log.Println("HTTP Server stopped")
		}
	}

	audit.AuditSvc.Stop()

	mirrorSvc.StopScheduler()

	log.Println("All servers stopped. Exiting.")
}

// initResources 初始化共享资源
func initResources() {
	log.Println("Initializing resources...")

	// 加载配置
	configs.Init()

	// 初始化数据库
	db.Init()

	db.InitLintRules()

	db.InitBindingMigration()

	settingssvc.LoadCodeReviewSettingsFromDB()

	settingssvc.InitDefaultReviewRules()

	// 初始化加密工具
	utils.InitEncryption()

	// 初始化业务服务
	initSyncV2Service()
	stats.InitStatsService()
	audit.InitAuditService()
	git.GlobalTaskManager.Init()
	initMirrorSystem()

	initQueue()

	log.Println("Resources initialized successfully")
}

// initSyncV2Service 初始化 V2 同步服务 (git-sync-service)
func initSyncV2Service() {
	log.Println("[SyncV2] Initializing git-sync-service...")
	if err := syncv2.GetService().Initialize(&configs.GlobalConfig); err != nil {
		log.Printf("[SyncV2] Warning: failed to initialize: %v\n", err)
		return
	}
	log.Println("[SyncV2] Initialized successfully")
}

func initQueue() {
	// The asynq pipeline was scaffolding that never enqueued anything (the
	// review path invokes RunReview directly); it was removed along with
	// biz/queue. LLM providers must still be initialized — they serve spec-AI,
	// workspace, and maintenance features regardless of code-review state.
	if !configs.GetCodeReviewConfig().Enabled {
		log.Println("[Queue] Code review disabled, skipping LLM provider init")
		return
	}
	llm.InitProviders()
	llm.InitProvidersFromDB()
	log.Println("[Queue] LLM providers initialized")
}

func initMirrorSystem() {
	cfg := configs.GlobalConfig

	backend, err := gitbackend.NewGitBackend(gitbackend.Options{Type: cfg.Mirror.GitBackend})
	if err != nil {
		log.Printf("[Mirror] Warning: git backend init failed: %v", err)
		return
	}

	q, err := pkgqueue.NewQueueFromConfig(cfg.Mirror, cfg.Redis)
	if err != nil {
		log.Printf("[Mirror] Warning: queue init failed: %v", err)
		return
	}

	var lockSvc lock.DistLock
	lockSvc, err = lock.NewDistLock(cfg.Lock)
	if err != nil {
		log.Printf("[Mirror] Warning: lock init failed: %v", err)
	}

	mirrorDAO := db.NewMirrorDAO()
	syncLogDAO := db.NewMirrorSyncLogDAO()

	svc := mirrorSvc.NewMirrorService(mirrorDAO, syncLogDAO, lockSvc, backend, q, cfg.Mirror)
	mirrorSvc.GlobalMirrorService = svc

	wp := pkgqueue.NewWorkerPool(q, func(req pkgqueue.SyncRequest) {
		svc.ProcessSyncRequest(req)
	}, cfg.Mirror.MaxWorkers)
	wp.Start()

	mirrorSvc.InitScheduler(mirrorDAO, q, cfg.Mirror)

	log.Println("[Mirror] System initialized successfully")
}

// startHTTPServer 启动 HTTP 服务器
func startHTTPServer() *hserver.Hertz {

	addr := fmt.Sprintf(":%d", configs.GlobalConfig.Server.Port)

	opts := []config.Option{
		hserver.WithHostPorts(addr),
	}

	if configs.GlobalConfig.Prometheus.Enabled {
		metricsAddr := fmt.Sprintf(":%d", configs.GlobalConfig.Prometheus.Port)
		metricsPath := configs.GlobalConfig.Prometheus.Path
		tracer := prometheus.NewServerTracer(
			metricsAddr,
			metricsPath,
			prometheus.WithRegistry(metrics.Registry),
			prometheus.WithEnableGoCollector(true),
		)
		opts = append(opts, hserver.WithTracer(tracer))
		log.Printf("Prometheus metrics enabled on %s%s\n", metricsAddr, metricsPath)
	}

	h := hserver.Default(opts...)

	// 初始化嵌入的静态文件系统
	router.SetEmbedFS(embed.GetPublicFS(), embed.GetDocsFS())

	// 注册路由
	router.GeneratedRegister(h)

	go func() {
		log.Printf("HTTP Server starting on %s\n", addr)
		if err := h.Run(); err != nil {
			log.Printf("HTTP Server stopped with error: %v\n", err)
		}
	}()

	return h
}

// startRPCServer 启动 RPC 服务器
func startRPCServer() kserver.Server {
	addr := fmt.Sprintf(":%d", configs.GlobalConfig.Rpc.Port)
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		log.Fatalf("Failed to resolve RPC address: %v", err)
	}

	svr := gitservice.NewServer(
		new(rpc_handler.GitServiceImpl),
		kserver.WithServiceAddr(tcpAddr),
	)

	go func() {
		log.Printf("RPC Server starting on %s\n", addr)
		if err := svr.Run(); err != nil {
			log.Printf("RPC Server stopped with error: %v\n", err)
		}
	}()

	return svr
}
