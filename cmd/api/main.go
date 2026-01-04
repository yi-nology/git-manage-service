package main

import (
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/yi-nology/git-manage-service/pkg/configs"
	"github.com/yi-nology/git-manage-service/biz/dal/db"
	"github.com/yi-nology/git-manage-service/biz/router"
	"github.com/yi-nology/git-manage-service/biz/service/audit"
	"github.com/yi-nology/git-manage-service/biz/service/stats"
	"github.com/yi-nology/git-manage-service/biz/service/sync"
	"github.com/yi-nology/git-manage-service/biz/utils"

	_ "github.com/yi-nology/git-manage-service/docs"
)

// @title Branch Management Tool API
// @version 1.1
// @description API documentation for Branch Management Tool.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.basic BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/

func main() {
	// 0. Init Config
	configs.Init()

	// 1. Init DB
	db.Init()

	// 2. Init Cron & Services
	sync.InitCronService()
	stats.InitStatsService()
	audit.InitAuditService()
	utils.InitEncryption()

	// 3. Init Server
	addr := fmt.Sprintf(":%d", configs.GlobalConfig.Server.Port)
	h := server.Default(server.WithHostPorts(addr))

	// 4. Register Routes
	router.GeneratedRegister(h)

	// 5. Spin
	h.Spin()
}
