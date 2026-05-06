# 项目复查后的后续改进建议

> 生成时间：2026-05-04  
> 背景：在第一轮《项目不合理点与后续改造建议》之后，项目已经做了一批修正。本文件只记录本次复查发现的剩余问题和下一步建议。

## 当前状态

这轮改动已经把第一批工程问题基本收口：

- `[x]` 服务端发布构建改为 `CGO_ENABLED=0`，桌面/Wails 构建单独使用 `CGO_ENABLED=1`。
- `[x]` `make test-unit` 已排除 `tests/integration`。
- `[x]` `tests/integration` 已在 `testing.Short()` 下跳过。
- `[x]` release workflow 已增加 `Quality Gates`，发布构建依赖 `make test-unit` 和 `make typecheck-frontend`。
- `[x]` 当前主要文档端口已收敛到 HTTP `12345`、RPC `8888`、前端 dev `5173`。
- `[~]` 第一份评审文档已标记为历史快照，但其中正文仍保留初始发现。
- `[x]` `make test-integration` 已恢复稳定通过。
- `[~]` API 字段命名仍有历史混用，但已增加新增 snake_case 冻结门禁。
- `[~]` 数据库 migration 已有版本表和数据迁移记录，schema 升级仍需继续拆分。
- `[x]` 生成代码边界已增加 CI diff 检查。

后续重点应从 P0 止血转向契约治理和升级可靠性。

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

结果：通过，并且不运行 `tests/integration`。

### 集成测试入口

命令：

```bash
go test -v -short ./tests/integration -count=1 -timeout=20s
```

结果：通过，快速 skip，不启动 Hertz server。

补充：普通 `make test-integration` 已通过。集成测试会反复启动 Hertz server，测试客户端已禁用 HTTP keep-alive，避免 cleanup 时 `h.Shutdown()` 等待空闲连接导致超时。

### 集成测试完整入口

命令：

```bash
timeout 150s make test-integration
```

结果：通过。

## 已完成的高优先级项

### P0：修正 CGO 策略

**状态：已完成**

- `Makefile` 已移除全局 `export CGO_ENABLED=1`。
- `.github/workflows/release.yml` 服务端发布包已使用 `CGO_ENABLED=0`。
- README、AGENTS、CLAUDE 已改为“服务端无需 CGO，桌面/Wails 需要 CGO”。

**验收**

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ./cmd/server
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build ./cmd/server
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build ./cmd/server
```

三条都通过。

### P0：让 `make test-unit` 真正只跑单元测试

**状态：已完成**

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

**状态：已完成**

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

**状态：已完成**

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

**状态：基本完成**

当前有效文档已统一使用：

- HTTP：`12345`
- RPC：`8888`
- 前端 dev：`5173`

`conf/config_test.yaml` 保留 `38080`，并已注明是测试专用端口。`docs/dev-notes/` 属于历史记录，可以保留旧端口。

**验收**

```bash
rg -n '38080|localhost:8080|localhost:3000' README.md DEPLOY.md docs .github conf
```

结果中只允许出现历史 dev-notes 或明确标注的测试配置。

### P1：更新第一份评审文档状态

**状态：已完成**

第一份文档开头已明确标记为历史快照，当前状态以本文档为准。

**验收**

读者不会再把已修复的问题误认为当前仍存在。

## 中期继续推进

### API 字段命名治理

当前 `biz/model/api` 中仍有大量 snake_case JSON tag，同时也存在 camelCase。直接大规模重命名会破坏前端兼容，所以先采用冻结策略。

**状态：已完成新增约束**

- `biz/model/api/json_tag_baseline.txt` 记录当前手写 API DTO 中已有的 snake_case JSON tag。
- `biz/model/api/json_tag_contract_test.go` 会阻止新增 snake_case JSON tag。
- `make check-api-json-tags` 提供显式本地检查入口。

推荐路线：

1. 决策：长期规范使用 camelCase。
2. 冻结：新增接口已通过测试禁止继续引入 snake_case。
3. 兼容：旧接口保留，新增 v2 DTO 或 adapter。
4. 生成：引入 OpenAPI/proto 到 TypeScript 的生成流程。
5. 收敛：按模块逐步减少 `json_tag_baseline.txt` 中的历史项。

### 数据库 migration 版本化

**状态：已完成基础版本表**

- 新增 `schema_migrations` 表。
- 现有 repo-provider binding 数据迁移已纳入版本记录，避免每次启动重复执行。
- 新增迁移 runner 单元测试，验证迁移只执行一次。

下一步建议：

- 新 schema 变化继续拆成独立 migration step。
- 逐步把 `AutoMigrate` 收敛为首次初始化兜底，而不是长期升级机制。
- 用旧 schema fixture 测试升级路径。

### 生成代码边界治理

**状态：已完成基础门禁**

- `Makefile` 新增 `make check-generated`。
- `.github/workflows/ci.yml` 的 `generated-code-check` 已从目录存在检查改为执行真实 codegen 并检查 git diff。
- `script/gen.sh` 和 `make hz-gen` 已统一使用当前 `.hz` 配置里的 `biz/handler`、`biz/router`、`biz/model` 目录，避免生成到旧的 `biz/*/hz` 路径。

后续仍建议继续补一条文档约束：只允许手写 `register.go`、`custom_routes.go`，业务逻辑不要进入 generated router/model。

## 推荐下一步顺序

1. API 字段命名按模块逐步从 baseline 中收敛。
2. 继续拆分 schema migration 和旧库升级 fixture。
3. 继续拆分大文件和全局初始化逻辑。

## 可拆任务清单

- P2：API 字段命名按模块减少 baseline。
- P2：把后续 schema 变更拆成独立 migration step。
- P2：补充生成代码边界文档约束。
