package configs

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Webhook    WebhookConfig    `mapstructure:"webhook"`
	Rpc        RpcConfig        `mapstructure:"rpc"`
	Storage    StorageConfig    `mapstructure:"storage"`
	Lock       LockConfig       `mapstructure:"lock"`
	Lint       LintConfig       `mapstructure:"lint"`
	CodeReview CodeReviewConfig `mapstructure:"code_review"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Prometheus PrometheusConfig `mapstructure:"prometheus"`
	Mirror     MirrorConfig     `mapstructure:"mirror"`
}

type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	ExternalURL string `mapstructure:"external_url"`
}

type RpcConfig struct {
	Port int `mapstructure:"port"`
}

type DatabaseConfig struct {
	Type     string `mapstructure:"type"` // sqlite, mysql, postgres
	DSN      string `mapstructure:"dsn"`  // Data Source Name (for mysql/postgres)
	Path     string `mapstructure:"path"` // For SQLite
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type WebhookConfig struct {
	Secret      string   `mapstructure:"secret"`
	RateLimit   int      `mapstructure:"rate_limit"`
	IPWhitelist []string `mapstructure:"ip_whitelist"`
}

// StorageConfig 对象存储配置
type StorageConfig struct {
	Type           string `mapstructure:"type"`             // "local" | "minio"
	LocalPath      string `mapstructure:"local_path"`       // 本地存储根目录
	Endpoint       string `mapstructure:"endpoint"`         // MinIO 端点
	AccessKey      string `mapstructure:"access_key"`       // MinIO Access Key
	SecretKey      string `mapstructure:"secret_key"`       // MinIO Secret Key
	UseSSL         bool   `mapstructure:"use_ssl"`          // 是否使用 SSL
	RepoBucket     string `mapstructure:"repo_bucket"`      // 仓库存储桶名称
	SSHKeyBucket   string `mapstructure:"ssh_key_bucket"`   // SSH密钥存储桶名称
	AuditLogBucket string `mapstructure:"audit_log_bucket"` // 审计日志存储桶名称
	BackupBucket   string `mapstructure:"backup_bucket"`    // 备份存储桶名称
}

// LockConfig 分布式锁配置
type LockConfig struct {
	Type          string `mapstructure:"type"`           // "memory" | "redis"
	RedisAddr     string `mapstructure:"redis_addr"`     // Redis 地址
	RedisPassword string `mapstructure:"redis_password"` // Redis 密码
	RedisDB       int    `mapstructure:"redis_db"`       // Redis 数据库号
}

type LintConfig struct {
	EnableRpmlint bool `mapstructure:"enable_rpmlint"`
}

type CodeReviewConfig struct {
	Enabled        bool                `mapstructure:"enabled"`
	DefaultLLM     string              `mapstructure:"default_llm"`
	MaxFiles       int                 `mapstructure:"max_files"`
	MaxDiffLines   int                 `mapstructure:"max_diff_lines"`
	AutoReviewOnMR bool                `mapstructure:"auto_review_on_mr"`
	BlockOnHigh    bool                `mapstructure:"block_on_high"`
	LLMProviders   []LLMProviderConfig `mapstructure:"llm_providers"`
	RAG            RAGConfig           `mapstructure:"rag"`
}

type RAGConfig struct {
	Enabled         bool    `mapstructure:"enabled"`
	EmbeddingModel  string  `mapstructure:"embedding_model"`
	MaxChunks       int     `mapstructure:"max_chunks"`
	TopK            int     `mapstructure:"top_k"`
	MinScore        float64 `mapstructure:"min_score"`
	MaxContextChars int     `mapstructure:"max_context_chars"`
}

type LLMProviderConfig struct {
	Name      string `mapstructure:"name"`
	Type      string `mapstructure:"type"`
	BaseURL   string `mapstructure:"base_url"`
	APIKey    string `mapstructure:"api_key"`
	Model     string `mapstructure:"model"`
	MaxTokens int    `mapstructure:"max_tokens"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type PrometheusConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Path    string `mapstructure:"path"`
}

type MirrorConfig struct {
	GitBackend          string `mapstructure:"git_backend"`
	QueueBackend        string `mapstructure:"queue_backend"`
	MaxWorkers          int    `mapstructure:"max_workers"`
	ScanInterval        int    `mapstructure:"scan_interval"`
	MaxRetry            int    `mapstructure:"max_retry"`
	LogRetentionDays    int    `mapstructure:"log_retention_days"`
	DefaultSyncInterval int    `mapstructure:"default_sync_interval"`
}
