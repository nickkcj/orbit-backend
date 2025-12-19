package handler

import (
	"github.com/labstack/echo/v4"
)

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	// Health
	e.GET("/health", h.Health)

	// API v1
	v1 := e.Group("/api/v1")

	// Tenants
	v1.GET("/tenants", h.ListTenants)
	v1.GET("/tenants/:slug", h.GetTenantBySlug)
	v1.POST("/tenants", h.CreateTenant)

	// Users
	v1.POST("/users", h.CreateUser)
	v1.GET("/users", h.GetUserByEmail)
}
