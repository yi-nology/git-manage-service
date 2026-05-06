# Git Manage Service 待改进问题清单

## P0：优先处理

### 1. 文件路径边界不安全

部分接口直接把用户传入路径拼到仓库路径下，例如 spec 文件读取、写入、删除，以及 worktree 文件读取。

风险：

- 可能通过 `../` 访问仓库外文件
- 可能误删或覆盖非仓库文件
- 后续功能越多，路径漏洞越容易扩散

建议：

- 增加统一 helper：`Clean + Abs + Rel` 校验
- 所有仓库内文件操作必须确认目标路径仍在 `repoPath` 下
- 对读、写、删、patch、spec、worktree 文件接口统一接入

---

### 2. 缺少认证和权限模型

当前系统基本没有真正的登录、token、角色、权限控制。审计日志操作者也固定为 `system`。

风险：

- 服务一旦暴露到局域网或公网，任何人都可能操作仓库、凭证、SSH key
- 审计日志无法追踪真实操作者
- 高风险操作无法做权限隔离

建议：

- 增加基础认证：本地 token / admin password / session
- 区分只读、写入、管理员、高风险操作权限
- 审计日志记录真实 operator
- 默认绑定 `127.0.0.1`，明确开启远程访问需要配置

---

### 3. 密钥加密方案偏弱

当前 `ENCRYPTION_KEY` 未配置时会使用硬编码默认 key，并且加密模式缺少认证校验。

风险：

- 默认 key 等于明文保护很弱
- 密文被篡改时不容易可靠发现
- 后续 key 轮换困难

建议：

- 生产环境强制配置 `ENCRYPTION_KEY`
- 改用 AES-GCM 或 XChaCha20-Poly1305
- 密文加版本前缀，支持后续 key rotation
- 启动时校验 key 长度和格式

---

### 4. 高风险 Git 操作保护不足

仓库瘦身、历史重写、强推、批量清理等操作风险很高，当前保护还不够硬。

风险：

- 误重写历史
- 误更新 tag
- 未提交内容处理不完整
- 命令执行失败后仓库状态难恢复

建议：

- 所有历史重写操作必须支持 dry-run
- 执行前自动创建备份 ref 或 bundle
- 要求用户二次确认影响范围
- 统一封装 Git 命令执行，强制设置 `Dir`
- 增加失败恢复策略和集成测试

---

## P1：中期治理

### 5. 数据库 migration 仍依赖大规模 AutoMigrate

虽然已经有 `schema_migrations`，但启动时仍然对大量表执行 `AutoMigrate`。

问题：

- 老版本升级路径不够明确
- schema 变化缺少逐步演进记录
- 开发环境和真实升级环境可能表现不同

建议：

- 新增字段、索引、表必须写独立 migration step
- `AutoMigrate` 只作为首次初始化兜底
- 增加旧 schema fixture 升级测试
- 每个 migration 使用稳定版本号，例如 `YYYYMMDDNN_desc`

---

### 6. API 字段命名仍有历史混用

当前新增 snake_case 已被门禁限制，但历史 API DTO 里仍同时存在 snake_case 和 camelCase。

问题：

- 前端容易字段错配
- 类型定义需要人工同步
- API 长期对外会难以演进

建议：

- 明确长期标准为 camelCase
- 保留旧接口兼容
- 新接口只使用 camelCase
- 引入 OpenAPI / proto 生成 TypeScript 类型
- 逐模块减少 `json_tag_baseline.txt` 中的历史项

---

### 7. 全局状态和启动初始化过重

启动时一次性初始化 config、DB、lint rules、migration、encryption、cron、stats、audit、queue 等。

问题：

- HTTP / RPC / desktop 模式边界不清晰
- 单元测试难隔离
- 子系统失败容易影响整个服务
- 后续拆 worker 或后台任务会困难

建议：

- 引入 `App` / `Container`
- 按运行模式初始化依赖
- 拆分 `LoadConfig`、`OpenDB`、`RunMigrations`、`BuildServices`、`StartJobs`
- 测试中使用独立 container 和临时 DB

---

### 8. 前端请求取消处理不合理

请求取消时返回永不 resolve 的 Promise。

问题：

- loading 状态可能永远不结束
- 调用方无法感知取消
- 排查异步状态问题困难

建议：

- 返回明确的 canceled error
- 在 composable 或页面层统一忽略取消错误
- loading 状态统一放在 `finally` 中收敛

---

## P2：长期优化

### 9. 前端和后端类型契约靠手写维护

前端 API module 和 types 中存在大量手写类型。

问题：

- 后端 DTO 变更容易漏改前端
- 同一实体可能出现多个 TS 定义
- 字段命名过渡期更容易出错

建议：

- 从后端 schema 生成 TS client/types
- API module 只保留调用逻辑
- 页面层只使用统一后的领域类型

---

### 10. 大文件和大组件较多

后端和前端都有多个 400-800 行文件，例如 AI service、repo detail page、sync task page、spec editor 等。

问题：

- 修改成本高
- 单测难写
- 业务规则分散
- 新人理解困难

建议：

- 按职责拆分 service
- 前端页面拆成容器组件、业务 composable、展示组件
- 高风险逻辑先补测试再拆
- 避免为了拆而拆，优先拆频繁变更和容易出错的模块

---

### 11. AI 功能链路还不够闭环

AI 面板、反馈、草案应用、任务语义映射还没有完全闭环。

问题：

- 有些页面更像固定动作入口，不是真正可追问助手
- 用户反馈如果没有持久化，无法改进结果
- 草案生成和应用之间缺少统一交互

建议：

- 区分 explain / suggest / draft / diagnose 四类任务
- 后端返回 invocation id
- 前端反馈写回后端
- 统一“生成草案 -> 用户确认 -> 应用”的交互模型
- 文档补充页面到 AI API 的映射关系

---

## 建议执行顺序

1. 文件路径边界保护
2. 认证和权限模型
3. 密钥加密升级
4. 高风险 Git 操作保护
5. 数据库 migration 收敛
6. API 契约和 TS 类型生成
7. 全局状态重构
8. 前端异步状态和组件拆分
9. AI 功能闭环
