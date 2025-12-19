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
	v1.GET("/users/:userId/tenants", h.ListUserTenants)

	// Categories (scoped by tenant)
	v1.GET("/tenants/:tenantId/categories", h.ListCategories)
	v1.POST("/categories", h.CreateCategory)
	v1.GET("/categories/:id", h.GetCategory)
	v1.PUT("/categories/:id", h.UpdateCategory)
	v1.DELETE("/categories/:id", h.DeleteCategory)

	// Posts (scoped by tenant)
	v1.GET("/tenants/:tenantId/posts", h.ListPosts)
	v1.POST("/posts", h.CreatePost)
	v1.GET("/posts/:id", h.GetPost)
	v1.PUT("/posts/:id", h.UpdatePost)
	v1.POST("/posts/:id/publish", h.PublishPost)
	v1.DELETE("/posts/:id", h.DeletePost)

	// Comments (scoped by post)
	v1.GET("/posts/:postId/comments", h.ListComments)
	v1.POST("/comments", h.CreateComment)
	v1.GET("/comments/:id", h.GetComment)
	v1.GET("/comments/:id/replies", h.ListReplies)
	v1.PUT("/comments/:id", h.UpdateComment)
	v1.DELETE("/comments/:id", h.DeleteComment)

	// Members (scoped by tenant)
	v1.GET("/tenants/:tenantId/members", h.ListMembers)
	v1.POST("/tenants/:tenantId/members", h.AddMember)
	v1.GET("/tenants/:tenantId/members/:userId", h.GetMember)
	v1.PUT("/tenants/:tenantId/members/:userId/role", h.UpdateMemberRole)
	v1.DELETE("/tenants/:tenantId/members/:userId", h.RemoveMember)
}
