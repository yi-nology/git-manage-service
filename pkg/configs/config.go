package configs

import (
	"log"
	"os"
)

var (
	GlobalConfig Config

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
