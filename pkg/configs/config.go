package configs

import (
	"fmt"
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

	if secret := os.Getenv("WEBHOOK_SECRET"); secret != "" {
		WebhookSecret = secret
		GlobalConfig.Webhook.Secret = secret
	}

	if dbType := os.Getenv("DB_TYPE"); dbType != "" {
		GlobalConfig.Database.Type = dbType
	}
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		GlobalConfig.Database.Path = dbPath
	}
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		GlobalConfig.Database.Host = dbHost
	}
	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		port := 0
		fmt.Sscanf(dbPort, "%d", &port)
		if port > 0 {
			GlobalConfig.Database.Port = port
		}
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		GlobalConfig.Database.User = dbUser
	}
	if dbPassword := os.Getenv("CONFIG_DATABASE_PASSWORD"); dbPassword != "" {
		GlobalConfig.Database.Password = dbPassword
	} else if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		GlobalConfig.Database.Password = dbPassword
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		GlobalConfig.Database.DBName = dbName
	}
}
