package config

import "os"

var (
	WebhookSecret      = "my-secret-key" // Default secret
	WebhookRateLimit   = 100             // Requests per minute
	WebhookIPWhitelist = []string{}      // Empty means no whitelist
)

func Init() {
	if secret := os.Getenv("WEBHOOK_SECRET"); secret != "" {
		WebhookSecret = secret
	}
	// Further config loading can be added here
}
