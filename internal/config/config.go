package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	Port        string
	Environment string
	JWTSecret   string
	BaseDomain  string
	FrontendURL string

	// Google OAuth
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	// R2/S3 Storage
	R2AccountID       string
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string

	// Redis
	RedisURL      string
	RedisPassword string
	RedisDB       int

	// Worker
	WorkerConcurrency int
	ShutdownTimeout   time.Duration

	// WebSocket
	WSPingInterval time.Duration
	WSWriteTimeout time.Duration

	// Cloudflare Stream
	CloudflareAccountID         string
	CloudflareStreamAPIToken    string
	CloudflareStreamSigningKey  string
	CloudflareStreamWebhookSecret string
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
		FrontendURL: getEnv("FRONTEND_URL", "http://localhost:3000"),

		// Google OAuth
		GoogleClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  getEnv("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/google/callback"),

		// R2/S3 Storage
		R2AccountID:       getEnv("R2_ACCOUNT_ID", ""),
		R2AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:      getEnv("R2_BUCKET_NAME", "orbit-videos"),

		// Redis
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvInt("REDIS_DB", 0),

		// Worker
		WorkerConcurrency: getEnvInt("WORKER_CONCURRENCY", 10),
		ShutdownTimeout:   getEnvDuration("SHUTDOWN_TIMEOUT", 30*time.Second),

		// WebSocket
		WSPingInterval: getEnvDuration("WS_PING_INTERVAL", 30*time.Second),
		WSWriteTimeout: getEnvDuration("WS_WRITE_TIMEOUT", 10*time.Second),

		// Cloudflare Stream
		CloudflareAccountID:         getEnv("CLOUDFLARE_ACCOUNT_ID", ""),
		CloudflareStreamAPIToken:    getEnv("CLOUDFLARE_STREAM_API_TOKEN", ""),
		CloudflareStreamSigningKey:  getEnv("CLOUDFLARE_STREAM_SIGNING_KEY", ""),
		CloudflareStreamWebhookSecret: getEnv("CLOUDFLARE_STREAM_WEBHOOK_SECRET", ""),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}
