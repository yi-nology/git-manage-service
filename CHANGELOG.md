# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

## [0.13.0] - 2026-11-13

仓级兼容垫片清仓（**BREAKING**）。**下游要做什么**：仓外若有引用下列符号的代码，
按下表迁移——

| 删除的符号 | 迁移方式 |
| --- | --- |
| `errno.AuthorizationFailedErr` | `errno.Unauthorized` |
| `errno.RecordNotFound` | `errno.NotFound` |
| `authSvc.ResolveAuthFromParams(t, k, s, 0)` | `authSvc.ResolveAuth(domain.AuthInfo{Type: t, Key: k, Secret: s, Source: "local"})`；需要数据库密钥时先查 key 再 `Source: "database", SSHKeyID: id` |
| `MaintenanceDAO.Update` / `AuthorIdentityDAO.Update` / `SSHKeyDAO.Update` | `BaseDAO.Save`（同实现同语义，upsert 整行） |

### Removed

- **`pkg/errno` 兼容旧错误码死别名**：`AuthorizationFailedErr = Unauthorized`、
  `RecordNotFound = NotFound`（仓内零引用）。
- **`AuthService.ResolveAuthFromParams`（兼容旧接口）**：散参 + sshKeyID 的旧签名。
  两处调用点（handler/system、handler/repo）改直接组 `domain.AuthInfo` 调 `ResolveAuth`。
- **三个 DAO 的 `Update` 转发**（与 `BaseDAO.Save` 同实现「保持向后兼容」）：
  调用点 maintenance_service（3 处）、author_service_fix（3 处）、sshkey_service、
  author_service 及两个 DAO 测试全部改用 `Save`。

### 保留（明确不动）

- `biz/service/git/git_service.go` 的 `RunCommand`：活跃原生 git 逃生门（20+ 调用），
  本版不删；后续仅标 Deprecated 观察。
- 生成层 `.pb.go` / handler 生成物未动。

## [0.12.2] - 2026-11-11

### 修复

- **/api/v1/stats/commits 时间窗静默失效（#24）**：GetCommits 此前把 since/until
  只打进一行 DEBUG 日志、从未参与过滤，接口对任意窗口都返回全量历史。现与同服务
  /analyze、/export/csv 三端点同源：新增 `pkg/timefmt.ParseDateWindow` /
  `InDateWindow` 作为窗口口径的单一真相源（本地时区，since 取当日零点、until 取
  当日末尾），commits 端点在 Go 侧按 author date（%ad）过滤——git log
  `--since/--until` 过滤的是 committer date，与本服务统计口径（%ad/%at）不符，
  rebase/amend 过的提交两个日期分叉，故不下推。同时把 stats 侧「time.ParseInLocation
  忽略错误、非法日期静默退化成不过滤」显式化为 `*timefmt.DateWindowError`，三端点
  统一按 400 回应。

### Changed

- `go get -u ./...` 全量依赖升级：go-git-platform v0.124.0、git-ferry-core
  v0.35.1、oekit/pkg v0.42.0、hertz v0.10.6、gorm/driver/postgres v1.6.3 等，
  `go test -race` 全量通过，无破坏性升级需要钉回。

## [0.12.1] - 2026-11-09

### Changed

- AddChannelPage 通知渠道表单拆分为 ChannelFormFields/ChannelTriggerEvents 组件，686→304 行（#15）
- Top3 大页面拆分：RemoteRepoDetail/SyncTask 组件化拆分，行为保持，vue-tsc+vite build 全绿，浏览器截图对照验收（#16）

## [0.12.0] - 2026-11-08

「全部 API 无鉴权」的 P0 缺口按最小代价方案收口（静态 Bearer token +
回环绑定回落），叠加 XSS 消毒与目录逃逸消毒——#14 单 issue 描述的
「XSS 即全权」攻击链三面同时闭合。

### 新增

- **静态 Bearer 鉴权门（#14）**：`server.auth_token` 配置后所有 /api 路由
  要求 `Authorization: Bearer <token>`（crypto/subtle 常量时间比较，探针
  /CORS 预检放行）；**为空时 HTTP 只绑 127.0.0.1 并打启动告警**——此前
  server 模式恒绑 0.0.0.0 且全 API 无鉴权（多轮审计的最大开放风险）。

### 修复

- **XSS：五处 v-html 全部收口消毒（#19）**：新增 `utils/markdown.ts`
  `renderMarkdownSafe`（escape → marked → DOMPurify 白名单）为唯一渲染
  出口，ReviewSummary/AIPanel 换用；`useAIChat.renderMarkdown` 与
  `AuthorFix.formatAI` 两个手写管线先转义再做格式替换；diff2html 输出加
  DOMPurify 兜底；`marked`/`dompurify` 声明为直接依赖（幽灵依赖消除）。
  AI/远端 CR 内容不再直达 innerHTML。
- **Clone 目录逃逸消毒（#20）**：`repoNameFromURL` 的 `..`/`.`/`?`/空段/
  分隔符/控制字符形态一律回退时间戳名——URL 结尾带 `..` 可把克隆目标抬到
  选目录父目录，且「已是 git 仓库则拒」守卫拦不住这些形态。
- **旧镜像队列键启动清扫（#17）**：v2 前卡死槽位残留键
  （`gms:mirror_sync_queue:set/list`）启动时检测即删并告警——v2 不读它们，
  删了无副作用。
- **Ollama 调用迁 httpjson 单源（#18）**：`Chat` 走 `httpjson.DoJSON`
  （SSRF 防护客户端原样传入，防护不降级）；`FetchOllamaModels` 走
  `DoRaw`（保 1MB 响应上限口径）。

 - 2026-03-08

### Added
- **UI/UX 优化**
  - 深色模式支持（可通过主题切换按钮切换）
  - 全局快捷键支持（Cmd/Ctrl + K 搜索，Cmd/Ctrl + N 新建）
  - 统一的 Toast 通知系统（useNotification composable）
  - 通用分页功能（usePagination composable）
  - 骨架屏加载组件（TableSkeleton、AppSkeleton）
  - 深色模式 CSS 变量支持
  - 优化的滚动条样式

- **新增组件**
  - `ThemeSwitch.vue`: 深色模式切换组件
  - `AppSkeleton.vue`: 通用骨架屏组件
  - `TableSkeleton.vue`: 表格骨架屏组件

- **新增 Composables**
  - `useNotification`: 统一的消息通知管理
  - `usePagination`: 通用的分页逻辑
  - `useKeyboard`: 全局快捷键支持

- **新增 Store**
  - `useUIStore`: UI 全局状态管理（深色模式、侧边栏等）

### Changed
- **仓库列表页优化** (`RepoListPage.vue`)
  - 添加搜索/筛选功能
  - 添加分页功能（10/20/50/100 条每页）
  - 添加骨架屏加载动画
  - 优化操作按钮（使用下拉菜单，减少视觉负担）
  - 添加表格排序功能
  - 优化空状态显示
  - 响应式设计优化

- **全局样式优化** (`style.css`)
  - 添加深色模式 CSS 变量
  - 优化滚动条样式
  - 添加过渡动画

- **布局优化** (`AppLayout.vue`)
  - 添加深色模式切换按钮
  - 添加快捷键提示
  - 优化响应式布局
  - 深色模式样式支持

### Fixed
- 修复 TypeScript 类型错误
- 修复分页逻辑中的类型定义问题
- 修复组件中的未使用导入警告

### Performance
- 优化大数据集渲染性能（分页）
- 添加骨架屏减少加载焦虑

## [Previous Releases]

查看 GitHub Releases 页面获取历史版本信息。
