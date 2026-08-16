# AI 整体切入点设计

## 1. 目标

这个项目不适合把 AI 做成一个“到处放聊天框”的产品。更合理的方向是：

1. 让 AI 参与高认知成本、低确定性的环节
2. 让确定性流程继续由规则、Git、数据库事务来做
3. AI 产出默认是“建议 / 草案 / 风险判断”，涉及写仓库、改配置、发评论、推远端时必须有人确认

一句话概括：**AI 负责理解、归纳、建议、起草；系统负责执行、校验、审计、回滚。**

---

## 2. 先定边界：什么该给 AI，什么不该

### 2.1 适合 AI 的能力

- 自然语言问答
- 复杂上下文总结
- 风险识别与解释
- 变更建议 / 修复草案
- 多来源信息归并
- 模糊匹配与相似性判断
- Review 评论草拟
- 面向用户的操作建议

### 2.2 不适合直接交给 AI 的能力

- Git 基础操作执行本身：`checkout / reset / push / merge / stash`
- 数据库写入逻辑和状态流转
- 权限判断
- 路由选择
- 精确统计计算
- CI 通过/失败判断
- Webhook 幂等与重试

### 2.3 AI 输出的四种级别

1. **Explain**
   只解释，不改数据
2. **Suggest**
   给建议，不落库
3. **Draft**
   生成可应用草案，用户确认后落地
4. **Act**
   允许自动执行，但必须只用于低风险、可回滚、可审计动作

当前项目建议默认只做到 `Explain / Suggest / Draft`，`Act` 只在少数场景启用。

---

## 3. 整体架构建议

## 3.1 统一 AI 能力层

当前项目已经有这些 AI 入口：

- `biz/service/ai/runner.go`
- `biz/service/spec/spec_service_ai.go`
- `biz/service/lint/ai_lint_service.go`
- `biz/service/git/author_ai_service.go`
- `biz/service/git/workspace_ai_service.go`
- `biz/service/git/maintenance_ai.go`
- `biz/service/codereview/review_service_publish.go`

这说明底座已经有了，但还偏“按功能点散接”。下一步应统一成三层：

1. **Task 层**
   定义 AI 任务类型、输入输出契约、提示词版本、风险等级
2. **Context Builder 层**
   每个域自己负责整理上下文，不让 handler 直接拼 prompt
3. **Execution 层**
   统一走 `ai.Runner`，负责 provider 选择、脱敏、限流、审计、超时、降级

建议扩展 `biz/service/ai/task.go`，把任务定义补齐：

- `repo.summary`
- `repo.register_advice`
- `file.explain`
- `commit.message`
- `commit.summary`
- `branch.risk`
- `merge.plan`
- `sync.plan`
- `sync.failure_analysis`
- `review.summary`
- `review.reply_draft`
- `webhook.failure_analysis`
- `audit.summary`
- `stats.insight`
- `provider.binding_recommendation`

## 3.2 统一前端交互形态

不要每个页面都自己长一套 AI UI。建议前端只保留三种标准入口：

1. **侧边 AI Panel**
   适合 Repo、Spec、Review、Remote Repo 详情页
2. **行内 Suggestion Card**
   适合冲突解决、作者修复、异常诊断、同步失败分析
3. **Apply Diff Dialog**
   适合任何会生成修改内容的场景，必须带 diff 和确认

所有 AI 能力尽量复用：

- 同一套消息结构
- 同一套“引用上下文”展示
- 同一套“应用变更 / 拒绝变更”流程
- 同一套审计记录

## 3.3 必须补的横切能力

### 可观测性

- 调了哪个 provider
- 用了哪个 task
- prompt version 是多少
- 输入来源有哪些
- 响应耗时
- 是否被用户采纳

### 审计

AI 相关落表建议至少记录：

- task type
- repo key / provider id / review task id 等业务对象
- input snapshot 摘要
- output 摘要
- apply / reject 结果
- operator

### 安全

- 仓库内容脱敏
- Token / 密钥 / URL 凭据过滤
- 大文件截断
- diff 上下文裁剪
- provider allowlist

---

## 4. 分功能域切入点

下面按“价值优先级 + 可实现性”来排。

## 4.1 Repo 仓库管理

### 现有能力

- 已有工作区 AI 冲突解决：`biz/service/git/workspace_ai_service.go`
- 已有维护分析：`biz/service/git/maintenance_ai.go`

### 建议新增

#### A. 仓库接入建议

适用页面：

- `RepoRegisterPage.vue`
- `RepoClonePage.vue`

AI 做什么：

- 根据远端地址、目录结构、已有分支、构建文件，判断仓库类型
- 给出初始化建议：默认分支、建议同步策略、建议排除目录、建议绑定 provider

输入：

- repo path
- remote URL
- 根目录文件清单
- 检测到的语言/构建工具

输出：

- 仓库类型判断
- 初始化建议卡片
- 风险提示

价值：

- 降低新仓接入成本

#### B. 仓库总览摘要

适用页面：

- `RepoDetailPage.vue`

AI 做什么：

- 总结仓库状态：分支脏状态、未推送提交、同步风险、最近异常、Spec/Review/Sync 告警

输出：

- “当前仓库建议先处理什么”
- 三到五条优先项

价值：

- 让首页从“信息堆叠”变成“行动入口”

#### C. 仓库操作前风险说明

适用动作：

- force push
- history rewrite
- branch delete
- bulk author rewrite

AI 做什么：

- 把底层 Git 风险翻译成人能快速判断的提示
- 不替代规则判断，只负责解释影响面

---

## 4.2 File / Commit / Diff

### 建议新增

#### A. 文件解释

适用页面：

- `FileExplorer.vue`
- `FileDiffViewer.vue`

AI 做什么：

- 解释当前文件作用
- 解释 diff 改了什么
- 识别高风险变更点

适合按钮：

- “解释文件”
- “解释本次修改”
- “识别风险”

#### B. Commit message 草拟

当前 `TaskCommitMessage` 已有枚举，但未形成完整入口。

建议接入：

- 提交前根据 staged diff 自动生成 commit message
- 支持 `simple / conventional / detailed` 三种风格

注意：

- 只能生成草稿，不自动提交

#### C. 提交摘要 / 变更说明

场景：

- 提交详情页
- 发 CR/MR 前
- Patch 导出前

AI 输出：

- 简版摘要
- 影响面
- 测试建议
- 发布风险

---

## 4.3 Branch / Merge / Stash / Patch / Submodule

### A. 分支策略建议

适用：

- `BranchActionsPage.vue`
- `BranchDetailPage.vue`
- branch rules 配置页

AI 做什么：

- 基于仓库现状给出分支命名、合流路径、保护策略建议
- 为远端仓库生成推荐 branch rule 草案

### B. 合并前分析

AI 输入：

- ours/theirs/base diff
- 目标分支状态
- 冲突文件类型

AI 输出：

- 是否建议 merge / rebase
- 冲突热点
- 推荐处理顺序

### C. 冲突解决增强

当前已有 `AIResolveConflict`，但还可以补：

- 冲突原因解释
- 变更来源对比
- 应用前 diff 可视化
- “只保留 ours / theirs / 混合方案”三种推荐模板

### D. Patch 风险分析

适用：

- `PatchManager.vue`

AI 做什么：

- 分析 patch 是否适合跨分支应用
- 识别依赖上下文缺失
- 给出失败原因预估

### E. Submodule 异常解释

AI 做什么：

- 解释 submodule 指针变化意味着什么
- 提醒是否需要同步主仓说明文档或版本记录

---

## 4.4 Spec 子系统

### 现有能力

- `spec_service_ai.go`
- `ai_lint_service.go`
- 前端 `useAIChat.ts`

这是当前最成熟的一条 AI 线。

### 建议继续增强

#### A. 从“问答”升级成“工作流”

当前模式已经有：

- `chat`
- `complete`
- `generate`
- `agent`

建议再补两个固定动作：

- `rewrite_section`
- `explain_errors`

#### B. 引入“规则 + AI 串联”

不要让 AI 单独做全部 lint。推荐流程：

1. 规则 lint 先跑
2. 把规则结果作为 AI 上下文
3. AI 只补充规则覆盖不到的问题
4. AIFix 只能针对用户选中的 issue 生成修复草案

#### C. 增加结构化引用

AI 输出除了文本，建议补：

- 关联 section
- 关联行号范围
- 推荐修改片段
- 风险等级

#### D. 增加模板生成能力

场景：

- 新建 spec
- 从源码目录推断 spec 基础骨架
- 从已有 spec 生成同类包变体

这条价值很高，优先级可排在 Spec 的下一轮。

---

## 4.5 Author 身份治理

### 现有能力

- `author_ai_service.go`
- 智能别名推荐
- 身份合并建议
- 历史重写风险评估

这条方向是对的，建议继续强化成“治理闭环”。

### 建议新增

#### A. 扫描结果自动归因

AI 不只是说“这个人可能是同一个人”，而是给出：

- 建议合并到哪个 canonical identity
- 理由
- 置信度
- 影响提交数

#### B. 批量修复方案预演

在真正执行 filter-branch 前，AI 先生成：

- 影响范围摘要
- 风险点
- 推荐分批方案
- 是否应先备份 / 打 tag

#### C. 团队身份规范建议

根据现有仓库提交历史，反推出：

- 推荐作者命名规范
- 邮箱域治理建议
- 机器人账号隔离建议

---

## 4.6 Sync / Remote Repo / Provider Binding

### A. 同步计划生成

适用：

- `SyncTaskPage.vue`
- `QuickSyncPanel.vue`

AI 做什么：

- 根据本地分支、远端默认分支、落后/超前状态，生成同步建议
- 给出“推荐走 pull --rebase / merge / push / 先解决冲突”的方案

### B. 同步失败诊断

这是高价值场景。

输入：

- sync run 日志
- Git stderr
- 当前 branch 状态
- remote tracking 信息

输出：

- 根因总结
- 建议下一步操作
- 是否可自动修复

### C. 远端仓与本地仓绑定建议

适用：

- `RemoteReposPage.vue`
- `RemoteRepoDetailPage.vue`

AI 做什么：

- 根据 repo name / owner / remote URL / default branch / 最近提交特征，推荐可能的绑定关系

### D. Provider 配置健康检查

AI 不负责校验 token 是否有效，这个必须程序判定。

AI 负责：

- 对失败响应做诊断说明
- 建议权限范围
- 提示缺少的 webhook / branch rule / review 配置

---

## 4.7 Review / CR / Code Review

### 现有能力

- `biz/service/codereview/*`
- 规则 review + LLM review

这是另一个适合做核心 AI 能力的域。

### 建议增强方向

#### A. Review 结果分层

现在更像“发现列表”。建议拆成：

1. 阻断问题
2. 高风险建议
3. 可选改进
4. 总结摘要

#### B. Review 配置化

按 repo / 语言 / provider / 变更类型走不同 prompt 策略：

- Go 后端
- Vue 前端
- Spec 文件
- SQL migration
- 配置类变更

#### C. CR 回复草拟

适用：

- `ReviewTaskDetailPage.vue`
- `CRManagementPage.vue`

AI 做什么：

- 帮开发者草拟 reviewer 回复
- 帮维护者草拟总结评论
- 根据 findings 自动生成 follow-up checklist

#### D. Merge 前汇总

AI 输出：

- 本次 CR 改了什么
- 风险集中在哪
- 是否建议合并
- 还缺什么验证

#### E. 规则与 LLM 去重

这个很关键。

规则能发现的不要再让 LLM 重复报。建议：

- 先跑 rule engine
- 再把已有 findings 传给 LLM
- LLM 只输出新增语义问题

---

## 4.8 Stats / Audit / Notification / Webhook

### A. Stats 洞察

适用：

- `RepoStatsTab.vue`
- `RepoLineStatsTab.vue`
- 首页 Dashboard

AI 做什么：

- 从提交趋势、作者活跃度、语言变化里提炼“可读结论”
- 不直接参与统计计算

输出示例：

- 本周提交集中在哪些目录
- 哪些模块 churn 高
- 哪些作者提交异常集中

### B. Audit 摘要与异常聚类

适用：

- `AuditLogPage.vue`

AI 做什么：

- 把大量 audit log 总结成操作主题
- 聚类异常操作
- 给出“近期值得关注的行为”

### C. Notification 文案生成

适用：

- 事件模板配置
- 失败通知
- Review 通知

AI 做什么：

- 根据事件上下文生成更可读的通知摘要
- 支持不同渠道文风：飞书 / 企业微信 / 邮件

### D. Webhook 失败解释

适用：

- `WebhookEventsPage.vue`

AI 做什么：

- 解释 webhook payload / 响应失败原因
- 建议重试还是人工处理

---

## 4.9 Settings / LLM / MCP

### A. LLM Provider 选型建议

适用：

- `LLMSettingsPage.vue`

AI 不自己给自己配模型，而是系统基于任务类型给出推荐：

- Spec lint 偏便宜模型
- Code review 偏强模型
- 长 diff 总结偏大上下文模型

### B. Prompt 模板管理

建议把现在散落在 service 里的 prompt，逐步收敛成：

- 系统默认模板
- repo 级覆盖
- 任务级变量

### C. MCP 辅助说明

适用：

- `McpPage.vue`

AI 做什么：

- 用自然语言解释每个 MCP 工具能做什么
- 根据用户意图推荐对应工具

---

## 5. 统一交互模型

每个功能域都不要重新发明一套 AI API。建议统一成下面的交互语义。

## 5.1 查询类

输入：

- `context`
- `question`

输出：

- `answer`
- `references`
- `confidence`

## 5.2 建议类

输入：

- `context`
- `goal`
- `constraints`

输出：

- `summary`
- `suggestions[]`
- `riskLevel`

## 5.3 草案类

输入：

- `context`
- `instruction`

输出：

- `draft`
- `changeSummary`
- `riskLevel`
- `applyContent` 或结构化 patch

## 5.4 诊断类

输入：

- `logs`
- `status`
- `recentActions`

输出：

- `rootCause`
- `evidence`
- `recommendedActions`

---

## 6. 前端落地建议

## 6.1 页面级 AI 入口优先级

第一批值得放 AI 的页面：

1. `RepoDetailPage.vue`
2. `SpecEditor.vue`
3. `ReviewTaskDetailPage.vue`
4. `SyncTaskPage.vue`
5. `RemoteRepoDetailPage.vue`
6. `AuditLogPage.vue`

## 6.2 一个统一 AI 面板就够了

建议抽一个通用组件，例如：

- `frontend/src/components/ai/AIPanel.vue`
- `frontend/src/components/ai/AIResultCard.vue`
- `frontend/src/components/ai/AIApplyDialog.vue`

避免继续把 AI 逻辑散在：

- `useAIChat.ts`
- 各种页面私有按钮
- 零散的 `request.post('/xxx/ai-...')`

## 6.3 前端展示必须有“来源感”

AI 结论不能像凭空冒出来。每个结果都应展示：

- 使用了哪些输入
- 针对哪个 repo / branch / file / sync run
- 是否使用历史消息
- 是否可直接应用

---

## 7. 后端实施建议

## 7.1 不建议继续“一个功能一个随意 JSON”

建议给 AI 输出统一 DTO，至少抽这几个：

- `AIAdviceResponse`
- `AIDraftResponse`
- `AIDiagnosisResponse`
- `AIReviewResponse`

统一字段：

- `summary`
- `confidence`
- `riskLevel`
- `references`
- `actions`
- `applyContent`
- `raw`

## 7.2 每个域补一个 Context Builder

例如：

- `biz/service/ai/context_repo.go`
- `biz/service/ai/context_review.go`
- `biz/service/ai/context_sync.go`
- `biz/service/ai/context_spec.go`

这样 prompt 不会继续分散在 handler/service 各处。

## 7.3 加任务级策略

不同任务需要不同策略：

- 最大输入长度
- 超时时间
- 默认 provider
- 是否允许自动应用
- 是否要求结构化 JSON 返回

这部分应该挂在 task 配置上，而不是散在每个 service 里。

---

## 8. 推荐实施顺序

## Phase 1：先把现有 AI 收敛成平台能力

目标：

- 统一 task 定义
- 统一返回 DTO
- 统一审计与脱敏
- 统一前端 AI 面板

优先做：

1. `spec`
2. `review`
3. `author`
4. `workspace conflict`

## Phase 2：做高价值诊断与摘要

优先做：

1. sync failure analysis
2. repo summary
3. audit summary
4. webhook failure analysis
5. stats insight

## Phase 3：做草案生成与可应用动作

优先做：

1. commit message draft
2. CR reply draft
3. branch rule recommendation
4. repo binding recommendation
5. spec template generation

## Phase 4：谨慎引入半自动执行

只建议开放给低风险能力：

- 应用 spec 修改草案
- 应用冲突解决草案
- 应用通知模板草案

不建议直接自动执行：

- push
- merge
- rewrite history
- delete branch
- bulk remote operation

---

## 9. 最终建议：这项目最值得做成 AI 核心能力的 6 条线

如果资源有限，我建议只先把下面 6 条线做深：

1. **Spec 智能编辑**
2. **Code Review 智能审查**
3. **Sync 失败诊断**
4. **Author 身份治理**
5. **冲突解决与合并建议**
6. **仓库总览摘要与下一步建议**

原因很简单：

- 它们都天然是高认知成本场景
- 已经有部分代码基础
- AI 能带来明确体验提升
- 风险相对可控

相反，像“普通 CRUD 页加聊天框”价值很低，先不要做。

---

## 10. 一句话落地策略

先把 AI 从“零散功能按钮”升级成“统一能力平台”，再在 `Spec / Review / Sync / Author / Conflict / Repo Summary` 六条主链上做深做透。
