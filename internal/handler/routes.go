package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/nickkcj/orbit-backend/internal/middleware"
)

func (h *Handler) RegisterRoutes(
	e *echo.Echo,
	authMiddleware *middleware.AuthMiddleware,
	tenantMiddleware *middleware.TenantMiddleware,
) {
	// Health
	e.GET("/health", h.Health)

	// Webhooks (public endpoints for external services)
	webhooks := e.Group("/webhooks")
	webhooks.POST("/r2", h.HandleR2Webhook)

	// API v1
	v1 := e.Group("/api/v1")

	// ============================================
	// GLOBAL ROUTES (no tenant subdomain required)
	// ============================================

	// Auth (public)
	v1.POST("/auth/register", h.Register)
	v1.POST("/auth/login", h.Login)

	// Auth (protected)
	v1.GET("/auth/me", h.Me, authMiddleware.RequireAuth)

	// Tenant management (for main domain operations)
	v1.GET("/tenants", h.ListTenants)
	v1.GET("/tenants/:slug", h.GetTenantBySlug)
	v1.POST("/tenants", h.CreateTenant, authMiddleware.RequireAuth)

	// User's tenants (protected)
	v1.POST("/users", h.CreateUser, authMiddleware.RequireAuth)
	v1.GET("/users", h.GetUserByEmail, authMiddleware.RequireAuth)
	v1.GET("/users/:userId/tenants", h.ListUserTenants, authMiddleware.RequireAuth)

	// ============================================
	// TENANT-SCOPED ROUTES (subdomain required)
	// ============================================

	// Tenant-scoped group - all routes require valid tenant subdomain
	tenantScoped := v1.Group("", tenantMiddleware.RequireTenant)
	tenantProtected := tenantScoped.Group("", authMiddleware.RequireAuth)

	// Categories (tenant-scoped)
	tenantScoped.GET("/categories", h.ListCategories)
	tenantProtected.POST("/categories", h.CreateCategory)
	tenantProtected.GET("/categories/:id", h.GetCategory)
	tenantProtected.PUT("/categories/:id", h.UpdateCategory)
	tenantProtected.DELETE("/categories/:id", h.DeleteCategory)

	// Posts (tenant-scoped)
	tenantScoped.GET("/posts", h.ListPosts)
	tenantScoped.GET("/posts/:id", h.GetPost)
	tenantProtected.POST("/posts", h.CreatePost)
	tenantProtected.PUT("/posts/:id", h.UpdatePost)
	tenantProtected.POST("/posts/:id/publish", h.PublishPost)
	tenantProtected.DELETE("/posts/:id", h.DeletePost)

	// Comments (tenant-scoped)
	tenantScoped.GET("/posts/:postId/comments", h.ListComments)
	tenantScoped.GET("/comments/:id", h.GetComment)
	tenantScoped.GET("/comments/:id/replies", h.ListReplies)
	tenantProtected.POST("/comments", h.CreateComment)
	tenantProtected.PUT("/comments/:id", h.UpdateComment)
	tenantProtected.DELETE("/comments/:id", h.DeleteComment)

	// Members (tenant-scoped)
	tenantScoped.GET("/members", h.ListMembers)
	tenantProtected.POST("/members", h.AddMember)
	tenantProtected.GET("/members/:userId", h.GetMember)
	tenantProtected.PUT("/members/:userId/role", h.UpdateMemberRole)
	tenantProtected.DELETE("/members/:userId", h.RemoveMember)

	// Uploads (tenant-scoped, protected)
	tenantProtected.POST("/uploads/presign", h.PresignUpload)
}
