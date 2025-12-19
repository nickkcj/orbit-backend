package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/nickkcj/orbit-backend/internal/database"
)

const UserContextKey = "user"

// GetUserFromContext retrieves the authenticated user from the request context
func GetUserFromContext(c echo.Context) *database.User {
	user, ok := c.Get(UserContextKey).(*database.User)
	if !ok {
		return nil
	}
	return user
}
