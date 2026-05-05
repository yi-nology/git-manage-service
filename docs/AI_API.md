# AI API 文档

## 概述

本文档描述 Git 管理服务的 AI 功能 API。所有 API 均以 `/api/v1/ai` 为前缀。

## API 列表

---

### 1. 同步失败诊断

分析同步失败的日志，给出根因分析和修复建议。

**Endpoint**: `POST /api/v1/ai/sync/failure`

**请求体**:

```json
{
  "repoKey": "repo/path",
  "logs": "同步日志内容",
  "stderr": "Git 错误输出",
  "currentBranch": "main",
  "trackingBranch": "origin/main",
  "recentActions": ["pull", "merge"]
}
```

**响应**:

```json
{
  "rootCause": "存在未解决的合并冲突",
  "evidence": ["检测到未合并的文件", "Git 状态显示正在合并中"],
  "recommendedActions": ["执行 git status 查看冲突文件", "手动解决冲突后执行 git commit"],
  "canAutoFix": false,
  "riskLevel": "medium",
  "fixDraft": null,
  "raw": "AI 原始响应内容"
}
```

**返回类型**: `AIDiagnosisResponse`

---

### 2. 仓库总览摘要

分析仓库状态，给出健康度摘要和行动建议。

**Endpoint**: `POST /api/v1/ai/repo/summary`

**请求体**:

```json
{
  "repoKey": "repo/path",
  "status": {
    "defaultBranch": "main",
    "branchCount": 5,
    "tagCount": 12,
    "commitCount": 256
  },
  "issues": ["存在脏分支", "有未推送的提交"],
  "pendingChanges": 3
}
```

**响应**:

```json
{
  "summary": "仓库整体健康，有少量待处理项",
  "riskLevel": "low",
  "suggestions": ["清理已合并的分支", "推送本地提交到远程"],
  "actions": [
    {
      "id": "cleanup_branches",
      "label": "清理分支",
      "type": "primary",
      "description": "一键清理已合并的本地分支"
    }
  ],
  "raw": "AI 原始响应内容"
}
```

**返回类型**: `AIAdviceResponse`

---

### 3. 生成 Commit Message

根据 Diff 自动生成符合规范的提交信息。

**Endpoint**: `POST /api/v1/ai/commit/message`

**请求体**:

```json
{
  "repoKey": "repo/path",
  "diff": "Git diff 内容",
  "style": "conventional"
}
```

**样式 (style) 选项**:
- `simple`: 简洁的一句话
- `conventional`: Conventional Commits 格式（默认）
- `detailed`: 包含标题和详细改动列表

**响应**:

```json
{
  "summary": "生成的提交信息",
  "applyContent": "feat(user): 添加用户登录功能\n\n- 实现 JWT 验证\n- 添加登录日志记录",
  "riskLevel": "low",
  "raw": "AI 原始响应内容"
}
```

**返回类型**: `AIDraftResponse`

---

### 4. 代码审查

分析代码变更，发现潜在问题。

**Endpoint**: `POST /api/v1/ai/review`

**请求体**:

```json
{
  "repoKey": "repo/path",
  "diff": "Git diff 内容",
  "changedFiles": ["src/main.go", "src/user.go"],
  "existingFindings": ["规则引擎发现的问题"],
  "language": "go"
}
```

**响应**:

```json
{
  "summary": "发现 2 个需要关注的问题",
  "blocking": [
    {
      "severity": "blocking",
      "category": "security",
      "message": "密码字段未做哈希处理",
      "filePath": "src/user.go",
      "startLine": 42,
      "endLine": 45,
      "suggestion": "使用 bcrypt 对密码进行哈希",
      "confidence": "high"
    }
  ],
  "highRisk": [
    {
      "severity": "high",
      "category": "performance",
      "message": "循环中执行数据库查询",
      "filePath": "src/main.go",
      "startLine": 78,
      "endLine": 85,
      "suggestion": "批量查询替代循环查询"
    }
  ],
  "optional": [],
  "riskLevel": "high",
  "shouldMerge": false,
  "mergeNotes": "建议修复阻塞性问题后再合并",
  "raw": "AI 原始响应内容"
}
```

**返回类型**: `AIReviewResponse`

---

### 5. 审查回复草稿

生成 Code Review 的回复模板。

**Endpoint**: `POST /api/v1/ai/review/reply`

**请求体**:

```json
{
  "repoKey": "repo/path",
  "reviewSummary": "发现 3 个问题，2 个阻塞性问题",
  "reviewerComments": ["代码结构需要优化", "缺少单元测试"],
  "tone": "professional"
}
```

**语气 (tone) 选项**:
- `professional`: 专业正式（默认）
- `friendly`: 友好协作
- `concise`: 简洁直接

**响应**:

```json
{
  "summary": "生成的回复草稿",
  "applyContent": "感谢您的仔细审查，我会尽快修复这些问题...",
  "riskLevel": "low",
  "raw": "AI 原始响应内容"
}
```

**返回类型**: `AIDraftResponse`

---

### 6. 解决冲突

分析冲突内容，给出合并后的代码建议。

**Endpoint**: `POST /api/v1/ai/conflict/resolve`

**请求体**:

```json
{
  "repoKey": "repo/path",
  "conflictDiff": "包含冲突标记的 diff 内容",
  "oursBranch": "main",
  "theirsBranch": "feature/user"
}
```

**响应**:

```json
{
  "summary": "已智能合并两方变更",
  "changeSummary": "保留了 ours 分支的验证逻辑和 theirs 分支的新字段",
  "applyContent": "合并后的完整代码",
  "riskLevel": "medium",
  "references": [
    {
      "type": "conflict",
      "id": "1",
      "label": "user.go 冲突位置"
    }
  ],
  "raw": "AI 原始响应内容"
}
```

**返回类型**: `AIDraftResponse`

---

### 7. 冲突解释

分析冲突原因，给出解决方案建议。

**Endpoint**: `POST /api/v1/ai/conflict/explain`

**请求体**: 同「解决冲突」API

**响应类型**: `AIAdviceResponse`

---

### 8. 分支策略建议

基于仓库现状，推荐分支命名规范和保护策略。

**Endpoint**: `POST /api/v1/ai/branch/rule`

**请求体**:

```json
{
  "repoKey": "repo/path",
  "existingBranches": ["main", "develop", "feature/user-auth"],
  "repoType": "standard"
}
```

**返回类型**: `AIDraftResponse`

---

### 9. Spec 模板生成

根据包信息生成 RPM Spec 文件骨架。

**Endpoint**: `POST /api/v1/ai/spec/template`

**请求体**:

```json
{
  "repoKey": "repo/path",
  "packageName": "my-app",
  "specType": "service",
  "existingSpecContent": "已有的 spec 内容（可选）"
}
```

**返回类型**: `AIDraftResponse`

---

### 10. Spec 重写

按指令重写 Spec 文件的指定部分。

**Endpoint**: `POST /api/v1/ai/spec/rewrite`

**请求体**:

```json
{
  "repoKey": "repo/path",
  "specContent": "完整的 spec 文件内容",
  "sectionName": "%install",
  "instruction": "添加 systemd 服务配置"
}
```

**返回类型**: `AIDraftResponse`

---

### 11. Provider 绑定推荐

推荐本地仓库与远端仓库的绑定关系。

**Endpoint**: `POST /api/v1/ai/provider/binding`

**请求体**:

```json
{
  "remoteRepos": ["org/repo1", "org/repo2"],
  "localRepos": ["repo/path1", "repo/path2"],
  "existingBindings": {
    "repo/path1": "org/repo1"
  }
}
```

**返回类型**: `AIAdviceResponse`

---

### 12. Patch 风险分析

分析 Patch 应用的兼容性和风险。

**Endpoint**: `POST /api/v1/ai/patch/analyze`

**请求体**:

```json
{
  "patchContent": "Patch 文件内容",
  "targetBranch": "main",
  "fileList": ["src/main.go", "src/config.go"]
}
```

**返回类型**: `AIDiagnosisResponse`

---

### 13. 审计日志摘要

分析审计日志，提炼异常行为摘要。

**Endpoint**: `POST /api/v1/ai/audit/summary`

**请求体**:

```json
{
  "events": ["日志条目1", "日志条目2"],
  "stats": {"push": 15, "pull": 8, "merge": 3},
  "anomalies": ["非常规时间操作", "大量权限变更"]
}
```

**返回类型**: `AIAdviceResponse`

---

### 14. 统计洞察

从开发统计数据中提炼有价值的洞察。

**Endpoint**: `POST /api/v1/ai/stats/insight`

**请求体**:

```json
{
  "stats": {"commits": 128, "authors": 8, "filesChanged": 45},
  "trends": {"weeklyCommits": [10, 15, 12, 18, 20]},
  "authorActivity": {"user1": 45, "user2": 32, "user3": 8}
}
```

**返回类型**: `AIAdviceResponse`

---

### 15. Webhook 失败分析

分析 Webhook 调用失败原因。

**Endpoint**: `POST /api/v1/ai/webhook/failure`

**请求体**:

```json
{
  "payload": "Webhook 请求载荷",
  "response": "远端响应内容",
  "statusCode": 500,
  "eventType": "push"
}
```

**返回类型**: `AIDiagnosisResponse`

---

## 通用数据结构

### AIRef - 引用对象

```typescript
{
  type: string           // 引用类型
  id: string             // 引用 ID
  label: string          // 显示标签
  filePath?: string      // 文件路径
  startLine?: number     // 起始行号
  endLine?: number       // 结束行号
  url?: string           // 外部链接
}
```

### AIAction - 行动对象

```typescript
{
  id: string             // 操作 ID
  label: string          // 按钮文字
  type: string           // 按钮类型
  description?: string   // 操作说明
}
```

### AIReviewFinding - 审查发现项

```typescript
{
  severity: string       // 严重程度: blocking/high/optional
  category: string       // 问题分类: security/bug/performance/style
  message: string        // 问题描述
  filePath?: string      // 文件路径
  startLine?: number     // 起始行号
  endLine?: number       // 结束行号
  suggestion?: string    // 修复建议
  confidence?: string    // 置信度
}
```

## 风险等级说明

| 等级 | 说明 | 建议 |
|------|------|------|
| `low` | 低风险 | 可自动应用或按推荐操作 |
| `medium` | 中等风险 | 需要用户确认后执行 |
| `high` | 高风险 | 建议手动审查后再执行 |
| `critical` | 严重风险 | 必须人工介入，禁止自动执行 |

## 页面-任务-API映射表

### 已实现（前端已接入）

| 页面 | 任务场景 | API 端点 | 响应类型 | 状态 |
|------|----------|----------|----------|------|
| **仓库详情页** (RepoDetailPage) | 仓库健康状态智能分析 | `POST /api/v1/ai/repo/summary` | `AIAdviceResponse` | ✅ 已实现 |
| **同步任务页** (SyncTaskPage) | 同步失败根因诊断 | `POST /api/v1/ai/sync/failure` | `AIDiagnosisResponse` | ✅ 已实现 |
| **评审详情页** (ReviewTaskDetailPage) | 审查结果智能总结分析 | `POST /api/v1/ai/review/summary` | `AIReviewResponse` | ✅ 已实现 |
| **评审详情页** (ReviewTaskDetailPage) | 变更风险智能评审（原始代码） | `POST /api/v1/ai/review` | `AIReviewResponse` | ✅ 已实现 |
| **通用** | AI 结果用户反馈 | `POST /api/v1/ai/feedback` | `{ success: boolean }` | ✅ 已实现 |

### 后端已实现（待前端接入）

| 页面 | 任务场景 | API 端点 | 响应类型 | 状态 |
|------|----------|----------|----------|------|
| **审计日志页** (AuditLogPage) | 审计日志异常模式识别 | `POST /api/v1/ai/audit/summary` | `AIAdviceResponse` | 🔧 后端就绪 |
| **统计面板页** (StatsDashboardPage) | 开发数据洞察分析 | `POST /api/v1/ai/stats/insight` | `AIAdviceResponse` | 🔧 后端就绪 |
| **补丁详情页** (PatchDetailPage) | 补丁变更智能诊断 | `POST /api/v1/ai/patch/analyze` | `AIDiagnosisResponse` | 🔧 后端就绪 |
| **Webhook管理页** (WebhookPage) | Webhook失败原因诊断 | `POST /api/v1/ai/webhook/failure` | `AIDiagnosisResponse` | 🔧 后端就绪 |
| **评审详情页** (ReviewTaskDetailPage) | 评审意见回复起草 | `POST /api/v1/ai/review/reply` | `AIDraftResponse` | 🔧 后端就绪 |
| **冲突解决页** | 冲突内容解释 | `POST /api/v1/ai/conflict/explain` | `AIAdviceResponse` | 🔧 后端就绪 |
| **冲突解决页** | 冲突解决方案生成 | `POST /api/v1/ai/conflict/resolve` | `AIDraftResponse` | 🔧 后端就绪 |
| **提交详情页** | Commit Message 智能生成 | `POST /api/v1/ai/commit/message` | `AIDraftResponse` | 🔧 后端就绪 |
| **分支管理页** | 分支命名规则推荐 | `POST /api/v1/ai/branch/rule` | `AIAdviceResponse` | 🔧 后端就绪 |
| **Provider绑定页** | Provider 绑定方案推荐 | `POST /api/v1/ai/provider/binding` | `AIAdviceResponse` | 🔧 后端就绪 |
| **Spec编辑器页** | Spec 章节智能重写 | `POST /api/v1/ai/spec/rewrite` | `AIDraftResponse` | 🔧 后端就绪 |
| **Spec编辑器页** | Spec 模板智能生成 | `POST /api/v1/ai/spec/template` | `AIDraftResponse` | 🔧 后端就绪 |

### 规划中

| 页面 | 任务场景 | API 端点 | 响应类型 | 状态 |
|------|----------|----------|----------|------|
| **分支对比页** (BranchComparePage) | 合并冲突预防分析 | `POST /api/v1/ai/branch/premerge-analysis` | `AIReviewResponse` | 📋 规划中 |

## 使用说明

1. 所有 AI API 均为异步调用，建议配合加载状态使用
2. 响应中的 `raw` 字段包含 AI 原始输出，用于调试
3. `applyContent` 字段的内容应该在用户确认后再应用到系统
4. 建议使用后端限流避免频繁调用
5. 所有 AI 响应包含 `invocationId`，可用于后续用户反馈提交
