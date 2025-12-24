package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"

	"github.com/nickkcj/orbit-backend/internal/cache"
	"github.com/nickkcj/orbit-backend/internal/config"
	"github.com/nickkcj/orbit-backend/internal/database"
	"github.com/nickkcj/orbit-backend/internal/handler"
	"github.com/nickkcj/orbit-backend/internal/middleware"
	"github.com/nickkcj/orbit-backend/internal/service"
	"github.com/nickkcj/orbit-backend/internal/websocket"
	"github.com/nickkcj/orbit-backend/internal/worker"
)

func main() {
	// Load configuration
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Database connection with connection pool settings
	conn, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Configure connection pool
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	if err := conn.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("Connected to database")

	// Initialize Redis connection for Asynq
	redisOpt := asynq.RedisClientOpt{
		Addr:     parseRedisAddr(cfg.RedisURL),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	// Initialize layers
	db := database.New(conn)

	// Initialize Redis cache
	var redisCache cache.Cache
	redisAddr := parseRedisAddr(cfg.RedisURL)
	redisCacheInstance, err := cache.NewRedisCache(redisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Printf("Warning: Failed to initialize Redis cache: %v (caching disabled)", err)
	} else {
		redisCache = redisCacheInstance
		log.Println("Redis cache initialized")
	}

	// Storage config (R2)
	storageConfig := &service.StorageConfig{
		AccountID:       cfg.R2AccountID,
		AccessKeyID:     cfg.R2AccessKeyID,
		SecretAccessKey: cfg.R2SecretAccessKey,
		BucketName:      cfg.R2BucketName,
	}

	// Stream config (Cloudflare Stream)
	streamConfig := &service.StreamConfig{
		AccountID:     cfg.CloudflareAccountID,
		APIToken:      cfg.CloudflareStreamAPIToken,
		SigningKey:    cfg.CloudflareStreamSigningKey,
		WebhookSecret: cfg.CloudflareStreamWebhookSecret,
	}

	// Google OAuth config
	var googleConfig *service.GoogleOAuthConfig
	if cfg.GoogleClientID != "" && cfg.GoogleClientSecret != "" {
		googleConfig = &service.GoogleOAuthConfig{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			RedirectURL:  cfg.GoogleRedirectURL,
			FrontendURL:  cfg.FrontendURL,
		}
		log.Println("Google OAuth configured")
	}

	services := service.New(db, cfg.JWTSecret, storageConfig, streamConfig, googleConfig, redisCache)

	// Initialize task client
	taskClient := worker.NewTaskClient(redisOpt)

	// Initialize handlers with task client
	handlers := handler.New(services, taskClient)

	authMiddleware := middleware.NewAuthMiddleware(services.Auth)
	tenantMiddleware := middleware.NewTenantMiddleware(services.Tenant, cfg.BaseDomain)
	permissionMiddleware := middleware.NewPermissionMiddleware(services.Permission)

	// Initialize WebSocket hub
	wsHub := websocket.NewHub()
	wsAuthenticator := websocket.NewAuthenticator(services.Auth, services.Tenant)
	wsHandler := websocket.NewHandler(wsHub, wsAuthenticator)

	// Initialize worker
	workerServer := worker.NewWorker(redisOpt, cfg.WorkerConcurrency, services)

	// Echo instance
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())

	// Register routes
	handlers.RegisterRoutes(e, authMiddleware, tenantMiddleware, permissionMiddleware, wsHandler)

	// Start WebSocket hub in background
	wsCtx, wsCancel := context.WithCancel(context.Background())
	go wsHub.Run(wsCtx)

	// Start worker in background
	go func() {
		if err := workerServer.Start(); err != nil {
			log.Printf("Worker error: %v", err)
		}
	}()

	// Start HTTP server in background
	go func() {
		log.Printf("Server starting on port %s (base domain: %s)", cfg.Port, cfg.BaseDomain)
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	// Shutdown sequence (order matters)

	// 1. Stop accepting new HTTP requests
	log.Println("Stopping HTTP server...")
	if err := e.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// 2. Stop WebSocket hub
	log.Println("Stopping WebSocket hub...")
	wsCancel()

	// 3. Stop worker (waits for in-flight tasks)
	log.Println("Stopping background worker...")
	workerServer.Shutdown()

	// 4. Close task client
	log.Println("Closing task client...")
	if err := taskClient.Close(); err != nil {
		log.Printf("Task client close error: %v", err)
	}

	// 5. Close Redis cache
	if redisCache != nil {
		log.Println("Closing Redis cache...")
		if err := redisCache.Close(); err != nil {
			log.Printf("Redis cache close error: %v", err)
		}
	}

	// 6. Close database connection
	log.Println("Closing database connection...")
	if err := conn.Close(); err != nil {
		log.Printf("Database close error: %v", err)
	}

	log.Println("Shutdown complete")
}

// parseRedisAddr extracts the host:port from a Redis URL
func parseRedisAddr(redisURL string) string {
	// Handle redis://host:port format
	if len(redisURL) > 8 && redisURL[:8] == "redis://" {
		return redisURL[8:]
	}
	// Handle rediss://host:port format (TLS)
	if len(redisURL) > 9 && redisURL[:9] == "rediss://" {
		return redisURL[9:]
	}
	return redisURL
}
