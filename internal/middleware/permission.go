package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nickkcj/orbit-backend/internal/service"
)

// PermissionMiddleware handles permission-based access control
type PermissionMiddleware struct {
	permissionService *service.PermissionService
}

// NewPermissionMiddleware creates a new permission middleware
func NewPermissionMiddleware(permissionService *service.PermissionService) *PermissionMiddleware {
	return &PermissionMiddleware{
		permissionService: permissionService,
	}
}

// RequirePermission creates middleware that checks for a specific permission
func (m *PermissionMiddleware) RequirePermission(permissionCode string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := GetUserFromContext(c)
			tenant := GetTenantFromContext(c)

			if user == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "authentication required",
				})
			}

			if tenant == nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "tenant context required",
				})
			}

			hasPermission, err := m.permissionService.HasPermission(
				c.Request().Context(),
				tenant.ID,
				user.ID,
				permissionCode,
			)

			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "failed to check permissions",
				})
			}

			if !hasPermission {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "permission denied: " + permissionCode,
				})
			}

			return next(c)
		}
	}
}

// RequireAnyPermission creates middleware that checks if user has at least one of the permissions
func (m *PermissionMiddleware) RequireAnyPermission(codes ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := GetUserFromContext(c)
			tenant := GetTenantFromContext(c)

			if user == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "authentication required",
				})
			}

			if tenant == nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "tenant context required",
				})
			}

			hasAny, err := m.permissionService.HasAnyPermission(
				c.Request().Context(),
				tenant.ID,
				user.ID,
				codes...,
			)

			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "failed to check permissions",
				})
			}

			if !hasAny {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "permission denied",
				})
			}

			return next(c)
		}
	}
}

// RequireAllPermissions creates middleware that checks if user has all specified permissions
func (m *PermissionMiddleware) RequireAllPermissions(codes ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := GetUserFromContext(c)
			tenant := GetTenantFromContext(c)

			if user == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "authentication required",
				})
			}

			if tenant == nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "tenant context required",
				})
			}

			hasAll, err := m.permissionService.HasAllPermissions(
				c.Request().Context(),
				tenant.ID,
				user.ID,
				codes...,
			)

			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "failed to check permissions",
				})
			}

			if !hasAll {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "permission denied",
				})
			}

			return next(c)
		}
	}
}

// RequireOwnership creates middleware that checks permission OR resource ownership
// permissionCode: the permission required if not owner
// ownPermissionCode: the permission for own resources
// getOwnerID: function to extract owner ID from request
func (m *PermissionMiddleware) RequireOwnership(
	permissionCode string,
	ownPermissionCode string,
	getOwnerID func(c echo.Context) (uuid.UUID, error),
) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := GetUserFromContext(c)
			tenant := GetTenantFromContext(c)

			if user == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "authentication required",
				})
			}

			if tenant == nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "tenant context required",
				})
			}

			// Get owner ID from request
			ownerID, err := getOwnerID(c)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "failed to determine resource owner",
				})
			}

			// Check if user is the owner
			isOwner := user.ID == ownerID

			// Determine which permission to check
			requiredPermission := permissionCode
			if isOwner {
				requiredPermission = ownPermissionCode
			}

			hasPermission, err := m.permissionService.HasPermission(
				c.Request().Context(),
				tenant.ID,
				user.ID,
				requiredPermission,
			)

			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "failed to check permissions",
				})
			}

			if !hasPermission {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "permission denied",
				})
			}

			return next(c)
		}
	}
}

// RequireOwnerOrAdmin creates middleware that requires owner or admin role
func (m *PermissionMiddleware) RequireOwnerOrAdmin() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			user := GetUserFromContext(c)
			tenant := GetTenantFromContext(c)

			if user == nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "authentication required",
				})
			}

			if tenant == nil {
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "tenant context required",
				})
			}

			isOwnerOrAdmin, err := m.permissionService.IsOwnerOrAdmin(
				c.Request().Context(),
				tenant.ID,
				user.ID,
			)

			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "failed to check role",
				})
			}

			if !isOwnerOrAdmin {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "owner or admin role required",
				})
			}

			return next(c)
		}
	}
}
