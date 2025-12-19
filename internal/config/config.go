package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
	Environment string
	JWTSecret   string
	BaseDomain  string

	// R2/S3 Storage
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string
}

func Load() *Config {
	// Load .env file (ignore error in production where env vars are set directly)
	_ = godotenv.Load()

	return &Config{
		DatabaseURL: getEnv("DATABASE_URL", ""),
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", "development"),
		JWTSecret:   getEnv("JWT_SECRET", "change-me-in-production"),
		BaseDomain:  getEnv("BASE_DOMAIN", "orbit.app.br"),

		// R2/S3 Storage
		R2AccountID:       getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:      getEnv("R2_BUCKET_NAME", "orbit-videos"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
