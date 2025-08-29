package config

import (
	"os"
)

type Config struct {
	DatabasePath string
}

func LoadConfig() *Config {
	return &Config{
		DatabasePath: getEnvWithDefault("DB_PATH", "app.db"),
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
