package configs

import (
	"log"
	"os"
	"sync"
)

var (
	GlobalConfig Config

	// configMu guards runtime mutations to GlobalConfig.CodeReview (the only
	// sub-config written by HTTP handlers at runtime). The rest of GlobalConfig
	// is set once during Init() and never mutated.
	configMu sync.RWMutex

	// Keep backward compatibility
	WebhookSecret      = "my-secret-key"
	WebhookRateLimit   = 100
	WebhookIPWhitelist = []string{}
	DebugMode          = false
)

func Init() {
	configPaths := []string{".", "./conf", "../conf"}

	config, err := LoadConfig(configPaths, "config", "yaml")
	if err != nil {
		log.Fatalf("Fatal error loading config: %s \n", err)
	}

	GlobalConfig = config

	// Update global variables for backward compatibility
	WebhookSecret = GlobalConfig.Webhook.Secret
	WebhookRateLimit = GlobalConfig.Webhook.RateLimit
	WebhookIPWhitelist = GlobalConfig.Webhook.IPWhitelist

	// Manual override for old ENV vars
	if secret := os.Getenv("WEBHOOK_SECRET"); secret != "" {
		WebhookSecret = secret
		GlobalConfig.Webhook.Secret = secret
	}

	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		GlobalConfig.Database.Path = dbPath
	}
}

// GetCodeReviewConfig returns a copy of the CodeReview config under a read lock.
// Use this instead of direct GlobalConfig.CodeReview access from goroutines.
func GetCodeReviewConfig() CodeReviewConfig {
	configMu.RLock()
	defer configMu.RUnlock()
	return GlobalConfig.CodeReview
}

// UpdateCodeReviewConfig mutates the CodeReview config under a write lock.
func UpdateCodeReviewConfig(fn func(*CodeReviewConfig)) {
	configMu.Lock()
	defer configMu.Unlock()
	fn(&GlobalConfig.CodeReview)
}
