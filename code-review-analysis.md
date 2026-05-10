# Code Review 模块分析

## 架构概览

```
前端 → Handler → Service → DAO → SQLite
         ↓           ↓
    Provider    Rule Engine + LLM
```

## 核心功能

### 1. 审查任务管理

- 创建/获取/列表/重试审查任务 (`biz/handler/review/review_service.go`)
- 支持两种创建方式：按仓库 Key 和按 Provider
- 任务状态流转：pending → running → success/blocked/failed

### 2. 规则引擎

`biz/service/codereview/rule_engine.go`

内置 5 条规则：

| 规则 | 文件 | 检测内容 |
|------|------|----------|
| SecretRule | `secret_rule.go` | 密码、API Key、Token 等敏感信息 |
| ProtectedFileRule | `protected_file_rule.go` | 受保护文件修改 |
| DiffSizeRule | `diff_size_rule.go` | 过大的 diff |
| MigrationRule | `migration_rule.go` | 数据库迁移 |
| TestRequiredRule | `test_required_rule.go` | 缺少测试 |

### 3. LLM 审查

`biz/service/codereview/review_service_publish.go`

- 集成 AI 进行语义级代码审查
- 支持重试机制（3次，指数退避）
- RAG 上下文增强（`retrieveRAGContext`）
- 自定义规则注入（`buildSystemPromptWithRules`）

### 4. 风险评估

`biz/service/codereview/aggregator.go`

- 5 级风险：critical > high > medium > low > info
- 聚合逻辑：critical 直接阻断，3个 medium 升级为 high
- 支持 `block_on_high` 配置阻断合并

### 5. 结果发布

`biz/service/codereview/review_service_publish.go`

- 生成 Markdown 摘要评论
- 创建内联讨论（critical/high/medium）
- 清理旧评论
- 更新 Commit Status

## 数据模型

| 模型 | 表名 | 说明 |
|------|------|------|
| ReviewTask | `review_tasks` | 审查任务主表 |
| ReviewFinding | `review_findings` | 发现的问题 |
| ReviewComment | `review_comments` | 发布的评论 |
| ReviewRepoConfig | `review_repo_configs` | 仓库审查配置 |
| MergeCheckResult | `merge_check_results` | 合并检查结果 |

## API 端点

```
POST   /api/v1/reviews/tasks              # 创建任务
GET    /api/v1/reviews/tasks/:id           # 获取任务
GET    /api/v1/reviews/tasks               # 任务列表
GET    /api/v1/reviews/tasks/:id/findings  # 发现项
POST   /api/v1/reviews/tasks/:id/retry     # 重试
GET    /api/v1/merge-checks                # 合并检查
GET    /api/v1/reviews/stats               # 统计数据
GET    /api/v1/review/remote-config/...    # 远端配置
```

## 前端页面

| 页面 | 说明 |
|------|------|
| `ReviewDashboardPage.vue` | 统计仪表盘 |
| `ReviewTaskListPage.vue` | 任务列表 |
| `ReviewTaskDetailPage.vue` | 任务详情 |
| `ReviewConfigPage.vue` | 配置管理 |
| `MergeCheckWidget.vue` | 合并检查组件 |

## 关键文件清单

### 后端

```
biz/handler/review/review_service.go        # Handler 层
biz/service/codereview/review_service.go    # Service 主逻辑
biz/service/codereview/review_service_config.go   # 配置管理
biz/service/codereview/review_service_publish.go  # LLM 调用与评论发布
biz/service/codereview/rule_engine.go       # 规则引擎
biz/service/codereview/aggregator.go        # 结果聚合
biz/service/codereview/policy.go            # 策略应用
biz/service/codereview/comment_builder.go   # 评论构建
biz/service/codereview/secret_rule.go       # 敏感信息规则
biz/service/codereview/diff_size_rule.go    # Diff 大小规则
biz/service/codereview/protected_file_rule.go  # 受保护文件规则
biz/service/codereview/migration_rule.go    # 迁移规则
biz/service/codereview/test_required_rule.go   # 测试要求规则
biz/service/codereview/review_stats.go      # 统计分析
biz/dal/db/review_task_dao.go               # 任务 DAO
biz/dal/db/review_finding_dao.go            # 发现项 DAO
biz/dal/db/review_comment_dao.go            # 评论 DAO
biz/dal/db/review_repo_config_dao.go        # 仓库配置 DAO
biz/model/po/review_task.go                 # 任务模型
biz/model/po/review_finding.go              # 发现项模型
biz/model/po/review_comment.go              # 评论模型
biz/model/po/review_repo_config.go          # 仓库配置模型
biz/model/api/review.go                     # API DTO
biz/router/review/review.go                 # 路由注册
```

### 前端

```
frontend/src/api/modules/review.ts          # API 接口
frontend/src/views/review/ReviewDashboardPage.vue   # 仪表盘
frontend/src/views/review/ReviewTaskListPage.vue    # 任务列表
frontend/src/views/review/ReviewTaskDetailPage.vue  # 任务详情
frontend/src/views/review/ReviewConfigPage.vue      # 配置页
frontend/src/views/review/MergeCheckWidget.vue      # 合并检查组件
```

## 潜在改进点

1. **前端 API 字段命名不一致**：后端用 camelCase，前端 DTO 用 snake_case（如 `repo_id` vs `repoId`）
2. **缺少 webhook 自动触发**：目前主要靠手动触发，可集成 GitLab/GitHub webhook
3. **规则配置化**：规则硬编码，可考虑数据库配置
4. **LLM 响应解析健壮性**：已有 JSON 修复逻辑，但可进一步增强
