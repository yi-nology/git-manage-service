# Code Review 模块改进计划

> 基于 [Code-Review-GPT-Gitlab](https://github.com/mimo-x/Code-Review-GPT-Gitlab) 项目的设计理念，结合当前项目架构进行改进。

## 核心约束

1. **使用 IDL 约束**：所有接口和数据结构通过 Proto 文件定义
2. **Hz 代码生成**：使用 `make hz-gen` 生成 Handler/Router/Model 代码
3. **减少 PO/VO**：复用 IDL 定义的 message，减少手写的 PO 和 VO
4. **保留原有模式**：兼容原有的 LLM API 调用模式

---

## 一、架构改进概览

### 当前架构

```
Webhook → Handler → Service → DAO → SQLite
              ↓           ↓
         Provider    Rule Engine + LLM
```

### 目标架构

```
Webhook → Event Router → Review Engine → Result Processor → Notification
              ↓               ↓               ↓
         Event Rules     CLI/API/LLM      Multi-channel
                        (可插拔模式)
```

### 审查模式

| 模式 | 说明 | 状态 |
|------|------|------|
| `llm` | 原有的 LLM API 调用模式 | 保留 |
| `claude_cli` | 使用 Claude CLI 进行本地仓库审查 | 新增 |
| `opencode_cli` | 使用 opencode CLI 进行审查 | 新增 |
| `qoder_cli` | 使用 qoder CLI 进行审查 | 新增 |
| `codex_cli` | 使用 codex CLI 进行审查 | 新增 |
| `hybrid` | 混合模式：先规则审查，再 LLM 审查 | 增强 |

---

## 二、IDL 文件更新

### 2.1 增强 `webhook_event.proto`

**文件**：`idl/biz/webhook_event.proto`

**新增内容**：

```protobuf
// WebhookEventRule 事件规则
message WebhookEventRule {
  uint64 id = 1;
  string name = 2;
  string event_type = 3;
  string description = 4;
  string match_rules = 5;        // JSON格式的匹配规则
  bool is_active = 6;
  int32 priority = 7;            // 规则优先级
  string created_at = 8;
  string updated_at = 9;
}

// CreateEventRuleRequest 创建事件规则请求
message CreateEventRuleRequest {
  string name = 1 [(api.body) = "name"];
  string event_type = 2 [(api.body) = "event_type"];
  string description = 3 [(api.body) = "description"];
  string match_rules = 4 [(api.body) = "match_rules"];
  bool is_active = 5 [(api.body) = "is_active"];
  int32 priority = 6 [(api.body) = "priority"];
}

// UpdateEventRuleRequest 更新事件规则请求
message UpdateEventRuleRequest {
  uint64 id = 1 [(api.path) = "id"];
  string name = 2 [(api.body) = "name"];
  string event_type = 3 [(api.body) = "event_type"];
  string description = 4 [(api.body) = "description"];
  string match_rules = 5 [(api.body) = "match_rules"];
  bool is_active = 6 [(api.body) = "is_active"];
  int32 priority = 7 [(api.body) = "priority"];
}

// ListEventRulesRequest 列出事件规则请求
message ListEventRulesRequest {
  string event_type = 1 [(api.query) = "event_type"];
  bool is_active = 2 [(api.query) = "is_active"];
  int32 page = 3 [(api.query) = "page"];
  int32 page_size = 4 [(api.query) = "page_size"];
}

// ListEventRulesResponse 列出事件规则响应
message ListEventRulesResponse {
  common.BaseResponse base = 1;
  repeated WebhookEventRule rules = 2;
  int64 total = 3;
}

// EventRuleResponse 事件规则响应
message EventRuleResponse {
  common.BaseResponse base = 1;
  WebhookEventRule rule = 2;
}

// 扩展 WebhookEventService
service WebhookEventService {
  // ... 现有接口 ...
  
  // ListRules 列出事件规则
  rpc ListRules(ListEventRulesRequest) returns (ListEventRulesResponse) {
    option (api.get) = "/api/v1/webhook/event-rules";
  }
  // GetRule 获取事件规则详情
  rpc GetRule(GetReviewTaskRequest) returns (EventRuleResponse) {
    option (api.get) = "/api/v1/webhook/event-rules/:id";
  }
  // CreateRule 创建事件规则
  rpc CreateRule(CreateEventRuleRequest) returns (EventRuleResponse) {
    option (api.post) = "/api/v1/webhook/event-rules";
  }
  // UpdateRule 更新事件规则
  rpc UpdateRule(UpdateEventRuleRequest) returns (EventRuleResponse) {
    option (api.put) = "/api/v1/webhook/event-rules/:id";
  }
  // DeleteRule 删除事件规则
  rpc DeleteRule(GetReviewTaskRequest) returns (common.EmptyResponse) {
    option (api.delete) = "/api/v1/webhook/event-rules/:id";
  }
}
```

### 2.2 增强 `review.proto`

**文件**：`idl/biz/review.proto`

**新增内容**：

```protobuf
// ReviewMode 审查模式枚举
enum ReviewMode {
  REVIEW_MODE_LLM = 0;           // 原有的 LLM API 调用模式
  REVIEW_MODE_CLAUDE_CLI = 1;    // Claude CLI 模式
  REVIEW_MODE_OPENCODE_CLI = 2;  // opencode CLI 模式
  REVIEW_MODE_QODER_CLI = 3;     // qoder CLI 模式
  REVIEW_MODE_CODEX_CLI = 4;     // codex CLI 模式
  REVIEW_MODE_HYBRID = 5;        // 混合模式
}

// ReviewCLIConfig CLI 配置
message ReviewCLIConfig {
  uint64 id = 1;
  string name = 2;
  string cli_type = 3;           // claude, opencode, qoder, codex
  string exec_path = 4;          // CLI 可执行文件路径
  string config_json = 5;        // CLI 特定配置
  bool is_active = 6;
  string created_at = 7;
  string updated_at = 8;
}

// 增强 ReviewRepoConfig
message ReviewRepoConfig {
  uint64 id = 1;
  uint64 provider_config_id = 2;
  string platform_owner = 3;
  string platform_repo = 4;
  bool enabled = 5;
  bool block_on_high = 6;
  bool auto_review_on_mr = 7;
  string llm_provider = 8;
  int32 max_files = 9;
  int32 max_diff_lines = 10;
  string rule_overrides_json = 11;
  string scope_note = 12;
  string created_at = 13;
  string updated_at = 14;
  
  // 新增：CLI 配置
  string review_mode = 15;       // 审查模式：llm, claude_cli, opencode_cli, qoder_cli, codex_cli, hybrid
  string cli_config_json = 16;   // CLI 特定配置（JSON）
  
  // 新增：自定义 Prompt
  string custom_prompt = 17;     // 自定义审查 Prompt
  bool use_custom_prompt = 18;   // 是否使用自定义 Prompt
  
  // 新增：文件过滤配置
  string exclude_file_types = 19; // 排除的文件类型（JSON数组）
  string ignore_patterns = 20;    // 忽略的文件模式（JSON数组）
}

// CreateCLIConfigRequest 创建 CLI 配置请求
message CreateCLIConfigRequest {
  string name = 1 [(api.body) = "name"];
  string cli_type = 2 [(api.body) = "cli_type"];
  string exec_path = 3 [(api.body) = "exec_path"];
  string config_json = 4 [(api.body) = "config_json"];
  bool is_active = 5 [(api.body) = "is_active"];
}

// UpdateCLIConfigRequest 更新 CLI 配置请求
message UpdateCLIConfigRequest {
  uint64 id = 1 [(api.path) = "id"];
  string name = 2 [(api.body) = "name"];
  string cli_type = 3 [(api.body) = "cli_type"];
  string exec_path = 4 [(api.body) = "exec_path"];
  string config_json = 5 [(api.body) = "config_json"];
  bool is_active = 6 [(api.body) = "is_active"];
}

// ListCLIConfigsResponse 列出 CLI 配置响应
message ListCLIConfigsResponse {
  common.BaseResponse base = 1;
  repeated ReviewCLIConfig configs = 2;
}

// CLIConfigResponse CLI 配置响应
message CLIConfigResponse {
  common.BaseResponse base = 1;
  ReviewCLIConfig config = 2;
}

// TestCLIConfigRequest 测试 CLI 配置请求
message TestCLIConfigRequest {
  uint64 id = 1 [(api.body) = "id"];
}

// TestCLIConfigResponse 测试 CLI 配置响应
message TestCLIConfigResponse {
  common.BaseResponse base = 1;
  bool success = 2;
  string message = 3;
  string version = 4;            // CLI 版本信息
}

// 扩展 ReviewService
service ReviewService {
  // ... 现有接口 ...
  
  // ListCLIConfigs 列出 CLI 配置
  rpc ListCLIConfigs(common.EmptyRequest) returns (ListCLIConfigsResponse) {
    option (api.get) = "/api/v1/reviews/cli-configs";
  }
  // GetCLIConfig 获取 CLI 配置详情
  rpc GetCLIConfig(GetReviewTaskRequest) returns (CLIConfigResponse) {
    option (api.get) = "/api/v1/reviews/cli-configs/:id";
  }
  // CreateCLIConfig 创建 CLI 配置
  rpc CreateCLIConfig(CreateCLIConfigRequest) returns (CLIConfigResponse) {
    option (api.post) = "/api/v1/reviews/cli-configs";
  }
  // UpdateCLIConfig 更新 CLI 配置
  rpc UpdateCLIConfig(UpdateCLIConfigRequest) returns (CLIConfigResponse) {
    option (api.put) = "/api/v1/reviews/cli-configs/:id";
  }
  // DeleteCLIConfig 删除 CLI 配置
  rpc DeleteCLIConfig(GetReviewTaskRequest) returns (common.EmptyResponse) {
    option (api.delete) = "/api/v1/reviews/cli-configs/:id";
  }
  // TestCLIConfig 测试 CLI 配置
  rpc TestCLIConfig(TestCLIConfigRequest) returns (TestCLIConfigResponse) {
    option (api.post) = "/api/v1/reviews/cli-configs/:id/test";
  }
}
```

### 2.3 增强 `audit.proto`

**文件**：`idl/biz/audit.proto`

**新增内容**：

```protobuf
// ReviewAuditLog 审计日志
message ReviewAuditLog {
  uint64 id = 1;
  uint64 task_id = 2;
  string action = 3;             // created, started, completed, failed, retry
  string status = 4;
  string error_message = 5;
  int32 duration = 6;            // 执行时长（秒）
  string metadata = 7;           // JSON 格式的元数据
  string created_at = 8;
}

// ListReviewAuditLogsRequest 列出审计日志请求
message ListReviewAuditLogsRequest {
  uint64 task_id = 1 [(api.query) = "task_id"];
  string action = 2 [(api.query) = "action"];
  string status = 3 [(api.query) = "status"];
  string start_time = 4 [(api.query) = "start_time"];
  string end_time = 5 [(api.query) = "end_time"];
  int32 page = 6 [(api.query) = "page"];
  int32 page_size = 7 [(api.query) = "page_size"];
}

// ListReviewAuditLogsResponse 列出审计日志响应
message ListReviewAuditLogsResponse {
  common.BaseResponse base = 1;
  repeated ReviewAuditLog logs = 2;
  int64 total = 3;
}

// 扩展 AuditService（如果存在）
service AuditService {
  // ... 现有接口 ...
  
  // ListReviewAuditLogs 列出审查审计日志
  rpc ListReviewAuditLogs(ListReviewAuditLogsRequest) returns (ListReviewAuditLogsResponse) {
    option (api.get) = "/api/v1/audit/review-logs";
  }
}
```

---

## 三、新增文件

### 3.1 CLI 服务接口

**文件**：`biz/service/codereview/cli_service.go`

```go
package codereview

import "context"

// CLIService CLI 服务接口
type CLIService interface {
    // ReviewCode 使用 CLI 进行代码审查
    ReviewCode(ctx context.Context, req *CLIReviewRequest) (*CLIReviewResult, error)
    
    // ValidateInstallation 验证 CLI 安装
    ValidateInstallation() error
    
    // GetVersion 获取 CLI 版本
    GetVersion() (string, error)
}

// CLIReviewRequest CLI 审查请求
type CLIReviewRequest struct {
    RepoPath     string // 本地仓库路径
    CommitRange  string // Git 提交范围（如 HEAD~1..HEAD）
    CustomPrompt string // 自定义 Prompt
    MaxTokens    int    // 最大 token 数
    WorkingDir   string // 工作目录
}

// CLIReviewResult CLI 审查结果
type CLIReviewResult struct {
    Content    string           // 原始审查内容
    Score      int              // 评分（0-100）
    Issues     []CLIReviewIssue // 问题列表
    Summary    string           // 总结
    Duration   int              // 执行时长（秒）
}

// CLIReviewIssue CLI 审查问题
type CLIReviewIssue struct {
    FilePath   string `json:"file_path"`
    LineNumber int    `json:"line_number"`
    Severity   string `json:"severity"`   // critical, high, medium, low, info
    Title      string `json:"title"`
    Message    string `json:"message"`
    Suggestion string `json:"suggestion"`
    RuleID     string `json:"rule_id"`
}
```

### 3.2 CLI 服务工厂

**文件**：`biz/service/codereview/cli_factory.go`

```go
package codereview

import "fmt"

// NewCLIService 创建 CLI 服务
func NewCLIService(cliType string, config map[string]interface{}) (CLIService, error) {
    switch cliType {
    case "claude":
        return NewClaudeCLIService(config), nil
    case "opencode":
        return NewOpenCodeCLIService(config), nil
    case "qoder":
        return NewQoderCLIService(config), nil
    case "codex":
        return NewCodexCLIService(config), nil
    default:
        return nil, fmt.Errorf("unsupported CLI type: %s", cliType)
    }
}

// GetSupportedCLITypes 获取支持的 CLI 类型
func GetSupportedCLITypes() []string {
    return []string{"claude", "opencode", "qoder", "codex"}
}
```

### 3.3 Claude CLI 服务

**文件**：`biz/service/codereview/claude_cli_service.go`

```go
package codereview

import (
    "context"
    "fmt"
    "os/exec"
    "strings"
)

// ClaudeCLIService Claude CLI 服务
type ClaudeCLIService struct {
    execPath string
    config   map[string]interface{}
}

// NewClaudeCLIService 创建 Claude CLI 服务
func NewClaudeCLIService(config map[string]interface{}) *ClaudeCLIService {
    execPath := "claude"
    if path, ok := config["exec_path"].(string); ok && path != "" {
        execPath = path
    }
    return &ClaudeCLIService{
        execPath: execPath,
        config:   config,
    }
}

// ReviewCode 使用 Claude CLI 进行代码审查
func (s *ClaudeCLIService) ReviewCode(ctx context.Context, req *CLIReviewRequest) (*CLIReviewResult, error) {
    args := []string{
        "--print",
        "--output-format", "json",
    }
    
    if req.CustomPrompt != "" {
        args = append(args, "--prompt", req.CustomPrompt)
    }
    
    if req.CommitRange != "" {
        args = append(args, "--commit-range", req.CommitRange)
    }
    
    cmd := exec.CommandContext(ctx, s.execPath, args...)
    cmd.Dir = req.RepoPath
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("claude CLI execution failed: %w, output: %s", err, string(output))
    }
    
    return s.parseOutput(string(output))
}

// ValidateInstallation 验证 Claude CLI 安装
func (s *ClaudeCLIService) ValidateInstallation() error {
    cmd := exec.Command(s.execPath, "--version")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("claude CLI not found: %w", err)
    }
    
    if !strings.Contains(string(output), "claude") {
        return fmt.Errorf("invalid claude CLI output: %s", string(output))
    }
    
    return nil
}

// GetVersion 获取 Claude CLI 版本
func (s *ClaudeCLIService) GetVersion() (string, error) {
    cmd := exec.Command(s.execPath, "--version")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}

// parseOutput 解析 Claude CLI 输出
func (s *ClaudeCLIService) parseOutput(output string) (*CLIReviewResult, error) {
    // TODO: 实现 JSON 输出解析
    return &CLIReviewResult{
        Content: output,
    }, nil
}
```

### 3.4 OpenCode CLI 服务

**文件**：`biz/service/codereview/opencode_cli_service.go`

```go
package codereview

import (
    "context"
    "fmt"
    "os/exec"
    "strings"
)

// OpenCodeCLIService OpenCode CLI 服务
type OpenCodeCLIService struct {
    execPath string
    config   map[string]interface{}
}

// NewOpenCodeCLIService 创建 OpenCode CLI 服务
func NewOpenCodeCLIService(config map[string]interface{}) *OpenCodeCLIService {
    execPath := "opencode"
    if path, ok := config["exec_path"].(string); ok && path != "" {
        execPath = path
    }
    return &OpenCodeCLIService{
        execPath: execPath,
        config:   config,
    }
}

// ReviewCode 使用 OpenCode CLI 进行代码审查
func (s *OpenCodeCLIService) ReviewCode(ctx context.Context, req *CLIReviewRequest) (*CLIReviewResult, error) {
    // 使用 opencode 的非交互模式
    args := []string{
        "--non-interactive",
        "--format", "json",
    }
    
    if req.CustomPrompt != "" {
        args = append(args, "--prompt", req.CustomPrompt)
    }
    
    cmd := exec.CommandContext(ctx, s.execPath, args...)
    cmd.Dir = req.RepoPath
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("opencode CLI execution failed: %w, output: %s", err, string(output))
    }
    
    return s.parseOutput(string(output))
}

// ValidateInstallation 验证 OpenCode CLI 安装
func (s *OpenCodeCLIService) ValidateInstallation() error {
    cmd := exec.Command(s.execPath, "version")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("opencode CLI not found: %w", err)
    }
    
    if len(output) == 0 {
        return fmt.Errorf("invalid opencode CLI output")
    }
    
    return nil
}

// GetVersion 获取 OpenCode CLI 版本
func (s *OpenCodeCLIService) GetVersion() (string, error) {
    cmd := exec.Command(s.execPath, "version")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}

// parseOutput 解析 OpenCode CLI 输出
func (s *OpenCodeCLIService) parseOutput(output string) (*CLIReviewResult, error) {
    // TODO: 实现 JSON 输出解析
    return &CLIReviewResult{
        Content: output,
    }, nil
}
```

### 3.5 Qoder CLI 服务

**文件**：`biz/service/codereview/qoder_cli_service.go`

```go
package codereview

import (
    "context"
    "fmt"
    "os/exec"
    "strings"
)

// QoderCLIService Qoder CLI 服务
type QoderCLIService struct {
    execPath string
    config   map[string]interface{}
}

// NewQoderCLIService 创建 Qoder CLI 服务
func NewQoderCLIService(config map[string]interface{}) *QoderCLIService {
    execPath := "qoder"
    if path, ok := config["exec_path"].(string); ok && path != "" {
        execPath = path
    }
    return &QoderCLIService{
        execPath: execPath,
        config:   config,
    }
}

// ReviewCode 使用 Qoder CLI 进行代码审查
func (s *QoderCLIService) ReviewCode(ctx context.Context, req *CLIReviewRequest) (*CLIReviewResult, error) {
    args := []string{
        "review",
        "--format", "json",
    }
    
    if req.CustomPrompt != "" {
        args = append(args, "--prompt", req.CustomPrompt)
    }
    
    if req.CommitRange != "" {
        args = append(args, "--range", req.CommitRange)
    }
    
    cmd := exec.CommandContext(ctx, s.execPath, args...)
    cmd.Dir = req.RepoPath
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("qoder CLI execution failed: %w, output: %s", err, string(output))
    }
    
    return s.parseOutput(string(output))
}

// ValidateInstallation 验证 Qoder CLI 安装
func (s *QoderCLIService) ValidateInstallation() error {
    cmd := exec.Command(s.execPath, "version")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("qoder CLI not found: %w", err)
    }
    
    if len(output) == 0 {
        return fmt.Errorf("invalid qoder CLI output")
    }
    
    return nil
}

// GetVersion 获取 Qoder CLI 版本
func (s *QoderCLIService) GetVersion() (string, error) {
    cmd := exec.Command(s.execPath, "version")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}

// parseOutput 解析 Qoder CLI 输出
func (s *QoderCLIService) parseOutput(output string) (*CLIReviewResult, error) {
    // TODO: 实现 JSON 输出解析
    return &CLIReviewResult{
        Content: output,
    }, nil
}
```

### 3.6 Codex CLI 服务

**文件**：`biz/service/codereview/codex_cli_service.go`

```go
package codereview

import (
    "context"
    "fmt"
    "os/exec"
    "strings"
)

// CodexCLIService Codex CLI 服务
type CodexCLIService struct {
    execPath string
    config   map[string]interface{}
}

// NewCodexCLIService 创建 Codex CLI 服务
func NewCodexCLIService(config map[string]interface{}) *CodexCLIService {
    execPath := "codex"
    if path, ok := config["exec_path"].(string); ok && path != "" {
        execPath = path
    }
    return &CodexCLIService{
        execPath: execPath,
        config:   config,
    }
}

// ReviewCode 使用 Codex CLI 进行代码审查
func (s *CodexCLIService) ReviewCode(ctx context.Context, req *CLIReviewRequest) (*CLIReviewResult, error) {
    args := []string{
        "--quiet",
        "--format", "json",
    }
    
    if req.CustomPrompt != "" {
        args = append(args, "--prompt", req.CustomPrompt)
    }
    
    cmd := exec.CommandContext(ctx, s.execPath, args...)
    cmd.Dir = req.RepoPath
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("codex CLI execution failed: %w, output: %s", err, string(output))
    }
    
    return s.parseOutput(string(output))
}

// ValidateInstallation 验证 Codex CLI 安装
func (s *CodexCLIService) ValidateInstallation() error {
    cmd := exec.Command(s.execPath, "version")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("codex CLI not found: %w", err)
    }
    
    if len(output) == 0 {
        return fmt.Errorf("invalid codex CLI output")
    }
    
    return nil
}

// GetVersion 获取 Codex CLI 版本
func (s *CodexCLIService) GetVersion() (string, error) {
    cmd := exec.Command(s.execPath, "version")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}

// parseOutput 解析 Codex CLI 输出
func (s *CodexCLIService) parseOutput(output string) (*CLIReviewResult, error) {
    // TODO: 实现 JSON 输出解析
    return &CLIReviewResult{
        Content: output,
    }, nil
}
```

---

## 四、修改现有文件

### 4.1 增强 `review_service.go`

**文件**：`biz/service/codereview/review_service.go`

**修改内容**：

1. 添加审查模式支持
2. 添加 CLI 审查执行函数

```go
// 在 executeReview 函数中添加模式选择
func executeReview(ctx context.Context, task *po.ReviewTask, params *reviewParams, taskDAO *db.ReviewTaskDAO) (*AggregatedResult, string, error) {
    // ... 现有代码 ...
    
    // 根据配置的审查模式选择执行方式
    cfg := GetMergedConfig(params.repoID)
    
    switch cfg.ReviewMode {
    case "claude_cli", "opencode_cli", "qoder_cli", "codex_cli":
        return executeCLIReview(ctx, task, params, cfg, taskDAO)
    case "hybrid":
        return executeHybridReview(ctx, task, params, cfg, taskDAO)
    default: // "llm"
        return executeLLMReview(ctx, task, params, cfg, taskDAO)
    }
}

// executeCLIReview 执行 CLI 审查
func executeCLIReview(ctx context.Context, task *po.ReviewTask, params *reviewParams, cfg reviewConfig, taskDAO *db.ReviewTaskDAO) (*AggregatedResult, string, error) {
    var processLog []*ProcessStep
    
    // 1. 创建 CLI 服务
    cliService, err := NewCLIService(cfg.ReviewMode, cfg.CLIConfig)
    if err != nil {
        return nil, "", fmt.Errorf("failed to create CLI service: %w", err)
    }
    
    processLog = append(processLog, &ProcessStep{
        Name:   "Init CLI Service",
        Status: "ok",
        Detail: fmt.Sprintf("CLI type: %s", cfg.ReviewMode),
    })
    
    // 2. 验证 CLI 安装
    if err := cliService.ValidateInstallation(); err != nil {
        return nil, "", fmt.Errorf("CLI validation failed: %w", err)
    }
    
    // 3. 克隆/更新仓库（如果需要）
    repoPath, err := prepareLocalRepo(ctx, params)
    if err != nil {
        return nil, "", fmt.Errorf("failed to prepare repo: %w", err)
    }
    
    processLog = append(processLog, &ProcessStep{
        Name:   "Prepare Repository",
        Status: "ok",
        Detail: fmt.Sprintf("Repo path: %s", repoPath),
    })
    
    // 4. 执行 CLI 审查
    commitRange := fmt.Sprintf("HEAD~1..HEAD") // 可以根据实际情况调整
    if task.CommitSHA != "" {
        commitRange = task.CommitSHA + "^.." + task.CommitSHA
    }
    
    result, err := cliService.ReviewCode(ctx, &CLIReviewRequest{
        RepoPath:     repoPath,
        CommitRange:  commitRange,
        CustomPrompt: cfg.CustomPrompt,
    })
    if err != nil {
        return nil, "", fmt.Errorf("CLI review failed: %w", err)
    }
    
    processLog = append(processLog, &ProcessStep{
        Name:   "CLI Review",
        Status: "ok",
        Detail: fmt.Sprintf("Duration: %d seconds, Issues: %d", result.Duration, len(result.Issues)),
    })
    
    // 5. 转换为标准格式
    findings := convertCLIIssuesToFindings(result.Issues)
    
    // 6. 聚合结果
    aggregated := Aggregate(findings, 0, 0, 0, cfg.BlockOnHigh, processLog)
    
    return aggregated, result.Content, nil
}

// executeHybridReview 执行混合审查
func executeHybridReview(ctx context.Context, task *po.ReviewTask, params *reviewParams, cfg reviewConfig, taskDAO *db.ReviewTaskDAO) (*AggregatedResult, string, error) {
    var allFindings []*Finding
    var processLog []*ProcessStep
    var rawDiff string
    
    // 1. 先执行规则审查
    // ... 现有规则审查代码 ...
    
    // 2. 再执行 LLM 审查
    // ... 现有 LLM 审查代码 ...
    
    // 3. 聚合结果
    return Aggregate(allFindings, 0, 0, 0, cfg.BlockOnHigh, processLog), rawDiff, nil
}

// convertCLIIssuesToFindings 转换 CLI 问题为标准格式
func convertCLIIssuesToFindings(issues []CLIReviewIssue) []*Finding {
    findings := make([]*Finding, 0, len(issues))
    for _, issue := range issues {
        severity := SeverityMedium
        switch issue.Severity {
        case "critical":
            severity = SeverityCritical
        case "high":
            severity = SeverityHigh
        case "medium":
            severity = SeverityMedium
        case "low":
            severity = SeverityLow
        case "info":
            severity = SeverityInfo
        }
        
        findings = append(findings, &Finding{
            RuleID:      issue.RuleID,
            Source:      "cli",
            Severity:    severity,
            FilePath:    issue.FilePath,
            NewLine:     issue.LineNumber,
            Title:       issue.Title,
            Message:     issue.Message,
            Suggestion:  issue.Suggestion,
            Fingerprint: computeFingerprint(issue.RuleID, issue.FilePath, issue.LineNumber, issue.Title),
        })
    }
    return findings
}
```

### 4.2 增强 `review_service_config.go`

**文件**：`biz/service/codereview/review_service_config.go`

**修改内容**：

1. 增强 `reviewConfig` 结构体
2. 增强 `GetMergedConfig` 函数

```go
// reviewConfig 审查配置
type reviewConfig struct {
    BlockOnHigh      bool
    MaxFiles         int
    MaxDiffLines     int
    LLMProvider      string
    AutoReviewOnMR   bool
    
    // 新增：CLI 配置
    ReviewMode       string                 // 审查模式
    CLIConfig        map[string]interface{} // CLI 特定配置
    
    // 新增：自定义 Prompt
    CustomPrompt     string
    UseCustomPrompt  bool
    
    // 新增：文件过滤
    ExcludeFileTypes []string
    IgnorePatterns   []string
}

// GetMergedConfig 获取合并后的配置
func GetMergedConfig(repoID uint) reviewConfig {
    repo, err := db.NewRepoDAO().FindByID(repoID)
    if err != nil {
        return getConfig()
    }
    
    cfg := getConfig()
    
    if repo.ProviderConfigID == 0 || repo.PlatformOwner == "" || repo.PlatformRepo == "" {
        return cfg
    }
    
    repoCfg, err := db.NewReviewRepoConfigDAO().FindByRemoteRepo(repo.ProviderConfigID, repo.PlatformOwner, repo.PlatformRepo)
    if err != nil {
        return cfg
    }
    
    if !repoCfg.Enabled {
        cfg.BlockOnHigh = false
        cfg.AutoReviewOnMR = false
        return cfg
    }
    
    // 合并配置
    cfg.BlockOnHigh = repoCfg.BlockOnHigh
    cfg.AutoReviewOnMR = repoCfg.AutoReviewOnMR
    if repoCfg.MaxFiles > 0 {
        cfg.MaxFiles = repoCfg.MaxFiles
    }
    if repoCfg.MaxDiffLines > 0 {
        cfg.MaxDiffLines = repoCfg.MaxDiffLines
    }
    if repoCfg.LLMProvider != "" {
        cfg.LLMProvider = repoCfg.LLMProvider
    }
    
    // 新增：CLI 配置
    if repoCfg.ReviewMode != "" {
        cfg.ReviewMode = repoCfg.ReviewMode
    }
    if repoCfg.CLIConfigJSON != "" {
        json.Unmarshal([]byte(repoCfg.CLIConfigJSON), &cfg.CLIConfig)
    }
    
    // 新增：自定义 Prompt
    if repoCfg.CustomPrompt != "" {
        cfg.CustomPrompt = repoCfg.CustomPrompt
        cfg.UseCustomPrompt = repoCfg.UseCustomPrompt
    }
    
    // 新增：文件过滤
    if repoCfg.ExcludeFileTypes != "" {
        json.Unmarshal([]byte(repoCfg.ExcludeFileTypes), &cfg.ExcludeFileTypes)
    }
    if repoCfg.IgnorePatterns != "" {
        json.Unmarshal([]byte(repoCfg.IgnorePatterns), &cfg.IgnorePatterns)
    }
    
    return cfg
}
```

---

## 五、数据库迁移

**文件**：`biz/dal/db/migration.go`

**新增迁移步骤**：

```go
// 在 RunMigrations 函数中添加
func RunMigrations() {
    // ... 现有迁移 ...
    
    // 2024051101: 增强 review_repo_configs 表
    migration_2024051101_enhance_review_repo_configs()
    
    // 2024051102: 创建 review_cli_configs 表
    migration_2024051102_create_review_cli_configs()
    
    // 2024051103: 创建 review_audit_logs 表
    migration_2024051103_create_review_audit_logs()
    
    // 2024051104: 增强 webhook_rules 表
    migration_2024051104_enhance_webhook_rules()
}

func migration_2024051101_enhance_review_repo_configs() {
    // 添加新字段到 review_repo_configs 表
    db.Exec(`ALTER TABLE review_repo_configs ADD COLUMN review_mode VARCHAR(32) DEFAULT 'llm'`)
    db.Exec(`ALTER TABLE review_repo_configs ADD COLUMN cli_config_json TEXT`)
    db.Exec(`ALTER TABLE review_repo_configs ADD COLUMN custom_prompt TEXT`)
    db.Exec(`ALTER TABLE review_repo_configs ADD COLUMN use_custom_prompt BOOLEAN DEFAULT FALSE`)
    db.Exec(`ALTER TABLE review_repo_configs ADD COLUMN exclude_file_types TEXT`)
    db.Exec(`ALTER TABLE review_repo_configs ADD COLUMN ignore_patterns TEXT`)
}

func migration_2024051102_create_review_cli_configs() {
    db.Exec(`
        CREATE TABLE IF NOT EXISTS review_cli_configs (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name VARCHAR(100) NOT NULL,
            cli_type VARCHAR(50) NOT NULL,
            exec_path VARCHAR(500) NOT NULL,
            config_json TEXT,
            is_active BOOLEAN DEFAULT TRUE,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
        )
    `)
}

func migration_2024051103_create_review_audit_logs() {
    db.Exec(`
        CREATE TABLE IF NOT EXISTS review_audit_logs (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            task_id INTEGER NOT NULL,
            action VARCHAR(50) NOT NULL,
            status VARCHAR(20),
            error_message TEXT,
            duration INTEGER DEFAULT 0,
            metadata TEXT,
            created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (task_id) REFERENCES review_tasks(id)
        )
    `)
    db.Exec(`CREATE INDEX idx_review_audit_logs_task_id ON review_audit_logs(task_id)`)
    db.Exec(`CREATE INDEX idx_review_audit_logs_action ON review_audit_logs(action)`)
    db.Exec(`CREATE INDEX idx_review_audit_logs_created_at ON review_audit_logs(created_at)`)
}

func migration_2024051104_enhance_webhook_rules() {
    db.Exec(`ALTER TABLE webhook_rules ADD COLUMN priority INTEGER DEFAULT 0`)
    db.Exec(`ALTER TABLE webhook_rules ADD COLUMN description TEXT`)
}
```

---

## 六、前端 API 更新

**文件**：`frontend/src/api/modules/review.ts`

**新增内容**：

```typescript
// CLI 配置相关
export interface ReviewCLIConfigDTO {
    id: number
    name: string
    cliType: string
    execPath: string
    configJson: string
    isActive: boolean
    createdAt: string
    updatedAt: string
}

export interface CreateCLIConfigRequest {
    name: string
    cliType: string
    execPath: string
    configJson?: string
    isActive?: boolean
}

export interface UpdateCLIConfigRequest {
    name?: string
    cliType?: string
    execPath?: string
    configJson?: string
    isActive?: boolean
}

export interface TestCLIConfigResponse {
    success: boolean
    message: string
    version?: string
}

// 审查模式枚举
export type ReviewMode = 'llm' | 'claude_cli' | 'opencode_cli' | 'qoder_cli' | 'codex_cli' | 'hybrid'

// 增强 ReviewRepoConfigDTO
export interface ReviewRepoConfigDTO {
    // ... 现有字段 ...
    
    // 新增
    reviewMode: ReviewMode
    cliConfigJson?: string
    customPrompt?: string
    useCustomPrompt: boolean
    excludeFileTypes?: string
    ignorePatterns?: string
}

// API 函数
export function listCLIConfigs() {
    return request.get<unknown, ReviewCLIConfigDTO[]>('/reviews/cli-configs')
}

export function getCLIConfig(id: number) {
    return request.get<unknown, ReviewCLIConfigDTO>(`/reviews/cli-configs/${id}`)
}

export function createCLIConfig(data: CreateCLIConfigRequest) {
    return request.post<unknown, ReviewCLIConfigDTO>('/reviews/cli-configs', data)
}

export function updateCLIConfig(id: number, data: UpdateCLIConfigRequest) {
    return request.put<unknown, ReviewCLIConfigDTO>(`/reviews/cli-configs/${id}`, data)
}

export function deleteCLIConfig(id: number) {
    return request.delete(`/reviews/cli-configs/${id}`)
}

export function testCLIConfig(id: number) {
    return request.post<unknown, TestCLIConfigResponse>(`/reviews/cli-configs/${id}/test`)
}
```

---

## 七、实施步骤

### 阶段一：IDL 更新和代码生成（1-2 天）

1. 更新 `idl/biz/webhook_event.proto`
2. 更新 `idl/biz/review.proto`
3. 更新 `idl/biz/audit.proto`
4. 运行 `make hz-gen` 生成代码
5. 提交生成的代码

### 阶段二：CLI 服务实现（3-5 天）

1. 创建 `biz/service/codereview/cli_service.go`（接口定义）
2. 创建 `biz/service/codereview/cli_factory.go`（工厂）
3. 实现 Claude CLI 服务
4. 实现 OpenCode CLI 服务
5. 实现 Qoder CLI 服务
6. 实现 Codex CLI 服务

### 阶段三：审查模式扩展（2-3 天）

1. 修改 `biz/service/codereview/review_service.go`
2. 修改 `biz/service/codereview/review_service_config.go`
3. 实现 CLI 审查执行函数
4. 实现混合审查执行函数
5. 保留原有 LLM 审查模式

### 阶段四：数据库和 Handler（2-3 天）

1. 实现数据库迁移
2. 更新 Handler 层实现
3. 实现 CLI 配置管理接口
4. 实现审计日志记录

### 阶段五：前端适配（2-3 天）

1. 更新前端 API 类型
2. 创建 CLI 配置管理页面
3. 更新审查配置页面
4. 测试和调试

---

## 八、关键文件清单

### 需要修改的文件

| 文件 | 修改内容 |
|------|----------|
| `idl/biz/webhook_event.proto` | 增强事件规则定义 |
| `idl/biz/review.proto` | 增强审查配置和 CLI 配置 |
| `idl/biz/audit.proto` | 增强审计日志定义 |
| `biz/service/codereview/review_service.go` | 支持多种审查模式 |
| `biz/service/codereview/review_service_config.go` | 增强配置管理 |
| `biz/dal/db/migration.go` | 添加数据库迁移 |
| `frontend/src/api/modules/review.ts` | 更新前端 API 类型 |

### 需要新增的文件

| 文件 | 说明 |
|------|------|
| `biz/service/codereview/cli_service.go` | CLI 服务接口 |
| `biz/service/codereview/cli_factory.go` | CLI 服务工厂 |
| `biz/service/codereview/claude_cli_service.go` | Claude CLI 服务 |
| `biz/service/codereview/opencode_cli_service.go` | OpenCode CLI 服务 |
| `biz/service/codereview/qoder_cli_service.go` | Qoder CLI 服务 |
| `biz/service/codereview/codex_cli_service.go` | Codex CLI 服务 |

---

## 九、注意事项

1. **代码生成**：所有接口和数据结构通过 IDL 定义，使用 `make hz-gen` 生成
2. **PO/VO 复用**：尽量复用 IDL 定义的 message，减少手写的 PO 和 VO
3. **向后兼容**：保留原有的 LLM 审查模式，新功能通过配置启用
4. **测试覆盖**：为新增的 CLI 服务编写单元测试
5. **文档更新**：更新 AGENTS.md 和相关文档

---

## 十、参考资源

- [Code-Review-GPT-Gitlab](https://github.com/mimo-x/Code-Review-GPT-Gitlab)
- [Hz 代码生成文档](https://www.cloudwego.io/docs/hertz/tutorials/toolkit/usage/)
- [Proto3 语法](https://protobuf.dev/programming-guides/proto3/)
