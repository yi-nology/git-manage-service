package constants

// 时间格式常量
const (
	TimeFormatFull = "2006-01-02 15:04:05"
	TimeFormatDate = "2006-01-02"
	TimeFormatTime = "15:04:05"
)

// 分页常量
const (
	DefaultPageSize = 100
	MaxPageSize     = 1000
	DefaultPage     = 1
)

// Git 默认值
const (
	DefaultRemoteName = "origin"
	DefaultBranch     = "main"
	DefaultVersion    = "v0.0.0"
)

// 认证类型常量
const (
	AuthTypeSSH      = "ssh"
	AuthTypePassword = "password"
	AuthTypeToken    = "token"
)

// 分支类型常量
const (
	BranchTypeLocal  = "local"
	BranchTypeRemote = "remote"
	BranchTypeAll    = "all"
)

// Context 键常量
const (
	ContextKeyRepo      = "repo"
	ContextKeyRequestID = "request_id"
	ContextKeyUserID    = "user_id"
)

// HTTP Header 常量
const (
	HeaderRequestID   = "X-Request-ID"
	HeaderContentType = "Content-Type"
)
