# AI 剩余问题清单

这份清单基于当前代码状态整理，重点列出“还能继续改，但不一定会直接编译失败”的问题。  
优先级按 `P1 > P2 > P3` 排。

## P1

### 1. 页面虽然接上了 AI 面板，但用户输入内容仍然没有真正参与分析

当前 3 个页面的 handler 都把面板输入参数写成了 `_message`，实际请求时没有使用用户输入内容：

- [frontend/src/views/repo/RepoDetailPage.vue](/opt/project/wechat_project/git-manage-service/frontend/src/views/repo/RepoDetailPage.vue:440)
- [frontend/src/views/sync/SyncTaskPage.vue](/opt/project/wechat_project/git-manage-service/frontend/src/views/sync/SyncTaskPage.vue:491)
- [frontend/src/views/review/ReviewTaskDetailPage.vue](/opt/project/wechat_project/git-manage-service/frontend/src/views/review/ReviewTaskDetailPage.vue:219)

现状：

- 用户在面板里输入什么，结果基本一样
- 快捷操作按钮只是触发一个固定流程
- 这更像“AI 工具入口”，不是“可追问的 AI 助手”

建议：

1. 把用户输入并入请求上下文
2. 区分“固定动作”与“自由提问”两种模式
3. 对 `repo summary / sync diagnosis / review analysis` 至少支持一句自然语言约束，例如：
   - “只看风险最高的问题”
   - “重点看同步失败原因，不要给泛化建议”
   - “从 reviewer 视角总结”

---

### 2. Review 页面调用的是 code review API，但传入的并不是代码 diff

位置：

- [frontend/src/views/review/ReviewTaskDetailPage.vue](/opt/project/wechat_project/git-manage-service/frontend/src/views/review/ReviewTaskDetailPage.vue:222)

当前传入的 `diff` 实际内容是：

- 已有 findings 的标题和 message 拼接文本

不是：

- 真正的 merge request diff
- 也不是文件 patch

影响：

- 后端 `aiApi.codeReview()` 的语义被破坏
- 模型看到的是“问题摘要的二次摘要”，不是代码审查
- 返回结果参考价值有限

建议：

1. 这里不要复用 `codeReview` 接口做“复盘”
2. 单独做一个 `review.summary` / `review.explain` 类任务更合适
3. 如果坚持用 `codeReview`，就必须提供真实 diff

---

### 3. Sync 诊断仍然只分析“最近一次失败”，没有和用户当前关注对象对齐

位置：

- [frontend/src/views/sync/SyncTaskPage.vue](/opt/project/wechat_project/git-manage-service/frontend/src/views/sync/SyncTaskPage.vue:492)

当前逻辑：

1. 拉整个 repo 的 sync history
2. 找失败记录
3. 排序
4. 只拿最新一条失败 run 做分析

问题：

- 用户可能想看某个特定 task
- 也可能是刚刚点开的某条历史记录
- 当前实现没有“选中上下文”

建议：

1. 优先分析当前用户选中的 task / 当前打开的历史项
2. 没有选中对象时，再回退到最近一次失败
3. 在面板中明确显示“本次分析对象是谁”

---

## P2

### 4. Repo summary 仍然缺少真实 commit 维度，`commitCount` 还是占位值

位置：

- [frontend/src/views/repo/RepoDetailPage.vue](/opt/project/wechat_project/git-manage-service/frontend/src/views/repo/RepoDetailPage.vue:471)

现状：

- `commitCount` 仍然固定传 `0`
- 这会让 AI 对仓库规模和活跃度的判断偏空

虽然这次已经补了：

- branch
- ahead / behind
- staged / unstaged / untracked / conflicted
- recent sync failure

但仓库总览仍然缺少：

- 实际 commit 总量
- 最近活跃度
- 最近变更密度

建议：

1. 如果后端已有 stats 接口，直接取真实 commit 量
2. 没有的话，把这个字段先从 prompt 里降权，避免误导模型

---

### 5. AI 面板的反馈按钮只是本地改状态，没有写回后端审计

位置：

- [frontend/src/components/ai/AIPanel.vue](/opt/project/wechat_project/git-manage-service/frontend/src/components/ai/AIPanel.vue:71)
- [frontend/src/components/ai/AIPanel.vue](/opt/project/wechat_project/git-manage-service/frontend/src/components/ai/AIPanel.vue:240)
- [biz/service/ai/audit.go](/opt/project/wechat_project/git-manage-service/biz/service/ai/audit.go:65)

现状：

- 前端点击“有用 / 没用”只是在内存里给 `msg.feedback` 赋值
- 后端其实已经有 `RecordUserFeedback()` 能力
- 但没有 API，也没有把 invocation id 传回前端

影响：

- 反馈数据完全丢失
- 审计表里的 `user_feedback` 字段没有形成闭环

建议：

1. 后端返回 invocation id
2. 增加 feedback API
3. 前端把点赞/点踩真正持久化

---

### 6. `AIApplyDialog` 组件目前仍未接入任何页面

现状：

- 组件文件已经存在
- 但当前代码里没有实际使用点

这说明“Draft / Apply”这条交互链还没真正落地。

建议优先接入：

1. conflict resolve
2. spec rewrite / template
3. review reply draft

原因：

- 这三类都有天然的“生成草案 -> 用户确认 -> 应用”的形态

---

### 7. `AIReviewResponse` 还保留了 `high` 和 `highRisk` 双字段兼容

位置：

- [frontend/src/types/ai.ts](/opt/project/wechat_project/git-manage-service/frontend/src/types/ai.ts:61)

现状：

- `high`
- `highRisk`

同时存在

这在当前阶段可以接受，但不应长期保留。

建议：

1. 等前后端全部稳定后统一删除 `high`
2. 保留单一标准字段 `highRisk`

---

## P3

### 8. AIPanel 的 `visible` props 目前没有实际参与行为控制

位置：

- [frontend/src/components/ai/AIPanel.vue](/opt/project/wechat_project/git-manage-service/frontend/src/components/ai/AIPanel.vue:128)

现状：

- 页面都在传 `:visible="showAIPanel"`
- 但组件内部几乎没有使用这个值

这不是 bug，但说明组件 contract 还不够干净。

建议二选一：

1. 要么删掉 `visible` props
2. 要么让它真正参与内部状态或动画控制

---

### 9. `docs/AI_API.md` 现在偏接口说明，缺少页面接入说明和任务语义映射

位置：

- [docs/AI_API.md](/opt/project/wechat_project/git-manage-service/docs/AI_API.md)

当前文档已经能说明接口，但还缺：

- 哪个页面用哪个接口
- 哪些接口属于 `Explain / Suggest / Draft / Diagnosis`
- 哪些接口会产生可应用草案
- 哪些接口当前只支持固定动作，不支持自由提问

建议：

1. 补一张“页面 -> AI 任务 -> API”映射表
2. 补“当前已落地 / 计划中 / 未接入”状态

---

## 建议处理顺序

### 第一轮

1. 让用户输入真正参与 3 个页面的 AI 请求
2. 把 Review 页从“伪 diff”改成真实 review summary 任务
3. 让 Sync 诊断支持“当前对象优先”

### 第二轮

1. 接入 `AIApplyDialog`
2. 打通反馈持久化
3. 清理 `high` / `highRisk` 兼容字段

### 第三轮

1. 清理 `AIPanel` contract
2. 补文档映射

---

## 当前结论

现在这套 AI 接入已经从“接口未接通”进展到“基础闭环可用”，但还没有到“真正好用”的阶段。  
下一步最关键的不是继续加更多按钮，而是把：

- 用户输入
- 当前业务对象
- 可应用草案
- 用户反馈

这四条链补完整。
