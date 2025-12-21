package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

type DashboardResponse struct {
	Stats         interface{} `json:"stats"`
	MembersGrowth interface{} `json:"members_growth"`
	PostsGrowth   interface{} `json:"posts_growth"`
	TopPosts      interface{} `json:"top_posts"`
	RecentMembers interface{} `json:"recent_members"`
}

func (h *Handler) GetDashboard(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	// Check if user is owner or admin
	ctx := c.Request().Context()
	isOwnerOrAdmin, err := h.services.Member.IsOwnerOrAdmin(ctx, tenant.ID, user.ID)
	if err != nil || !isOwnerOrAdmin {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: "only owners and admins can view analytics"})
	}

	// Get all analytics data
	stats, err := h.services.Analytics.GetStats(ctx, tenant.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get stats"})
	}

	membersGrowth, err := h.services.Analytics.GetMembersGrowth(ctx, tenant.ID, 30)
	if err != nil {
		membersGrowth = []interface{}{}
	}

	postsGrowth, err := h.services.Analytics.GetPostsGrowth(ctx, tenant.ID, 30)
	if err != nil {
		postsGrowth = []interface{}{}
	}

	topPosts, err := h.services.Analytics.GetTopPosts(ctx, tenant.ID, 5)
	if err != nil {
		topPosts = []interface{}{}
	}

	recentMembers, err := h.services.Analytics.GetRecentMembers(ctx, tenant.ID, 5)
	if err != nil {
		recentMembers = []interface{}{}
	}

	return c.JSON(http.StatusOK, DashboardResponse{
		Stats:         stats,
		MembersGrowth: membersGrowth,
		PostsGrowth:   postsGrowth,
		TopPosts:      topPosts,
		RecentMembers: recentMembers,
	})
}

func (h *Handler) GetAnalyticsStats(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	ctx := c.Request().Context()
	isOwnerOrAdmin, err := h.services.Member.IsOwnerOrAdmin(ctx, tenant.ID, user.ID)
	if err != nil || !isOwnerOrAdmin {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: "only owners and admins can view analytics"})
	}

	stats, err := h.services.Analytics.GetStats(ctx, tenant.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get stats"})
	}

	return c.JSON(http.StatusOK, stats)
}

func (h *Handler) GetMembersGrowth(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	ctx := c.Request().Context()
	isOwnerOrAdmin, err := h.services.Member.IsOwnerOrAdmin(ctx, tenant.ID, user.ID)
	if err != nil || !isOwnerOrAdmin {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: "only owners and admins can view analytics"})
	}

	days := 30
	if d := c.QueryParam("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 365 {
			days = parsed
		}
	}

	growth, err := h.services.Analytics.GetMembersGrowth(ctx, tenant.ID, days)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get growth data"})
	}

	return c.JSON(http.StatusOK, growth)
}

func (h *Handler) GetTopPosts(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	ctx := c.Request().Context()
	isOwnerOrAdmin, err := h.services.Member.IsOwnerOrAdmin(ctx, tenant.ID, user.ID)
	if err != nil || !isOwnerOrAdmin {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: "only owners and admins can view analytics"})
	}

	limit := int32(10)
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = int32(parsed)
		}
	}

	posts, err := h.services.Analytics.GetTopPosts(ctx, tenant.ID, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get top posts"})
	}

	return c.JSON(http.StatusOK, posts)
}
