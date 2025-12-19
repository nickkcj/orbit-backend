package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/nickkcj/orbit-backend/internal/middleware"
)

func (h *Handler) RegisterRoutes(e *echo.Echo, authMiddleware *middleware.AuthMiddleware) {
	// Health
	e.GET("/health", h.Health)

	// API v1
	v1 := e.Group("/api/v1")

	// Auth (public)
	v1.POST("/auth/register", h.Register)
	v1.POST("/auth/login", h.Login)

	// Protected routes
	protected := v1.Group("", authMiddleware.RequireAuth)

	// Me
	protected.GET("/auth/me", h.Me)

	// Tenants (public list, protected create)
	v1.GET("/tenants", h.ListTenants)
	v1.GET("/tenants/:slug", h.GetTenantBySlug)
	protected.POST("/tenants", h.CreateTenant)

	// Users
	protected.POST("/users", h.CreateUser)
	protected.GET("/users", h.GetUserByEmail)
	protected.GET("/users/:userId/tenants", h.ListUserTenants)

	// Categories (scoped by tenant)
	v1.GET("/tenants/:tenantId/categories", h.ListCategories)
	protected.POST("/categories", h.CreateCategory)
	protected.GET("/categories/:id", h.GetCategory)
	protected.PUT("/categories/:id", h.UpdateCategory)
	protected.DELETE("/categories/:id", h.DeleteCategory)

	// Posts (public read, protected write)
	v1.GET("/tenants/:tenantId/posts", h.ListPosts)
	v1.GET("/posts/:id", h.GetPost)
	protected.POST("/posts", h.CreatePost)
	protected.PUT("/posts/:id", h.UpdatePost)
	protected.POST("/posts/:id/publish", h.PublishPost)
	protected.DELETE("/posts/:id", h.DeletePost)

	// Comments (public read, protected write)
	v1.GET("/posts/:postId/comments", h.ListComments)
	v1.GET("/comments/:id", h.GetComment)
	v1.GET("/comments/:id/replies", h.ListReplies)
	protected.POST("/comments", h.CreateComment)
	protected.PUT("/comments/:id", h.UpdateComment)
	protected.DELETE("/comments/:id", h.DeleteComment)

	// Members (scoped by tenant)
	v1.GET("/tenants/:tenantId/members", h.ListMembers)
	protected.POST("/tenants/:tenantId/members", h.AddMember)
	protected.GET("/tenants/:tenantId/members/:userId", h.GetMember)
	protected.PUT("/tenants/:tenantId/members/:userId/role", h.UpdateMemberRole)
	protected.DELETE("/tenants/:tenantId/members/:userId", h.RemoveMember)
}
