package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/nickkcj/orbit-backend/internal/database"
)

const (
	UserContextKey   = "user"
	TenantContextKey = "tenant"
)

// GetUserFromContext retrieves the authenticated user from the request context
func GetUserFromContext(c echo.Context) *database.User {
	user, ok := c.Get(UserContextKey).(*database.User)
	if !ok {
		return nil
	}
	return user
}

// GetTenantFromContext retrieves the current tenant from the request context
func GetTenantFromContext(c echo.Context) *database.Tenant {
	tenant, ok := c.Get(TenantContextKey).(*database.Tenant)
	if !ok {
		return nil
	}
	return tenant
}
