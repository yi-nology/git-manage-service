# AI 后续收口清单

这份文档整理了当前 AI 接入在完成主链路之后，仍然需要继续收口的 4 个问题。  
这些问题不一定会直接导致编译失败，但会影响联调准确性、契约清晰度和后续维护成本。

---

## 1. AI API 文档与真实实现漂移

### 问题

当前 [AI_API.md](/opt/project/wechat_project/git-manage-service/docs/AI_API.md) 中的“页面-任务-API 映射表”已经和真实代码不一致。

典型例子：

- 文档写的是 `POST /api/v1/ai/review/summary`
- 当前实际注册和接入的是 `POST /api/v1/ai/review`

此外，文档中还列出了一些当前代码里并不存在的内容，例如：

- `/api/v1/ai/review/reply-draft`
- `/api/v1/ai/branch/premerge-analysis`
- `PatchDetailPage`
- `StatsDashboardPage`

### 影响

- 前后端联调会被文档误导
- 后续新增页面容易建立在错误契约上
- “已实现”和“规划中”混在一起，团队很难判断真实落地状态

### 建议改法

把文档拆成两层：

1. **已实现**
   - 只列当前代码中真实存在、已注册、已有前端调用的接口
2. **规划中**
   - 单独列未来打算做的能力

建议文档里明确增加一列：

- `状态`: `已实现 / 部分实现 / 规划中`

### 验收标准

- 文档中的每个 endpoint 都能在代码里找到实际注册
- 文档中的每个页面都能在 `frontend/src/views` 中找到对应实现
- 不再出现“文档写了，但代码没有”的 AI 路由映射

---

## 2. 前端 AIReviewResponse 契约还没有完全收口

### 问题

当前前端 [frontend/src/types/ai.ts](/opt/project/wechat_project/git-manage-service/frontend/src/types/ai.ts) 中，`AIReviewResponse` 还同时保留了两个字段：

- `high`
- `highRisk`

而当前后端 DTO、prompt 和返回语义已经统一到 `highRisk`。

### 影响

- 调用方会继续写兼容逻辑
- 容易掩盖后端错误返回旧字段的问题
- 后续维护者无法快速确认“唯一真实字段”是什么

### 建议改法

1. 前端类型只保留 `highRisk`
2. 页面逻辑里去掉 `response.highRisk || response.high || []` 这类兼容分支
3. 如果确实要保留兼容，兼容层应明确写注释并设置删除时机

### 验收标准

- `frontend/src/types/ai.ts` 中 `AIReviewResponse` 只保留单一高风险字段
- 页面逻辑不再依赖双字段兜底
- Review 相关响应契约前后端一致

---

## 3. AIPanel 反馈提交流程还不够可靠

### 问题

当前 [frontend/src/components/ai/AIPanel.vue](/opt/project/wechat_project/git-manage-service/frontend/src/components/ai/AIPanel.vue) 中，`sendFeedback()` 的处理顺序是：

1. 先把 `msg.feedback` 写到本地状态
2. 再请求 `/api/v1/ai/feedback`

如果后端请求失败：

- UI 仍然会显示用户已经点过赞/踩
- 按钮不会再次出现
- 用户没有重试机会

### 影响

- 用户以为反馈成功，实际上没有持久化
- `user_feedback` 审计数据会丢
- 前端状态和后端真实状态不一致

### 建议改法

推荐改成下面两种之一：

#### 方案 A：成功后再落本地状态

1. 点击按钮
2. 发起反馈请求
3. 请求成功后再写 `msg.feedback`

#### 方案 B：先乐观更新，但失败时回滚

1. 先写本地状态
2. 请求失败时回滚 `msg.feedback`
3. 给用户明确错误提示
4. 允许再次点击重试

更推荐 **方案 A**，逻辑更干净。

### 验收标准

- 反馈请求失败时，用户仍然可以再次提交
- UI 状态和后端持久化结果一致
- 用户能明确知道反馈成功还是失败

---

## 4. AI Service 响应组装代码存在重复赋值噪音

### 问题

当前 [biz/service/ai/service.go](/opt/project/wechat_project/git-manage-service/biz/service/ai/service.go) 中，多处存在这种模式：

- `result.Raw = resp.Content`
- `result.InvocationID = resp.InvocationID`

而且有些地方出现重复赋值，说明这部分代码是多次复制后残留下来的。

### 影响

- 增加阅读噪音
- 让维护者误以为这里有特殊补丁逻辑
- 后续新增 AI task 时容易继续复制出同类问题

### 建议改法

抽一个统一 helper，例如：

- `fillAdviceResponseMeta(...)`
- `fillDraftResponseMeta(...)`
- `fillDiagnosisResponseMeta(...)`
- `fillReviewResponseMeta(...)`

或者抽一个更通用的内部 helper，只负责填：

- `Raw`
- `InvocationID`

目标不是做大重构，而是把重复赋值清干净。

### 验收标准

- `biz/service/ai/service.go` 中不再出现明显重复的 `InvocationID` / `Raw` 赋值
- 新旧任务的响应填充方式一致
- response 组装逻辑更容易审查

---

## 推荐处理顺序

### 第一轮

1. 修文档映射漂移
2. 收口 `AIReviewResponse` 字段契约

### 第二轮

1. 修 `AIPanel` 的反馈持久化流程
2. 增加失败提示和重试能力

### 第三轮

1. 清理 `biz/service/ai/service.go` 中的重复元信息赋值
2. 抽小 helper 统一响应填充

---

## 当前结论

当前 AI 主链路已经具备“可调用、可返回、可反馈”的基础能力。  
剩下这批问题主要属于：

- **文档与实现一致性**
- **接口契约收口**
- **交互可靠性**
- **代码可维护性**

这轮建议不要再继续扩新能力，优先把这 4 个点收干净。  
收完之后，后面的 AI 迭代会轻很多。
