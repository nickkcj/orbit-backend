package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AddMemberRequest struct {
	UserID      string `json:"user_id" validate:"required"`
	RoleID      string `json:"role_id"`
	DisplayName string `json:"display_name"`
}

type UpdateMemberRoleRequest struct {
	RoleID string `json:"role_id" validate:"required"`
}

type UpdateMemberStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

type UpdateProfileRequest struct {
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio"`
}

func (h *Handler) AddMember(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	var req AddMemberRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
	}

	var member interface{}

	if req.RoleID != "" {
		roleID, err := uuid.Parse(req.RoleID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid role_id"})
		}
		member, err = h.services.Member.Add(c.Request().Context(), tenant.ID, userID, roleID, req.DisplayName)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
	} else {
		member, err = h.services.Member.AddWithDefaultRole(c.Request().Context(), tenant.ID, userID, req.DisplayName)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		}
	}

	// Send welcome notification
	go func() {
		h.services.Notification.NotifyWelcome(c.Request().Context(), tenant.ID, userID, tenant.Name)
	}()

	return c.JSON(http.StatusCreated, member)
}

func (h *Handler) GetMember(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
	}

	member, err := h.services.Member.GetWithRole(c.Request().Context(), tenant.ID, userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "member not found"})
	}

	return c.JSON(http.StatusOK, member)
}

func (h *Handler) ListMembers(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	members, err := h.services.Member.ListByTenant(c.Request().Context(), tenant.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list members"})
	}

	return c.JSON(http.StatusOK, members)
}

func (h *Handler) ListUserTenants(c echo.Context) error {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
	}

	tenants, err := h.services.Member.ListTenantsByUser(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list tenants"})
	}

	return c.JSON(http.StatusOK, tenants)
}

func (h *Handler) UpdateMemberRole(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
	}

	var req UpdateMemberRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid role_id"})
	}

	member, err := h.services.Member.UpdateRole(c.Request().Context(), tenant.ID, userID, roleID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, member)
}

func (h *Handler) RemoveMember(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
	}

	if err := h.services.Member.Remove(c.Request().Context(), tenant.ID, userID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

// GetMemberProfile returns detailed profile info for a member
func (h *Handler) GetMemberProfile(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
	}

	profile, err := h.services.Member.GetProfile(c.Request().Context(), tenant.ID, userID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "member not found"})
	}

	return c.JSON(http.StatusOK, profile)
}

// GetMyProfile returns the current user's profile
func (h *Handler) GetMyProfile(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	profile, err := h.services.Member.GetProfile(c.Request().Context(), tenant.ID, user.ID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "profile not found"})
	}

	return c.JSON(http.StatusOK, profile)
}

// UpdateMyProfile updates the current user's profile
func (h *Handler) UpdateMyProfile(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	_, err := h.services.Member.UpdateProfile(c.Request().Context(), tenant.ID, user.ID, req.DisplayName, req.Bio)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update profile"})
	}

	// Return updated profile
	profile, err := h.services.Member.GetProfile(c.Request().Context(), tenant.ID, user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get profile"})
	}

	return c.JSON(http.StatusOK, profile)
}

// GetMemberPosts returns posts by a specific member
func (h *Handler) GetMemberPosts(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid user_id"})
	}

	posts, err := h.services.Post.ListByAuthor(c.Request().Context(), tenant.ID, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get posts"})
	}

	return c.JSON(http.StatusOK, posts)
}
