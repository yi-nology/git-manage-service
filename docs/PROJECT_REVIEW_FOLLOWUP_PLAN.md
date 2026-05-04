# 项目复查后的后续改进建议

> 生成时间：2026-05-04  
> 背景：在第一轮《项目不合理点与后续改造建议》之后，项目已经做了一批修正。本文件只记录本次复查发现的剩余问题和下一步建议。

## 复查结论

这轮改动方向是对的：默认端口已经向 `12345` 收敛，release/desktop workflow 的 Go 版本有统一趋势，Makefile 增加了 `test-unit`、`test-integration`、`typecheck-frontend`，并新增了 `.rgignore` 排除构建产物。

但现在还有几个关键点没有闭环：

1. `CGO_ENABLED=1` 的判断很可能是错误方向，服务端发布包更适合 `CGO_ENABLED=0`。
2. `make test-unit` 仍会跑到 `tests/integration`，所以还不是稳定快速的单元测试入口。
3. release workflow 还没有把 Go 单元测试和前端类型检查作为发布门禁。
4. 当前文档里仍有大量旧端口 `38080`、`8080`、`3000`。
5. API 字段命名仍然是 snake_case 和 camelCase 混用。

## 本次验证结果

### 前端类型检查

命令：

```bash
make typecheck-frontend
```

结果：通过。

### 服务端无 CGO 跨平台构建

命令：

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /tmp/gms-linux-amd64-nocgo ./cmd/server
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o /tmp/gms-darwin-amd64-nocgo ./cmd/server
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o /tmp/gms-windows-amd64-nocgo.exe ./cmd/server
```

结果：全部通过。

说明：当前项目使用 `github.com/glebarez/sqlite`，依赖链包含 `modernc.org/sqlite`，服务端二进制不必默认强制 CGO。

### CGO 跨平台构建

命令：

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build -o /tmp/gms-linux-amd64-cgo ./cmd/server
```

结果：失败。

原因：当前机器是 Darwin/arm64，开启 CGO 后跨编译 Linux/amd64 需要匹配的 C 交叉编译工具链。GitHub Actions 如果在 Ubuntu runner 上用同样方式构建 darwin/windows，也会遇到类似问题，除非额外配置跨平台 C 工具链。

### 单元测试入口

命令：

```bash
timeout 90s make test-unit
```

结果：超时终止。

原因：`make test-unit` 当前实际执行 `go test -v -short ./...`，仍然包含 `tests/integration`。

对比命令：

```bash
go list ./... | rg -v '/tests/integration$' | xargs go test -short
```

结果：通过，约 8 秒完成。

### 集成测试入口

命令：

```bash
go test -v -short ./tests/integration -count=1 -timeout=20s
```

结果：失败。

现象：`tests/integration` 没有在 `testing.Short()` 下跳过，并且在 Hertz server shutdown/连接等待处超时。

## 需要优先处理的问题

### P0：修正 CGO 策略

**现状**

- `Makefile` 强制 `export CGO_ENABLED=1`。
- `AGENTS.md`、`README.md`、`CLAUDE.md` 仍写 SQLite 需要 CGO。
- `.github/workflows/release.yml` release matrix 设置 `CGO_ENABLED=1`。

**问题**

服务端发布包在 `CGO_ENABLED=0` 下已经可以跨平台构建；继续强制 CGO 会让 release matrix 的跨平台构建变复杂，并可能失败。

**建议修改**

- `Makefile`：移除全局 `export CGO_ENABLED=1`，或者只在确实需要的 desktop/Wails 目标里设置。
- `.github/workflows/release.yml`：服务端二进制构建改为 `CGO_ENABLED=0`。
- 文档：把“SQLite 必须 CGO”改为“服务端默认使用 pure Go SQLite 驱动，可无 CGO 构建；桌面端按 Wails 平台依赖准备环境”。

**验收**

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./cmd/server
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build ./cmd/server
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ./cmd/server
```

三条都通过。

### P0：让 `make test-unit` 真正只跑单元测试

**现状**

`make test-unit` 当前执行：

```bash
go test -v -short ./...
```

这会包含 `tests/integration`。

**建议修改**

```makefile
test-unit:
	go list ./... | grep -v '/tests/integration$$' | xargs go test -v -short

test-integration:
	go test -v ./tests/integration -count=1 -timeout=2m
```

如果希望兼容没有 `grep` 的环境，可以改成脚本实现，但当前项目主要是类 Unix 开发环境，Makefile 里直接写可以先落地。

**验收**

```bash
timeout 90s make test-unit
```

应稳定通过，目标耗时控制在 1 分钟以内。

### P0：集成测试支持 short 模式跳过

**现状**

`go test -short ./tests/integration` 仍会运行集成测试。

**建议修改**

在 `tests/integration/test_helper.go` 的 `SetupSuite(t)` 开头增加：

```go
if testing.Short() {
	t.Skip("skip integration tests in short mode")
}
```

同时建议在 `go test ./tests/integration` 中保留明确超时：

```bash
go test -v ./tests/integration -count=1 -timeout=2m
```

**验收**

```bash
go test -short ./tests/integration
```

应快速显示 skip，而不是启动 Hertz server。

### P1：release workflow 增加质量门禁

**现状**

release workflow 现在安装依赖后直接构建前端和二进制，没有先跑测试和类型检查。

**建议修改**

增加独立 `quality` job：

```yaml
quality:
  name: Quality Gates
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v5
      with:
        go-version: '1.25'
        cache: true
    - uses: actions/setup-node@v4
      with:
        node-version: '20'
        cache: 'npm'
        cache-dependency-path: frontend/package-lock.json
    - run: go mod download
    - run: make test-unit
    - working-directory: frontend
      run: npm ci
    - run: make typecheck-frontend
```

然后让 release build job：

```yaml
needs: quality
```

**验收**

手动触发 release workflow 时，测试或类型检查失败会阻止发布构建。

### P1：继续统一当前文档端口

**现状**

代码和部分文档已使用：

- HTTP：`12345`
- RPC：`8888`
- 前端 dev：`5173`

但仍有当前文档写旧值，例如：

- `DEPLOY.md` 中仍有 `localhost:8080`
- `docs/api.md` 中仍有 `localhost:38080`
- `docs/configuration.md` 中仍有 `38080`
- `docs/deployment/*.md` 中仍有 `38080`
- `.github/workflows/desktop.yml` release 文案仍写桌面端后端端口 `38080`
- `conf/config.yaml` 注释仍写 `Default: 8080`

**建议修改**

- 当前有效文档统一改成 `12345`。
- `docs/dev-notes/` 可以保留历史端口，但在目录 README 或每篇开头加“历史记录，不保证当前可用”。
- `conf/config_test.yaml` 如果是测试专用端口，可以保留 `38080`，但建议加注释说明它不是默认运行端口。

**验收**

```bash
rg -n '38080|localhost:8080|localhost:3000' README.md DEPLOY.md docs .github conf
```

结果中只允许出现历史 dev-notes 或明确标注的测试配置。

### P1：更新第一份评审文档状态

**现状**

`docs/PROJECT_REVIEW_AND_IMPROVEMENT_PLAN.md` 仍记录第一轮发现，例如 release 的 `CGO_ENABLED=0` 问题、loader 默认端口 `8080` 问题。部分内容已经被修改，文档变成“历史快照”和“当前待办”混在一起。

**建议修改**

把第一份文档改成 checklist 状态：

- `[x]` 已完成
- `[~]` 部分完成
- `[ ]` 未完成

或者在开头明确说明：第一份是初始评审快照，当前状态以本文件为准。

**验收**

读者不会再把已修复的问题误认为当前仍存在。

## 中期继续推进

### API 字段命名治理

当前 `biz/model/api` 中仍有大量 snake_case JSON tag，同时也存在 camelCase。建议先不要直接大规模重命名，因为会破坏前端兼容。

推荐路线：

1. 决策：正式选择 snake_case 或 camelCase 作为长期规范。
2. 冻结：新增接口必须使用选定规范。
3. 兼容：旧接口保留，新增 v2 DTO 或 adapter。
4. 生成：引入 OpenAPI/proto 到 TypeScript 的生成流程。
5. 检查：增加 lint 或测试，防止新接口继续混用。

### 数据库 migration 版本化

第一轮提到的数据库迁移问题还没有从根上解决。下一步建议：

- 新增 `schema_migrations` 表。
- 每次 schema 变化写独立迁移。
- `AutoMigrate` 只用于首次初始化或开发环境。
- 用旧 schema fixture 测试升级路径。

### 生成代码边界治理

当前仍建议继续做：

- 统一唯一 codegen 入口。
- CI 执行 codegen 后检查 `git diff --exit-code`。
- 明确只允许手写 `register.go`、`custom_routes.go`。

## 推荐下一步顺序

1. 改 `CGO_ENABLED` 策略，修 release workflow 和文档。
2. 改 `make test-unit`，排除 `tests/integration`。
3. 给 `tests/integration` 加 short skip 和明确 timeout。
4. release workflow 加 `quality` job。
5. 清理当前文档端口。
6. 更新第一份评审文档的状态说明。
7. 再进入 API 字段命名和 migration 版本化治理。

## 可拆任务清单

- P0：`release.yml` 服务端二进制构建改 `CGO_ENABLED=0`。
- P0：删除或改写 Makefile/README/AGENTS/CLAUDE 中“必须 CGO”的描述。
- P0：`make test-unit` 排除 `tests/integration`。
- P0：`tests/integration` 在 short mode 下跳过。
- P1：release workflow 增加 `make test-unit` 和 `make typecheck-frontend`。
- P1：统一当前文档端口到 `12345/8888/5173`。
- P1：更新第一份评审文档为状态化 checklist 或历史快照说明。
- P2：API 字段命名规范冻结并逐步治理。
- P2：数据库 migration 版本化。
- P2：生成代码边界 CI 检查。
