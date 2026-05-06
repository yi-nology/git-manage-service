# AI Review 后续收口文档

这份文档整理了当前 AI 接入完成后，仍然存在的 4 个主要收口问题。  
这些问题不会马上导致构建失败，但会持续影响：

- 文档可信度
- 接口语义一致性
- 前后端契约清晰度
- 用户反馈链路可靠性

建议按本文档顺序继续处理。

---

## 1. AI API 文档仍然和实际实现不一致

### 问题描述

当前 [AI_API.md](/opt/project/wechat_project/git-manage-service/docs/AI_API.md) 中“页面-任务-API 映射表”仍然包含若干当前并未真实注册或接入的接口，例如：

- `POST /api/v1/ai/review/summary`
- `POST /api/v1/ai/review/reply-draft`
- `POST /api/v1/ai/branch/premerge-analysis`

同时，文档中还列出了部分当前并未真实接入的页面或场景，例如：

- `PatchDetailPage`
- `StatsDashboardPage`

而实际代码里当前可确认存在的实现仍然是类似：

- `/api/v1/ai/review`
- `/api/v1/ai/review/reply`
- `/api/v1/ai/branch/rule`

### 影响

1. 联调时会直接被错误文档带偏
2. 新页面接入会建立在错误 endpoint 假设上
3. 维护者无法快速判断“哪些已经上线，哪些只是规划”

### 根因

当前文档把两类内容写在了一起：

1. **代码里已存在的真实接口**
2. **设计上规划但尚未落地的能力**

文档层没有做状态分层，所以产生了漂移。

### 建议改法

建议把映射表拆成两层：

#### A. 已实现

只保留：

- 当前已注册路由
- 当前前端已调用
- 当前后端已有 handler/service 的 AI API

#### B. 规划中

单独列出：

- 未来计划能力
- 尚未注册的 endpoint
- 尚未接入的页面

建议在文档表格中增加一列：

- `状态`
  - `已实现`
  - `部分实现`
  - `规划中`

### 验收标准

- `AI_API.md` 中所有“已实现”接口都能在路由注册中找到
- 所有“已实现”页面都能在 `frontend/src/views` 中找到实际调用
- 不再出现“文档存在，但代码没有”的 endpoint 映射

---

## 2. Review 页面仍然在用 codeReview 接口做“问题复盘”

### 问题描述

当前 [ReviewTaskDetailPage.vue](/opt/project/wechat_project/git-manage-service/frontend/src/views/review/ReviewTaskDetailPage.vue) 中，AI 面板已经支持：

- 用户输入问题
- 用户输入进入请求
- 返回结果写回面板

但当前调用的接口仍然是：

- `POST /api/v1/ai/review`

而传给这个接口的 `diff` 实际不是原始代码 diff，而是：

- 现有 findings 的标题与 message 拼接出的摘要文本

也就是说，当前页面实际做的是：

- “对审查结果做二次分析和复盘”

而不是：

- “对代码 diff 进行审查”

### 影响

1. 页面语义和接口语义不一致
2. 模型拿不到真实代码上下文
3. 输出质量会受限，本质上变成“总结 findings”
4. 后续维护时很难区分：
   - code review
   - review summary
   - review explain

### 根因

当前实现复用了已有 `/api/v1/ai/review` 能力，但没有为“评审复盘 / 审查结果解释”单独建模。

### 建议改法

有两条路，建议二选一：

#### 方案 A：新增 review.summary / review.explain 任务

新增语义更准确的 AI 任务：

- `review.summary`
- `review.explain`

这类任务的输入应是：

- review task 基本信息
- findings 列表
- summary
- risk level
- 用户额外问题

而不是伪装成 diff。

#### 方案 B：如果坚持复用 /review

那就必须传入真实 diff，包括：

- 原始 merge request diff
- changed files
- 原始变更上下文

否则这个接口名本身就不成立。

### 推荐方案

推荐 **方案 A**。  
因为当前 ReviewTaskDetailPage 的交互目标明显更偏向：

- “帮我解读这次审查结果”
- “帮我总结风险”
- “帮我整理 reviewer 视角”

这不是代码审查本身，而是审查结果分析。

### 验收标准

- Review 页面使用的 AI 任务名称与页面语义一致
- 不再把 findings 文本伪装成 `diff`
- Review 页面与真正的 code review 接口职责清晰分离

---

## 3. AIReviewResponse 前端类型还没真正收口

### 问题描述

当前 [frontend/src/types/ai.ts](/opt/project/wechat_project/git-manage-service/frontend/src/types/ai.ts) 中，`AIReviewResponse` 仍然保留两个字段：

- `high`
- `highRisk`

而当前后端和 prompt 语义已经统一到了：

- `highRisk`

### 影响

1. 调用方还会继续写兼容逻辑
2. 前端难以及时发现后端是否又错误返回旧字段
3. 契约长期保持“半兼容”状态，不利于收口

### 根因

为了兼容旧返回格式，前端保留了双字段，但目前已经过了“临时兼容”的最佳时机。

### 建议改法

1. 删除前端 `AIReviewResponse.high`
2. 页面中删除类似：
   - `response.highRisk || response.high || []`
3. 统一只使用：
   - `response.highRisk`

如果担心后端历史数据，可以短期通过解析层做兼容，但不要把双字段长期留在主类型定义里。

### 推荐策略

建议：

1. 主类型立即收口
2. 兼容逻辑只放在单个适配层
3. 页面逻辑不再直接处理双字段

### 验收标准

- `frontend/src/types/ai.ts` 中只保留 `highRisk`
- 页面逻辑不再读取 `high`
- AI review 相关响应契约前后端一致

---

## 4. 反馈提交流程仍然是乐观更新且失败无回滚

### 问题描述

当前 [AIPanel.vue](/opt/project/wechat_project/git-manage-service/frontend/src/components/ai/AIPanel.vue) 中的 `sendFeedback()` 流程仍然是：

1. 先把 `msg.feedback` 写到本地
2. 再请求 `/api/v1/ai/feedback`

如果请求失败：

- UI 上已经看不到反馈按钮
- 用户以为反馈成功
- 实际并没有写入后端
- 用户也无法再次尝试

### 影响

1. 前端状态和后端真实状态不一致
2. `user_feedback` 审计字段可能丢数据
3. 用户没有失败感知，也没有重试能力

### 根因

当前实现采用了“乐观更新”，但没有失败回滚机制。

### 建议改法

推荐两种实现方案：

#### 方案 A：请求成功后再写本地状态

流程：

1. 用户点击反馈按钮
2. 发请求
3. 请求成功后：
   - 写入 `msg.feedback`
   - 隐藏反馈按钮
4. 请求失败后：
   - 保持原状态
   - 提示失败

#### 方案 B：保留乐观更新，但失败时回滚

流程：

1. 先记录旧状态
2. 先写 `msg.feedback`
3. 请求失败时恢复旧状态
4. 给用户提示可重试

### 推荐方案

推荐 **方案 A**，因为：

- 逻辑最清晰
- UI 状态和后端状态天然一致
- 不需要维护额外回滚逻辑

### 验收标准

- `/api/v1/ai/feedback` 请求失败时，用户仍可再次点击反馈
- UI 不会错误显示“已经反馈成功”
- 反馈提交结果对用户可感知

---

## 建议处理顺序

### 第一轮

1. 修 `AI_API.md` 文档映射漂移
2. 收口 `AIReviewResponse` 前端类型

### 第二轮

1. 让 Review 页面切到正确语义的 AI 任务
2. 不再把 findings 文本伪装成 diff

### 第三轮

1. 修 `AIPanel` 反馈提交流程
2. 增加失败提示与可重试能力

---

## 当前结论

现在 AI 主链路已经具备这些能力：

- 可请求
- 可回写面板
- 可记录 invocationId
- 可提交用户反馈

但后续如果不继续收口，风险会集中在 4 个方向：

1. 文档越来越不可信
2. Review 能力的语义越来越模糊
3. 前后端契约长期双轨
4. 用户反馈数据看起来有、实际上不稳定

因此，下一步不建议继续扩新能力，优先把这 4 个点彻底收干净。
