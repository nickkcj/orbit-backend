package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"

	"github.com/nickkcj/orbit-backend/internal/config"
	"github.com/nickkcj/orbit-backend/internal/database"
	"github.com/nickkcj/orbit-backend/internal/handler"
	"github.com/nickkcj/orbit-backend/internal/middleware"
	"github.com/nickkcj/orbit-backend/internal/service"
)

func main() {
	// Load configuration
	cfg := config.Load()

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Database connection
	conn, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Connected to database")

	// Initialize layers
	db := database.New(conn)

	// Storage config (R2)
	storageConfig := &service.StorageConfig{
		AccountID:       cfg.R2AccountID,
		AccessKeyID:     cfg.R2AccessKeyID,
		SecretAccessKey: cfg.R2SecretAccessKey,
		BucketName:      cfg.R2BucketName,
	}

	services := service.New(db, cfg.JWTSecret, storageConfig)
	handlers := handler.New(services)
	authMiddleware := middleware.NewAuthMiddleware(services.Auth)
	tenantMiddleware := middleware.NewTenantMiddleware(services.Tenant, cfg.BaseDomain)

	// Echo instance
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())

	// Register routes
	handlers.RegisterRoutes(e, authMiddleware, tenantMiddleware)

	// Start server
	log.Printf("Server starting on port %s (base domain: %s)", cfg.Port, cfg.BaseDomain)
	if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
		log.Fatal("Failed to start server:", err)
	}
}
