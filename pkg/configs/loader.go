package configs

import (
	"log"

	"github.com/spf13/viper"
)

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPaths []string, configName string, configType string) (Config, error) {
	v := viper.New()
	v.SetConfigName(configName)
	v.SetConfigType(configType)

	for _, path := range configPaths {
		v.AddConfigPath(path)
	}

	// Set defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("rpc.port", 8888)
	v.SetDefault("database.type", "sqlite")
	v.SetDefault("database.path", "git_sync.db")
	v.SetDefault("webhook.secret", "my-secret-key")
	v.SetDefault("webhook.rate_limit", 100)
	v.SetDefault("webhook.ip_whitelist", []string{})

	// Environment variables override
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("Config file not found, using defaults")
		} else {
			return Config{}, err
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return Config{}, err
	}

	return config, nil
}
