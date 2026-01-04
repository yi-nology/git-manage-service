package configs

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Webhook  WebhookConfig  `mapstructure:"webhook"`
	Rpc      RpcConfig      `mapstructure:"rpc"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
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
