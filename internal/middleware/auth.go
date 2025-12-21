package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/nickkcj/orbit-backend/internal/database"
	"github.com/nickkcj/orbit-backend/internal/service"
)

const (
	UserContextKey = "user"
)

type AuthMiddleware struct {
	authService *service.AuthService
}

func NewAuthMiddleware(authService *service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{authService: authService}
}

// RequireAuth middleware requires a valid JWT token
func (m *AuthMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := extractToken(c)
		if token == "" {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "missing authorization token",
			})
		}

		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "invalid or expired token",
			})
		}

		// Get user from database
		user, err := m.authService.GetUserByID(c.Request().Context(), claims.UserID)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "user not found",
			})
		}

		// Check user status
		if user.Status != "active" {
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "user account is not active",
			})
		}

		// Store user in context
		c.Set(UserContextKey, &user)

		return next(c)
	}
}

// OptionalAuth middleware extracts user if token is present but doesn't require it
func (m *AuthMiddleware) OptionalAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token := extractToken(c)
		if token == "" {
			return next(c)
		}

		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			return next(c)
		}

		user, err := m.authService.GetUserByID(c.Request().Context(), claims.UserID)
		if err == nil && user.Status == "active" {
			c.Set(UserContextKey, &user)
		}

		return next(c)
	}
}

func extractToken(c echo.Context) string {
	// Try Authorization header first
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// Try cookie as fallback (for cross-subdomain auth)
	cookie, err := c.Cookie("auth_token")
	if err == nil && cookie.Value != "" {
		return cookie.Value
	}

	// Try query param as last resort
	return c.QueryParam("token")
}

// Helper to get user from context (use in handlers)
func GetUserFromContext(c echo.Context) *database.User {
	user, ok := c.Get(UserContextKey).(*database.User)
	if !ok {
		return nil
	}
	return user
}
