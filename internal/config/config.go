package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabasePath       string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	AppBaseURL         string
	JWTSecret          string
	OAuthStateSecret   string
}

func LoadConfig() *Config {
	_ = godotenv.Load()
	return &Config{
		DatabasePath:       getEnvWithDefault("DB_PATH", "app.db"),
		GoogleClientID:     getEnvWithDefault("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnvWithDefault("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnvWithDefault("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/auth/google/callback"),
		AppBaseURL:         getEnvWithDefault("APP_BASE_URL", "http://localhost:8080"),
		JWTSecret:          getEnvWithDefault("JWT_SECRET", "your-secret-key"),
		OAuthStateSecret:   getEnvWithDefault("OAUTH_STATE_SECRET", "changemechangemechangemechangeme"),
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return defaultValue
}
