# 代码重构拆分计划

## git_maintenance.go 拆分方案

当前文件: 718 行

### 拆分后结构

| 文件 | 功能 | 行估算 |
|---|---|---|
| maintenance_snapshot.go | TakeSnapshot 快照统计 | ~80 |
| maintenance_health.go | AnalyzeHealth 健康检查 + matchExclude | ~80 |
| maintenance_large_files.go | FindLargeFiles, FindStashLargeObjects, FindReflogLargeObjects | ~180 |
| maintenance_gitdir.go | ScanGitDirBreakdown, calcLooseObjSize | ~60 |
| maintenance_slim.go | SlimHistory, GarbageCollect, appendToGitignore | ~150 |
| maintenance_record.go | CreateMaintenanceRecord | ~30 |
| maintenance_utils.go | dirSize, formatSize 等工具函数 | ~40 |

## SpecEditor.vue 拆分方案

当前文件: 571 行

### 拆分后结构

| 文件 | 功能 |
|---|---|
| SpecEditor.vue | 主组件容器，状态管理 |
| SpecFileTree.vue | 文件树组件 |
| SpecEditorPane.vue | Monaco 编辑器包装 |
| SpecLintPanel.vue | Lint 结果面板 |
| SpecAIPanel.vue | AI 辅助功能面板 |
| SpecToolbar.vue | 顶部工具栏 |
| useSpecEditor.ts | Composable 抽离状态逻辑 |

## 实施优先级

### Phase 1 (已完成)
- ✅ maintenance_ai.go (AI 分析已拆分)

### Phase 2 (建议)
- maintenance_snapshot.go + maintenance_health.go
- maintenance_large_files.go

### Phase 3 (后续)
- maintenance_gitdir.go
- maintenance_slim.go
- SpecEditor.vue 组件拆分
