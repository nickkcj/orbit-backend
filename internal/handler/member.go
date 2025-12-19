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
