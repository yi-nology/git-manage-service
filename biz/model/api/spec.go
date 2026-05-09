package api

import "time"

type SpecFile struct {
	Name     string     `json:"name"`
	Path     string     `json:"path"`
	IsDir    bool       `json:"is_dir"`
	Children []SpecFile `json:"children,omitempty"`
	Size     int64      `json:"size,omitempty"`
	ModTime  time.Time  `json:"mod_time,omitempty"`
}

type FileContent struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type LintRule struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Severity    string    `json:"severity"`
	Pattern     string    `json:"pattern"`
	Enabled     bool      `json:"enabled"`
	Priority    int       `json:"priority"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type UpdateLintRuleReq struct {
	ID          string `json:"id"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Category    string `json:"category,omitempty"`
	Severity    string `json:"severity,omitempty"`
	Pattern     string `json:"pattern,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
	Priority    *int   `json:"priority,omitempty"`
}

type CreateLintRuleReq struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Severity    string `json:"severity"`
	Pattern     string `json:"pattern"`
	Enabled     bool   `json:"enabled"`
	Priority    int    `json:"priority"`
}

type SaveSpecContentReq struct {
	RepoKey    string `json:"repo_key" form:"repo_key"`
	Path       string `json:"path" form:"path"`
	Content    string `json:"content" form:"content"`
	Message    string `json:"message" form:"message"`
	AutoCommit bool   `json:"autoCommit" form:"autoCommit"`
}

type CommitSpecReq struct {
	RepoKey string `json:"repo_key" form:"repo_key"`
	Path    string `json:"path" form:"path"`
	Message string `json:"message" form:"message"`
	Content string `json:"content" form:"content,omitempty"`
}

// SpecFileInfo spec 文件信息
type SpecFileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size,omitempty"`
	ModTime string `json:"mod_time,omitempty"`
}

// SaveSpecReq 保存 spec 文件请求
type SaveSpecReq struct {
	RepoKey       string `json:"repo_key" form:"repo_key"`
	Path          string `json:"path" form:"path"`
	Content       string `json:"content" form:"content"`
	CommitMessage string `json:"commit_message" form:"commit_message"`
}

// CreateSpecFileReq 创建新 spec 文件请求
type CreateSpecFileReq struct {
	RepoKey string `json:"repo_key" form:"repo_key"`
	Path    string `json:"path" form:"path"`
	Name    string `json:"name" form:"name"`
	Content string `json:"content" form:"content"` // 可选，如果提供则使用此内容，否则使用模板
}

// DeleteSpecFileReq 删除 spec 文件请求
type DeleteSpecFileReq struct {
	RepoKey       string `json:"repo_key" form:"repo_key"`
	Path          string `json:"path" form:"path"`
	CommitMessage string `json:"commit_message" form:"commit_message"`
}

type SaveSpecResponse struct {
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

type CommitResponse struct {
	Message string `json:"message"`
	Output  string `json:"output,omitempty"`
}

type SaveWithValidationResponse struct {
	Message          string      `json:"message"`
	ValidationResult interface{} `json:"validationResult,omitempty"`
}

type CreateFileResponse struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type AIAssistResponse struct {
	Result       string `json:"result"`
	ApplyContent string `json:"applyContent,omitempty"`
}

type AIFixResponse struct {
	Content string `json:"content"`
}

type SpecContentResponse struct {
	Content string `json:"content"`
	Path    string `json:"path"`
}

type FormatChangeDTO struct {
	Line   int    `json:"line"`
	Type   string `json:"type"`
	Before string `json:"before"`
	After  string `json:"after"`
	Reason string `json:"reason"`
}

type FormatResponse struct {
	Content string            `json:"content"`
	Changes []FormatChangeDTO `json:"changes"`
}
