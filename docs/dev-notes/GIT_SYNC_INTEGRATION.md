# git-sync-service 集成指南

## 概述

本项目已集成 `github.com/yi-nology/git-sync-service` 作为底层同步核心引擎。
集成采用 **混合模式**：优先使用 git-sync-service，初始化失败时自动降级为原有实现。

## 目录结构

```
biz/service/sync/
├── adapter/                    # 适配层
│   ├── model_converter.go      # 数据模型转换
│   ├── config.go              # 配置转换
│   ├── gitsync_service.go     # git-sync-service 包装服务
│   ├── dao_bridge.go          # DAO 层桥接
│   └── init.go               # 全局初始化
├── sync_service_v2.go         # 新版同步服务（推荐使用）
├── sync_service.go            # 原有实现（fallback）
└── sync_executor.go          # 原有执行器
```

## 快速开始

### 1. 添加依赖

在 `go.mod` 中添加：

```go
require github.com/yi-nology/git-sync-service v0.1.0
```

执行：
```bash
go mod tidy
```

### 2. 全局初始化（可选）

在应用启动时：

```go
import "github.com/yi-nology/git-manage-service/biz/service/sync/adapter"

func main() {
    // 初始化 git-sync-service
    adapter.InitGlobalGitSyncService()
    defer adapter.ShutdownGlobalGitSyncService()
    
    // ... 其他初始化
}
```

### 3. 使用新版同步服务

```go
import "github.com/yi-nology/git-manage-service/biz/service/sync"

// 创建服务
svc := sync.NewSyncServiceV2()

// 检查是否使用 git-sync-service
if svc.IsUsingGitSyncService() {
    fmt.Println("Using git-sync-service core")
} else {
    fmt.Println("Using native implementation (fallback mode)")
}

// 执行同步任务
err := svc.RunTask("task-key-xxx")

// 带触发源执行
err := svc.RunTaskWithTrigger("task-key-xxx", po.TriggerSourceWebhook)

// 直接执行同步（无需创建任务）
err := svc.ExecuteSync(&po.SyncTask{...})

// 预览同步
preview, err := svc.PreviewSync(task)
fmt.Printf("Need sync: %v, Commits: %d\n", preview.NeedSync, preview.CommitCount)

// 批量同步
err := svc.BatchSync([]string{"task1", "task2", "task3"})
```

## 特性说明

### ✅ 自动降级机制

- git-sync-service 初始化成功 → 使用核心库
- 初始化失败（依赖缺失、配置错误等）→ 自动降级为原有实现
- 不影响现有功能运行

### ✅ 完全兼容 API

- 保持所有原有 API 接口不变
- Handler 层无需修改
- 前端调用方式完全一致

### ✅ 渐进式迁移

- 可以逐步切换到 git-sync-service
- 保留原有代码作为安全网
- 支持 A/B 测试两种实现

## 配置说明

git-sync-service 自动从现有的配置中读取：

- **Database**: 使用相同的数据库配置（SQLite/MySQL/PostgreSQL）
- **Redis**: 使用相同的 Redis 配置（分布式锁）
- **Git**: 自动创建临时工作目录
- **Sync**: 并发数从 `mirror.max_workers` 读取

## 实现对比

| 特性 | git-manage-service (native) | git-sync-service (core) |
|------|---------------------------|-----------------------|
| Git 操作 | go-git 纯 Go | 系统 git 命令 |
| 本地源支持 | ✅ | ⚠️ 待完善 |
| 预览模式 | 基础 | 增强 |
| Webhook 规则引擎 | 基础 | 完整 |
| 多平台支持 | 有限 | 完整（git-platform-sdk）|
| 可复用性 | 项目绑定 | 独立可复用 |

## 如何切换回原有实现

如果需要强制使用原有实现，修改 `sync_service_v2.go`:

```go
func NewSyncServiceV2() *SyncServiceV2 {
    svc := &SyncServiceV2{
        // ...
        useNativeImpl: true,  // 强制使用原有实现
    }
    // ...
}
```

或者在 Handler 层直接使用 `NewSyncService()`。

## 注意事项

1. **系统 git 依赖**: git-sync-service 需要系统安装 git 命令
   ```bash
   git --version  # 确保 >= 2.30.0
   ```

2. **临时目录**: git-sync-service 会在系统临时目录创建工作区，确保有写入权限

3. **数据一致性**: 两个实现共享相同的数据库表，数据完全兼容

## 后续优化计划

- [ ] 完整的 dry-run 模式支持
- [ ] 分支过滤功能（白名单/黑名单）
- [ ] 子模块递归同步
- [ ] 同步超时控制
- [ ] Git 操作速率限制

## 问题反馈

如遇到 git-sync-service 的问题，请在以下地址提交 Issue：
https://github.com/yi-nology/git-sync-service/issues
