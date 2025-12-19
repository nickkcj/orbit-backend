package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/nickkcj/orbit-backend/internal/database"
	"github.com/nickkcj/orbit-backend/internal/service"
)

const (
	TenantContextKey = "tenant"
)

type TenantMiddleware struct {
	tenantService *service.TenantService
	baseDomain    string
}

func NewTenantMiddleware(tenantService *service.TenantService, baseDomain string) *TenantMiddleware {
	return &TenantMiddleware{
		tenantService: tenantService,
		baseDomain:    baseDomain,
	}
}

// RequireTenant middleware extracts subdomain, looks up tenant, and requires it to exist
func (m *TenantMiddleware) RequireTenant(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		slug := m.extractSubdomain(c.Request().Host)

		// No subdomain = main domain request (not tenant-scoped)
		if slug == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "tenant subdomain required",
				"code":  "TENANT_REQUIRED",
			})
		}

		// Look up tenant by slug
		tenant, err := m.tenantService.GetBySlug(c.Request().Context(), slug)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "comunidade nao encontrada",
				"code":  "TENANT_NOT_FOUND",
			})
		}

		// Check tenant status
		if tenant.Status != "active" {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "comunidade nao encontrada",
				"code":  "TENANT_INACTIVE",
			})
		}

		// Store tenant in context
		c.Set(TenantContextKey, &tenant)

		return next(c)
	}
}

// OptionalTenant extracts tenant if subdomain present, but doesn't require it
func (m *TenantMiddleware) OptionalTenant(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		slug := m.extractSubdomain(c.Request().Host)

		if slug == "" {
			return next(c)
		}

		tenant, err := m.tenantService.GetBySlug(c.Request().Context(), slug)
		if err == nil && tenant.Status == "active" {
			c.Set(TenantContextKey, &tenant)
		}

		return next(c)
	}
}

// extractSubdomain parses the Host header and extracts subdomain
// Examples:
//   - joao.orbit.app.br -> "joao"
//   - orbit.app.br -> ""
//   - www.orbit.app.br -> "" (www is treated as main domain)
//   - localhost:8080 -> "" (development)
func (m *TenantMiddleware) extractSubdomain(host string) string {
	// Remove port if present
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Handle localhost for development
	if host == "localhost" || host == "127.0.0.1" {
		return ""
	}

	// Check if host ends with base domain
	if !strings.HasSuffix(host, m.baseDomain) {
		return ""
	}

	// Extract subdomain
	subdomain := strings.TrimSuffix(host, "."+m.baseDomain)

	// If subdomain equals host, there was no subdomain
	if subdomain == host || subdomain == "" {
		return ""
	}

	// Treat "www" as main domain
	if subdomain == "www" {
		return ""
	}

	return subdomain
}

// GetTenantFromContext retrieves the tenant from the request context
func GetTenantFromContext(c echo.Context) *database.Tenant {
	tenant, ok := c.Get(TenantContextKey).(*database.Tenant)
	if !ok {
		return nil
	}
	return tenant
}
