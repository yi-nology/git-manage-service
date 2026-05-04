# API 字段命名规范

## 当前状态

项目中存在 snake_case 和 camelCase 混用情况：

### Backend DTO JSON Tag 统计

| 命名方式 | 示例 | 出现频率 |
|---|---|---|
| snake_case | `created_at`, `repo_id`, `is_default` | 较高 |
| camelCase | `ruleId`, `endLine`, `autoCommit` | 较低（主要在 spec 和 review 模块）|

### Frontend TypeScript 类型

前端类型文件中同样存在混用：
- `frontend/src/types/*.ts` 中大量使用 `created_at`, `updated_at`
- 部分新接口使用 `createdAt`

---

## 决策方案

### 方案 A: 统一使用 snake_case（推荐，兼容性最好）

**理由:**
- GORM 默认使用 snake_case，PO 字段直接复用多
- 与数据库字段命名一致，减少心智负担
- 现有大部分接口已使用 snake_case
- 改动量最小

**实施步骤:**
1. 新增规范文档，明确 snake_case 为默认命名
2. 新增接口严格遵循 snake_case
3. 逐步将 camelCase 字段兼容处理（后端 adapter 层同时输出两种）

### 方案 B: 统一使用 camelCase

**理由:**
- JavaScript/TypeScript 社区标准
- RESTful API 常见做法

**实施步骤:**
1. 全局搜索替换所有 JSON tag
2. 同步修改前端类型定义
3. 后端增加兼容层（snake_case → camelCase）
4. 改动量较大，需配合前端全面回归测试

---

## 规范细则

### 字段命名

```go
// ✅ 推荐
type Repo struct {
    ID        uint      `json:"id"`
    RepoID    uint      `json:"repo_id"`       // 不是 repoId
    Name      string    `json:"name"`
    IsDefault bool      `json:"is_default"`    // 不是 isDefault
    CreatedAt time.Time `json:"created_at"`    // 不是 createdAt
    UpdatedAt time.Time `json:"updated_at"`
}
```

### 特殊情况处理

| 术语 | 命名 |
|---|---|
| ID 类 | `repo_id`, `commit_id`, `task_id` |
| 布尔前缀 | `is_xxx`, `has_xxx`, `can_xxx` |
| 时间戳 | `created_at`, `updated_at`, `deleted_at` |
| 计数后缀 | `file_count`, `commit_count` |

---

## 检查工具

可使用 `golangci-lint` 的 `tagliatelle` linter 自动检查 JSON tag 命名：

```yaml
linters-settings:
  tagliatelle:
    case:
      rules:
        json: snake
```
