# AI 改进清单

基于当前代码审查结果，下面这份清单按优先级整理，优先处理会阻断功能可用性的项。

## P1 必修

### 1. 补齐 AI 路由注册

问题：

- `biz/handler/ai/ai_handler.go` 已经实现了一组 `/api/v1/ai/*` handler
- 但 `biz/router/register.go` 没有注册对应路由
- 前端 `frontend/src/api/modules/ai.ts` 当前会统一命中 `404`

需要改：

1. 新增 `biz/router/ai/`
2. 在该目录中注册全部 AI 路由
3. 在 `GeneratedRegister()` 中引入并调用 `ai.Register(h)`

建议覆盖的路由：

- `/api/v1/ai/sync/failure`
- `/api/v1/ai/repo/summary`
- `/api/v1/ai/commit/message`
- `/api/v1/ai/review`
- `/api/v1/ai/review/reply`
- `/api/v1/ai/conflict/resolve`
- `/api/v1/ai/conflict/explain`
- `/api/v1/ai/branch/rule`
- `/api/v1/ai/spec/template`
- `/api/v1/ai/spec/rewrite`
- `/api/v1/ai/provider/binding`
- `/api/v1/ai/patch/analyze`
- `/api/v1/ai/audit/summary`
- `/api/v1/ai/stats/insight`
- `/api/v1/ai/webhook/failure`

验收：

- 以上接口不再返回 `404`
- 至少补一轮最小接口烟测

---

### 2. 统一 CodeReview 高风险字段命名

问题：

- prompt 要求模型返回 `high`
- 后端 DTO 定义的是 `highRisk`
- 前端类型写的是 `high`

结果：

- 模型如果按 prompt 返回 `high`，后端无法正确解析
- 即使后端改成 `highRisk`，前端也会对不上
- “高风险但非 blocking” 的结果会被静默丢失

建议统一：

- 统一成 `highRisk`

需要改：

- `biz/model/api/ai.go`
- `biz/service/ai/service.go`
- `frontend/src/types/ai.ts`
- 全局搜索所有 `high` / `highRisk` 相关引用并收口

补充建议：

- 短期可做兼容解析，同时接受 `high` 和 `highRisk`
- 但最终对外契约只能保留一个字段

验收：

- AI review 返回高风险项时，后端能解析
- 前端能正确展示

---

## P2 应修

### 3. 修复 `AIApplyDialog` 的 props 双向绑定问题

问题：

- `visible` 直接通过 `v-model` 绑定到 props
- `commitMessage` 也直接通过 `v-model` 绑定到 props

风险：

- Vue 会产生 prop mutation warning
- 父组件状态不同步
- 真正接入页面后，弹窗开关和提交信息编辑都可能失效

建议改法：

1. 引入本地状态：
   - `localVisible`
   - `localCommitMessage`
2. 通过 `watch(props.visible)` 同步到本地
3. 通过 `watch(props.commitMessage)` 同步到本地
4. 用户操作后通过事件回传：
   - `update:visible`
   - 如有必要，增加 `update:commitMessage`

验收：

- 不再出现 prop mutation warning
- 弹窗显示/关闭与父组件同步正常
- 提交信息输入可编辑且能正确回传

---

### 4. 修复 `AIPanel` 的可选值与类型问题

问题：

- `quickActions` 是可选值，但模板里直接访问 `quickActions.length`
- 未传值时会触发 `undefined.length`
- `setMessagesRef` 使用了 `ComponentPublicInstance`，但脚本没有导入类型

风险：

- 存在运行时错误
- `vue-tsc` 会报错
- 通用 AI 面板无法稳定复用

建议改法：

1. 给 `quickActions` 提供默认值 `[]`
2. 补充 `ComponentPublicInstance` 类型导入
3. 顺手检查 `marked.parse()` 的返回类型与 `v-html` 使用是否一致
4. 统一组件对外暴露的方法和接入方式

验收：

- `AIPanel` 在未传 `quickActions` 时可正常运行
- `make typecheck-frontend` 通过

---

## P2 建议顺手做

### 5. 把通用 AI 组件真正接入业务页面

当前问题：

- 通用组件文件已经存在
- 但还没有形成稳定的页面闭环

建议优先接入页面：

1. `frontend/src/views/sync/SyncTaskPage.vue`
2. `frontend/src/views/repo/RepoDetailPage.vue`
3. `frontend/src/views/review/ReviewTaskDetailPage.vue`

至少要做到的闭环：

1. 发起请求
2. 展示 AI 返回结果
3. 展示引用信息
4. 对草案支持“确认应用 / 拒绝”

---

### 6. 为新增 AI API 补最小可用测试

后端建议补：

- handler bind test
- service JSON parse test
- 至少一条路由可达性测试

前端建议补：

- 通用 AI 组件最小挂载测试
- `vue-tsc` 检查纳入验收

目标：

- 不让这套 AI 平台层停留在“文件已创建，但无法保证可用”

---

## P3 优化项

### 7. 为 CodeReview 返回结果增加兼容层

建议：

- 兼容 `high` 和 `highRisk`
- 只作为短期过渡方案

目的：

- 避免旧 prompt / 旧前端 / 旧缓存数据导致直接失效

注意：

- 兼容层不能长期保留
- 收口后只保留一个标准字段

---

### 8. 补 AI API 说明文档

建议在 `docs/` 中增加一份 API 说明，内容包括：

- 路径
- 入参
- 返回 DTO
- 任务类型
- 是否属于 `Explain / Suggest / Draft / Diagnosis`

价值：

- 降低后续前后端联调成本
- 避免 prompt、DTO、页面接入继续各写各的

---

## 推荐执行顺序

### 第一轮

1. 补 AI 路由注册
2. 统一 `highRisk` 字段命名
3. 修 `AIPanel`
4. 修 `AIApplyDialog`

### 第二轮

1. 将通用 AI 组件接入 `SyncTaskPage`
2. 接入 `RepoDetailPage`
3. 接入 `ReviewTaskDetailPage`

### 第三轮

1. 补测试
2. 补 AI API 文档
3. 清理兼容层

---

## 验收清单

完成后至少执行：

1. `make typecheck-frontend`
2. `make build`
3. `make test`

并确认以下行为：

- `/api/v1/ai/*` 不再 `404`
- CodeReview 的高风险项不会丢失
- `AIPanel` 无类型错误、无运行时报错
- `AIApplyDialog` 无 prop mutation warning
- 至少一个页面已经完整跑通 AI 请求到结果展示的闭环
