package adapter

import (
	"os"
	"path/filepath"
)

// Config git-sync-service 配置（内部定义）
type Config struct {
	Database DatabaseConfig
	Redis    RedisConfig
	Git      GitConfig
	Sync     SyncConfig
}

type DatabaseConfig struct {
	Driver  string
	DSN     string
	MaxIdle int
	MaxOpen int
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type GitConfig struct {
	TempDir string
}

type SyncConfig struct {
	MaxConcurrent  int
	DefaultTimeout int
	RetryCount     int
}

// DefaultConfig 获取默认配置
func DefaultConfig() *Config {
	tempDir := getTempDir()

	return &Config{
		Database: DatabaseConfig{
			Driver:  "sqlite",
			DSN:     "data/git_sync.db",
			MaxIdle: 10,
			MaxOpen: 100,
		},
		Redis: RedisConfig{
			Addr:     "127.0.0.1:6379",
			Password: "",
			DB:       0,
		},
		Git: GitConfig{
			TempDir: tempDir,
		},
		Sync: SyncConfig{
			MaxConcurrent:  5,
			DefaultTimeout: 300,
			RetryCount:     3,
		},
	}
}

// getTempDir 获取 Git 临时工作目录
func getTempDir() string {
	baseTemp := os.Getenv("TEMP")
	if baseTemp == "" {
		baseTemp = os.TempDir()
	}
	return filepath.Join(baseTemp, "git-sync-service")
}

// CredentialAdapter 凭证适配器，将现有凭证系统转换为 git-sync-service 使用的格式
type CredentialAdapter struct {
	// 后续扩展：将 SSHKey、Credential 等转换为 git-sync-service 可用格式
}

// NewCredentialAdapter 创建凭证适配器
func NewCredentialAdapter() *CredentialAdapter {
	return &CredentialAdapter{}
}
