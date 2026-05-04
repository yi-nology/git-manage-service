# 项目不合理点与后续改造建议

> 生成时间：2026-05-04  
> 范围：基于当前仓库的静态扫描、关键配置阅读、部分测试执行结果整理。本文不替代完整架构评审，但可以作为后续拆任务的优先级清单。

## 结论摘要

这个项目的核心能力比较多，后端、前端、桌面端、文档站、部署脚本、CI 发布链路都放在一个仓库里。当前最大的问题不是单点代码写法，而是**工程边界和约定不够稳定**：端口、字段命名、生成代码边界、构建方式、数据库迁移、测试入口之间存在多套说法。

建议按下面顺序改：

1. 先统一构建、端口、文档和 CI，避免“本地能跑、发布不可用”。
2. 再统一 API 字段命名和前端类型生成，减少前后端字段错配。
3. 然后治理数据库迁移、全局状态、测试隔离这些基础设施问题。
4. 最后再做模块拆分和 UI/业务重构。

## 已验证现状

- `npx vue-tsc --noEmit`：通过。
- `go test ./...`：大部分单元包已通过，但命令超过 2 分钟仍未结束，随后手动停止。当前全量测试存在卡住风险，疑似集成测试或外部依赖等待没有明确超时。
- 当前工作区已有较多未提交改动，包括 `AGENTS.md`、`README.md`、维护模块、LLM 设置、集成测试等；本文只新增文档，不评价这些变更是否应提交。

## 主要不合理点

### 1. 构建链路存在互相矛盾的假设

**表现**

- `Makefile` 明确导出 `CGO_ENABLED=1`。
- `AGENTS.md` 也说明 SQLite 需要 CGO。
- `.github/workflows/release.yml` 发布构建时却设置 `CGO_ENABLED=0`。
- `go.mod` 使用 `go 1.25.0`，release workflow 使用 Go 1.25，desktop workflow 却使用 Go 1.21。

**影响**

- 本地构建、CI 发布、桌面构建使用不同环境，容易出现发布产物不可用。
- 如果 SQLite 驱动或平台编译确实依赖 CGO，release 构建会和本地测试结果脱节。

**建议**

- 统一 Go 版本，建议在 `go.mod`、`.github/workflows/*.yml`、README、AGENTS 中只保留一个版本基线。
- 明确 SQLite 驱动是否真的需要 CGO；如果需要，CI 全部使用 `CGO_ENABLED=1` 并安装对应 C 工具链；如果不需要，删除文档和 Makefile 里的 CGO 强约束。
- 将 `make build-full` 作为 release workflow 的唯一构建入口，避免 CI 里重复手写构建逻辑。

**优先级：P0**

### 2. 端口和访问地址在文档、脚本、配置中大量不一致

**表现**

- `conf/config.yaml` 使用 HTTP `12345`。
- `pkg/configs/loader.go` 默认 HTTP 端口是 `8080`。
- `frontend/vite.config.ts` dev server 默认 `3000`。
- `scripts/dev.sh` 启动 Vite 时使用 `5173`。
- 多份文档仍写 `38080`、`8080`、`3000`。

**影响**

- 新人按不同文档启动会访问错地址。
- 桌面端、Web 端、部署文档的排障成本很高。
- CI release notes 写错地址会直接误导用户。

**建议**

- 选定一套正式端口：例如后端 `12345`、RPC `8888`、前端 dev `5173`。
- 把 `pkg/configs/loader.go` 默认值、`conf/config.yaml`、Vite、dev 脚本、Docker/K8s 文档统一。
- 建一个短文档 `docs/development.md` 或更新 README，只保留一个“本地开发启动方式”。
- 对历史 dev-notes 标记“历史记录，不保证当前可用”，避免被当成当前文档。

**优先级：P0**

### 3. API JSON 字段命名没有统一

**表现**

- `AGENTS.md` 写后端 API DTO 应使用 camelCase。
- 实际 `biz/model/api/` 中同时存在 snake_case 和 camelCase，例如 `repo_id`、`created_at`、`is_default`，也有 `ruleId`、`createdAt`、`autoCommit`。
- 前端类型中也大量使用 `created_at`、`updated_at`、`rule_id`、`is_default`。

**影响**

- 前端容易读不到字段，尤其是新接口和旧接口混用时。
- DTO 命名规则靠人工记忆，后续每个功能都可能重复踩坑。
- 如果未来开放 API，字段规范会很难向后兼容。

**建议**

- 先做字段命名决策：全量保持 snake_case，或新 API 统一 camelCase。不要两套规则同时新增。
- 短期建议保持兼容：后端 response DTO 通过 adapter 层输出稳定字段，不直接暴露 PO。
- 中期引入 OpenAPI 或 proto 生成 TypeScript 类型，前端不再手写接口类型。
- 为 API 字段命名加测试或 lint，至少禁止同一领域内新老命名混用。

**优先级：P0**

### 4. 数据库迁移策略不安全

**表现**

- `biz/dal/db/init.go` 中先检查一批表是否存在，如果都存在就直接跳过 schema migration。
- `AutoMigrate` 列表里包含新表，但“所有旧表存在”时新增字段、索引、部分新表可能被跳过。
- 数据迁移函数和建表逻辑混在初始化流程里，缺少版本号和幂等迁移记录。

**影响**

- 老用户升级时可能缺字段或缺表。
- 某些环境第一次启动正常，升级后才出现运行时 SQL 错误。
- 无法知道数据库处在哪个 schema 版本。

**建议**

- 引入显式 migration 版本表，例如 `schema_migrations`。
- 每次 schema 变化写独立迁移，包含 `up`、幂等检查、回滚策略或不可回滚说明。
- `AutoMigrate` 只用于开发环境或首次初始化，生产升级走 migration。
- 给关键迁移补集成测试：从旧 schema fixture 升级到当前 schema。

**优先级：P0**

### 5. 生成代码边界不够硬

**表现**

- 约定里说 `biz/model/**` 和部分 `biz/router/**` 是生成代码，不应手改。
- 当前工作区已有 `biz/model/maintenance/maintenance.pb.go`、`biz/router/maintenance/*.go` 等生成文件处于修改状态。
- `Makefile hz-gen` 与 `script/gen.sh` 的生成目录参数不完全一致，一个使用 `biz/router/hz`，另一个使用 `biz/router`。

**影响**

- 手改生成代码容易被下一次 `make gen` 覆盖。
- 不同生成入口产生不同目录结构，导致评审时无法判断哪些改动是业务逻辑，哪些是生成结果。

**建议**

- 统一唯一生成入口：只保留 `make gen`，其他脚本调用它。
- 为生成文件加 CI 检查：执行 codegen 后 `git diff --exit-code`。
- 明确哪些 router 文件允许手写，例如只允许 `register.go`、`custom_routes.go`。
- 业务逻辑必须放 handler/service，不在 generated router/model 中补逻辑。

**优先级：P1**

### 6. 全局状态和启动初始化过重

**表现**

- `configs.GlobalConfig`、`db.DB`、LLM provider registry、sync cron、stats、audit、queue 等都在启动时全局初始化。
- `cmd/server/main.go:initResources()` 同时负责配置、数据库、规则、迁移、加密、定时任务、队列。
- 桌面端和 Web/RPC 端共享部分后端能力，但入口初始化路径不够清晰。

**影响**

- 单元测试难隔离，需要依赖全局状态重置。
- 某个子系统初始化失败会影响整个服务，即使当前模式不需要它。
- 后续想拆 worker、HTTP、RPC、desktop 会越来越困难。

**建议**

- 引入 `App` 或 `Container`，把 config、db、services、jobs 显式挂进去。
- 按运行模式初始化：HTTP 不需要的 worker 不启动，RPC 不需要的静态资源不加载。
- 把 `initResources()` 拆为 `LoadConfig`、`OpenDB`、`RunMigrations`、`BuildServices`、`StartBackgroundJobs`。
- 测试中使用独立 container 和临时数据库，避免污染全局。

**优先级：P1**

### 7. 测试入口不够可靠

**表现**

- `go test ./...` 超过 2 分钟未结束，已完成的多数单元包通过，但全量命令无法快速稳定返回。
- 集成测试目录存在，但外部依赖、超时、跳过条件不明确。
- 前端类型检查没有进入 Makefile，也没有在已检查的 release workflow 中单独作为质量门禁。

**影响**

- 开发者无法把 `make test` 当成提交前可靠信号。
- CI 如果没有跑全量测试，发布质量依赖人工。
- 集成测试卡住时定位成本高。

**建议**

- 区分 `make test-unit`、`make test-integration`、`make test-all`。
- 集成测试必须有 `testing.Short()` 跳过策略、明确环境变量、统一超时。
- CI 至少跑 `go test -short ./...`、`cd frontend && npm ci && npx vue-tsc --noEmit`。
- 对需要真实 Git 仓库、Redis、数据库的测试使用容器或 fixture。

**优先级：P1**

### 8. 文档体系重复且过期内容较多

**表现**

- 根目录有 `README.md`、`DEPLOY.md`、`AGENT.md`、`AGENTS.md`、`CLAUDE.md`。
- `docs/`、`docs-site/`、`docs/dev-notes/` 同时存在，且多处端口、路径、功能描述互相矛盾。
- README 宣称大量能力，但部分部署和 release 文案仍指向旧端口或旧路径。

**影响**

- 文档越多越难维护，用户不确定哪份可信。
- 开发者容易根据过期文档做错误变更。

**建议**

- 定义文档分层：README 只放快速开始；`docs/` 放当前稳定文档；`docs/dev-notes/` 标记为历史记录。
- 删除或合并 `AGENT.md`、`CLAUDE.md` 与 `AGENTS.md` 中重复且冲突的内容。
- 给端口、构建、部署文档做一次全局校准。

**优先级：P1**

### 9. 前端模块多，但类型和 API 契约靠手写维护

**表现**

- `frontend/src/api/modules/` 每个领域手写一套 API。
- `frontend/src/types/` 与 API modules 中都有类型定义。
- 后端 DTO 命名不统一，前端需要自己猜字段名。

**影响**

- 字段变更需要多处同步，漏改概率高。
- 同一实体在不同页面可能出现多个 TypeScript 定义。

**建议**

- 建立后端 API schema 生成流程，输出 TypeScript client 或类型。
- API module 只保留调用函数，不重复定义实体结构。
- 对历史接口做兼容层，页面层只接触统一后的类型。

**优先级：P2**

### 10. 业务模块规模开始失控

**表现**

- `biz/service/git/git_maintenance.go` 约 700 行。
- `frontend/src/components/spec/SpecEditor.vue` 约 570 行。
- Spec 子系统存在 rule lint、AI lint、legacy validate 三套实现。

**影响**

- 新需求容易继续堆到大文件里。
- 规则变更、AI 变更、UI 状态变更互相影响。
- Review 难判断改动边界。

**建议**

- 先按职责拆文件，不急着大改架构。
- Spec：把文件 IO、规则 lint、AI lint、legacy validate、formatter 明确分层，并给每层定义输入输出。
- 前端 SpecEditor：拆成 toolbar、file tree、editor pane、lint panel、AI assistant panel，状态通过 composable 管理。
- Git maintenance：把 health report、gc、bfg、ignore suggestion、AI analysis 拆成独立 service。

**优先级：P2**

### 11. 构建产物和嵌入资源容易干扰本地分析

**表现**

- `pkg/embed/`、`public/`、`frontend/dist/` 被 gitignore，但本地仍存在较大构建产物。
- 普通 `rg` 或文件扫描如果不排除这些目录，会扫到压缩后的前端 bundle，输出非常噪声。

**影响**

- 代码搜索结果被构建产物污染。
- 代码评审和自动化分析容易误判。

**建议**

- 在开发文档中要求搜索时排除 `pkg/embed`、`public`、`frontend/dist`。
- 考虑新增 `.ignore` 或 `.rgignore`。
- `make clean-frontend` 可同时清理 `pkg/embed/public` 和 `pkg/embed/docs`，保持工作区干净。

**优先级：P2**

### 12. 部署和产品形态边界不清

**表现**

- 同时支持 Web 服务、RPC、Wails desktop、Docker、K8s、docs-site。
- 桌面文档中仍有旧式“打开 localhost:38080”的描述，而当前 Wails 入口是内嵌前端资源。
- release workflow 和 desktop workflow 各自维护构建逻辑。

**影响**

- 每次改前端、配置、端口，都需要人工同步多个形态。
- 发布链路越多，越容易出现某个形态长期不可用。

**建议**

- 明确产品优先级：Web 服务和桌面端是否同等维护。
- 对每种形态定义最小验收：Web 二进制启动并能访问 `/`；desktop 能打开主窗口；Docker healthcheck 能通过。
- 共用构建脚本，不在 workflow 中复制逻辑。

**优先级：P2**

## 推荐改造路线

### 第一阶段：先止血，保证开发和发布一致

目标：让任何人按 README、Makefile、CI 得到同一个结果。

任务：

- 统一端口：后端 `12345`、RPC `8888`、前端 dev `5173`，或另选一套但只能有一套。
- 修正 release workflow 的 Go 版本和 `CGO_ENABLED`。
- 把 `npx vue-tsc --noEmit` 加进 Makefile 和 CI。
- 增加 `make test-unit`，默认只跑稳定单元测试。
- 增加 `.rgignore`，排除构建产物。

验收：

- `make build-full` 通过。
- `make test-unit` 在 1 分钟内稳定完成。
- README 中的启动命令可直接访问正确端口。

### 第二阶段：统一 API 契约

目标：减少前后端字段错配。

任务：

- 盘点所有 `biz/model/api` 的 JSON 字段，确定 snake_case 或 camelCase。
- 新增 API DTO 命名规范，不再直接复用 PO 输出。
- 引入类型生成，至少让前端从统一 schema 生成基础类型。
- 给高频实体补契约测试：repo、sync、review、spec、provider、credential。

验收：

- 同一领域不再同时出现 `created_at` 和 `createdAt`。
- 前端页面不再重复定义同一 DTO。
- 新增接口必须有类型或 schema 来源。

### 第三阶段：治理数据库和初始化

目标：升级可靠，测试可隔离。

任务：

- 新增 schema migration 版本表。
- 将已有 `AutoMigrate` 拆成首次初始化和版本迁移。
- 将 `initResources()` 拆成可测试的初始化步骤。
- 服务层逐步从全局 `db.DB` 迁移到显式依赖注入。

验收：

- 从旧数据库 fixture 升级到当前版本测试通过。
- 单元测试可使用临时 DB，不依赖全局状态。

### 第四阶段：模块化重构

目标：降低后续功能开发成本。

任务：

- 拆分大文件：SpecEditor、git maintenance、spec lint/validate。
- 对 spec 子系统定义统一 `LintResult`、`LintIssue`、`FixSuggestion` 模型。
- 后台任务、cron、queue 独立为 worker 入口或显式服务。
- 清理过期 dev-notes 和重复文档。

验收：

- 主要大文件降到合理规模，每个文件职责单一。
- 新增 spec 规则或 AI 能力不需要同时改三套逻辑。
- 文档入口只剩 README + 当前 docs。

## 建议新增的工程规则

- 所有端口只允许从配置读取，文档中引用必须同步更新。
- 不允许手改生成代码；改 proto 后必须运行 codegen 并提交完整生成结果。
- 后端 API response 不直接返回 GORM PO。
- 前端类型来自 API schema 或集中类型文件，页面内不重复定义接口实体。
- 集成测试默认不在 `go test ./...` 中阻塞，除非依赖和超时明确。
- CI 至少包含：Go 单元测试、前端类型检查、构建、生成代码 diff 检查。

## 可拆分任务清单

- P0：修正 release workflow 中 `CGO_ENABLED` 与 Go 版本。
- P0：统一所有当前文档和脚本中的 HTTP/RPC/前端端口。
- P0：确定 API 字段命名规范并冻结新接口规则。
- P0：修复数据库迁移“表存在就跳过”的升级风险。
- P1：统一 `make gen` 和脚本生成目录。
- P1：增加 `make test-unit`、`make test-integration`、`make typecheck-frontend`。
- P1：给集成测试增加超时、依赖检查和 short 模式跳过。
- P1：合并或标记过期文档。
- P2：引入前端 API 类型生成。
- P2：拆分 SpecEditor 和 git maintenance 大文件。
- P2：增加 `.rgignore` 或同类搜索排除配置。
