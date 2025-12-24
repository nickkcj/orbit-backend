package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/nickkcj/orbit-backend/internal/middleware"
	"github.com/nickkcj/orbit-backend/internal/websocket"
)

func (h *Handler) RegisterRoutes(
	e *echo.Echo,
	authMiddleware *middleware.AuthMiddleware,
	tenantMiddleware *middleware.TenantMiddleware,
	permissionMiddleware *middleware.PermissionMiddleware,
	wsHandler *websocket.Handler,
) {
	// Health
	e.GET("/health", h.Health)

	// WebSocket endpoint (no middleware - auth handled in handler)
	if wsHandler != nil {
		e.GET("/ws", wsHandler.HandleWebSocket)
	}

	// Webhooks (public endpoints for external services)
	webhooks := e.Group("/webhooks")
	webhooks.POST("/r2", h.HandleR2Webhook)
	webhooks.POST("/stream", h.HandleStreamWebhook)

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
	tenantProtected.POST("/categories", h.CreateCategory, permissionMiddleware.RequirePermission("categories.manage"))
	tenantProtected.GET("/categories/:id", h.GetCategory)
	tenantProtected.PUT("/categories/:id", h.UpdateCategory, permissionMiddleware.RequirePermission("categories.manage"))
	tenantProtected.DELETE("/categories/:id", h.DeleteCategory, permissionMiddleware.RequirePermission("categories.manage"))

	// Posts (tenant-scoped)
	tenantScoped.GET("/posts", h.ListPosts)
	tenantScoped.GET("/posts/:id", h.GetPost)
	tenantProtected.POST("/posts", h.CreatePost, permissionMiddleware.RequirePermission("posts.create"))
	tenantProtected.PUT("/posts/:id", h.UpdatePost, permissionMiddleware.RequireAnyPermission("posts.edit", "posts.edit_own"))
	tenantProtected.POST("/posts/:id/publish", h.PublishPost, permissionMiddleware.RequireAnyPermission("posts.edit", "posts.edit_own"))
	tenantProtected.DELETE("/posts/:id", h.DeletePost, permissionMiddleware.RequireAnyPermission("posts.delete", "posts.delete_own"))

	// Comments (tenant-scoped)
	tenantScoped.GET("/posts/:postId/comments", h.ListComments)
	tenantScoped.GET("/comments/:id", h.GetComment)
	tenantScoped.GET("/comments/:id/replies", h.ListReplies)
	tenantProtected.POST("/comments", h.CreateComment, permissionMiddleware.RequirePermission("comments.create"))
	tenantProtected.PUT("/comments/:id", h.UpdateComment, permissionMiddleware.RequirePermission("comments.edit_own"))
	tenantProtected.DELETE("/comments/:id", h.DeleteComment, permissionMiddleware.RequireAnyPermission("comments.delete", "comments.delete_own"))

	// Members (tenant-scoped)
	tenantScoped.GET("/members", h.ListMembers)
	tenantProtected.POST("/members", h.AddMember, permissionMiddleware.RequirePermission("members.invite"))
	tenantProtected.GET("/members/:userId", h.GetMember)
	tenantProtected.PUT("/members/:userId/role", h.UpdateMemberRole, permissionMiddleware.RequirePermission("members.manage"))
	tenantProtected.DELETE("/members/:userId", h.RemoveMember, permissionMiddleware.RequirePermission("members.remove"))

	// Profile (tenant-scoped)
	tenantScoped.GET("/profile/:userId", h.GetMemberProfile)
	tenantScoped.GET("/profile/:userId/posts", h.GetMemberPosts)
	tenantProtected.GET("/profile/me", h.GetMyProfile)
	tenantProtected.PUT("/profile/me", h.UpdateMyProfile)

	// Uploads (tenant-scoped, protected)
	tenantProtected.POST("/uploads/presign", h.PresignUpload)
	tenantProtected.POST("/uploads/presign-image", h.PresignImageUpload)

	// Tenant Settings (tenant-scoped, protected - requires settings.edit permission)
	tenantProtected.PUT("/settings", h.UpdateTenantSettings, permissionMiddleware.RequirePermission("settings.edit"))
	tenantProtected.PUT("/settings/logo", h.UpdateTenantLogo, permissionMiddleware.RequirePermission("settings.edit"))

	// Notifications (tenant-scoped, protected)
	tenantProtected.GET("/notifications", h.ListNotifications)
	tenantProtected.GET("/notifications/unread/count", h.GetUnreadCount)
	tenantProtected.POST("/notifications/read-all", h.MarkAllNotificationsRead)
	tenantProtected.POST("/notifications/:id/read", h.MarkNotificationRead)
	tenantProtected.DELETE("/notifications/:id", h.DeleteNotification)

	// Analytics (tenant-scoped, protected - owner/admin only)
	tenantProtected.GET("/analytics/dashboard", h.GetDashboard, permissionMiddleware.RequireOwnerOrAdmin())
	tenantProtected.GET("/analytics/stats", h.GetAnalyticsStats, permissionMiddleware.RequireOwnerOrAdmin())
	tenantProtected.GET("/analytics/members/growth", h.GetMembersGrowth, permissionMiddleware.RequireOwnerOrAdmin())
	tenantProtected.GET("/analytics/posts/top", h.GetTopPosts, permissionMiddleware.RequireOwnerOrAdmin())

	// Likes (tenant-scoped, protected)
	tenantProtected.GET("/posts/:id/like", h.GetPostLikeStatus)
	tenantProtected.POST("/posts/:id/like", h.LikePost)
	tenantProtected.DELETE("/posts/:id/like", h.UnlikePost)
	tenantProtected.POST("/comments/:id/like", h.LikeComment)
	tenantProtected.DELETE("/comments/:id/like", h.UnlikeComment)

	// Videos (tenant-scoped)
	tenantScoped.GET("/videos", h.ListVideos, permissionMiddleware.RequirePermission("videos.view"))
	tenantScoped.GET("/videos/:id", h.GetVideo, permissionMiddleware.RequirePermission("videos.view"))
	tenantProtected.POST("/videos", h.InitiateVideoUpload, permissionMiddleware.RequirePermission("videos.upload"))
	tenantProtected.POST("/videos/:id/confirm", h.ConfirmVideoUpload, permissionMiddleware.RequirePermission("videos.upload"))
	tenantProtected.GET("/videos/:id/token", h.GetVideoPlaybackToken, permissionMiddleware.RequirePermission("videos.view"))
	tenantProtected.DELETE("/videos/:id", h.DeleteVideo, permissionMiddleware.RequireAnyPermission("videos.delete", "videos.delete_own"))

	// Roles (tenant-scoped, protected - requires roles.manage permission)
	tenantProtected.GET("/roles", h.ListRoles, permissionMiddleware.RequirePermission("roles.manage"))
	tenantProtected.POST("/roles", h.CreateRole, permissionMiddleware.RequirePermission("roles.manage"))
	tenantProtected.GET("/roles/:id", h.GetRole, permissionMiddleware.RequirePermission("roles.manage"))
	tenantProtected.PUT("/roles/:id", h.UpdateRole, permissionMiddleware.RequirePermission("roles.manage"))
	tenantProtected.DELETE("/roles/:id", h.DeleteRole, permissionMiddleware.RequirePermission("roles.manage"))
	tenantProtected.PUT("/roles/:id/permissions", h.SetRolePermissions, permissionMiddleware.RequirePermission("roles.manage"))

	// Permissions (tenant-scoped, protected - requires roles.manage permission)
	tenantProtected.GET("/permissions", h.ListPermissions, permissionMiddleware.RequirePermission("roles.manage"))

	// ============================================
	// COURSES (tenant-scoped)
	// ============================================

	// Courses - Public (published only)
	tenantScoped.GET("/courses", h.ListCourses, permissionMiddleware.RequirePermission("courses.view"))
	tenantScoped.GET("/courses/:id", h.GetCourse, permissionMiddleware.RequirePermission("courses.view"))
	tenantScoped.GET("/courses/:id/structure", h.GetCourseStructure, permissionMiddleware.RequirePermission("courses.view"))

	// Courses - Protected (CRUD)
	tenantProtected.POST("/courses", h.CreateCourse, permissionMiddleware.RequirePermission("courses.create"))
	tenantProtected.PUT("/courses/:id", h.UpdateCourse, permissionMiddleware.RequireAnyPermission("courses.edit", "courses.edit_own"))
	tenantProtected.POST("/courses/:id/publish", h.PublishCourse, permissionMiddleware.RequirePermission("courses.publish"))
	tenantProtected.POST("/courses/:id/unpublish", h.UnpublishCourse, permissionMiddleware.RequirePermission("courses.publish"))
	tenantProtected.DELETE("/courses/:id", h.DeleteCourse, permissionMiddleware.RequireAnyPermission("courses.delete", "courses.delete_own"))

	// Modules - Protected
	tenantProtected.GET("/courses/:courseId/modules", h.ListModules, permissionMiddleware.RequirePermission("courses.view"))
	tenantProtected.POST("/courses/:courseId/modules", h.CreateModule, permissionMiddleware.RequireAnyPermission("courses.edit", "courses.edit_own"))
	tenantProtected.GET("/modules/:id", h.GetModule, permissionMiddleware.RequirePermission("courses.view"))
	tenantProtected.PUT("/modules/:id", h.UpdateModule, permissionMiddleware.RequireAnyPermission("courses.edit", "courses.edit_own"))
	tenantProtected.PUT("/modules/:id/reorder", h.ReorderModule, permissionMiddleware.RequireAnyPermission("courses.edit", "courses.edit_own"))
	tenantProtected.DELETE("/modules/:id", h.DeleteModule, permissionMiddleware.RequireAnyPermission("courses.edit", "courses.edit_own"))

	// Lessons - Protected
	tenantProtected.GET("/modules/:moduleId/lessons", h.ListLessons, permissionMiddleware.RequirePermission("courses.view"))
	tenantProtected.POST("/modules/:moduleId/lessons", h.CreateLesson, permissionMiddleware.RequireAnyPermission("courses.edit", "courses.edit_own"))
	tenantProtected.GET("/lessons/:id", h.GetLesson, permissionMiddleware.RequirePermission("courses.view"))
	tenantProtected.PUT("/lessons/:id", h.UpdateLesson, permissionMiddleware.RequireAnyPermission("courses.edit", "courses.edit_own"))
	tenantProtected.PUT("/lessons/:id/reorder", h.ReorderLesson, permissionMiddleware.RequireAnyPermission("courses.edit", "courses.edit_own"))
	tenantProtected.DELETE("/lessons/:id", h.DeleteLesson, permissionMiddleware.RequireAnyPermission("courses.edit", "courses.edit_own"))

	// ============================================
	// ENROLLMENTS (tenant-scoped)
	// ============================================

	// Enrollments - User's own enrollments
	tenantProtected.POST("/enrollments", h.EnrollInCourse, permissionMiddleware.RequirePermission("enrollments.enroll"))
	tenantProtected.GET("/enrollments", h.GetMyEnrollments, permissionMiddleware.RequirePermission("enrollments.view"))
	tenantProtected.GET("/enrollments/continue", h.GetContinueLearning, permissionMiddleware.RequirePermission("enrollments.view"))

	// Course enrollment status
	tenantProtected.GET("/courses/:id/enrollment", h.GetCourseEnrollmentStatus, permissionMiddleware.RequirePermission("enrollments.view"))
	tenantProtected.DELETE("/courses/:id/enrollment", h.DropCourseEnrollment, permissionMiddleware.RequirePermission("enrollments.view"))
	tenantProtected.GET("/courses/:id/progress", h.GetCourseProgress, permissionMiddleware.RequirePermission("enrollments.view"))

	// Admin: List course enrollments
	tenantProtected.GET("/courses/:id/enrollments", h.ListCourseEnrollments, permissionMiddleware.RequirePermission("enrollments.manage"))

	// ============================================
	// LESSON PLAYER (tenant-scoped)
	// ============================================

	// Lesson player endpoints
	tenantProtected.GET("/learn/lessons/:id", h.GetLessonForPlayer, permissionMiddleware.RequirePermission("enrollments.view"))
	tenantProtected.POST("/learn/lessons/:id/complete", h.MarkLessonComplete, permissionMiddleware.RequirePermission("enrollments.view"))
	tenantProtected.DELETE("/learn/lessons/:id/complete", h.UnmarkLessonComplete, permissionMiddleware.RequirePermission("enrollments.view"))
	tenantProtected.PUT("/learn/lessons/:id/video-progress", h.UpdateLessonVideoProgress, permissionMiddleware.RequirePermission("enrollments.view"))
}
