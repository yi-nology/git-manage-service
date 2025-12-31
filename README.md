# Git Manage Service (Git 管理服务)

Git Manage Service 是一个轻量级的多仓库、多分支自动化同步管理系统。它提供了友好的 Web 界面，支持定时任务、Webhook 触发以及详细的同步日志记录。

## 🚀 功能特性

- **多仓库管理**：轻松注册和管理本地 Git 仓库。
- **灵活同步规则**：支持任意 Remote 和分支之间的同步（如 `origin/main` -> `ky/main`）。
- **自动化执行**：内置 Cron 调度器，支持定时同步。
- **Webhook 集成**：支持通过外部系统（如 CI/CD）触发同步。
- **安全可靠**：支持冲突检测、Fast-Forward 检查及 Force Push 保护。
- **可视化界面**：提供直观的 Web UI，查看历史、日志及管理任务。

## 📚 文档

- [产品手册与使用说明](docs/product_manual.md)
- [Webhook 接口文档](docs/webhook.md)

## 🛠 快速开始

### 1. 编译运行
```bash
go mod tidy
go build -o git-manage-service
./git-manage-service
```

### 2. 访问界面
浏览器打开: [http://localhost:8080](http://localhost:8080)

## 📦 项目结构
```
.
├── biz/            # 业务逻辑 (Service, Handler, Model)
├── docs/           # 项目文档
├── public/         # 前端静态资源
├── test/           # 测试工具
├── main.go         # 入口文件
└── go.mod          # 依赖定义
```

## 📝 License
MIT
