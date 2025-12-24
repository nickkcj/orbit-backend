package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nickkcj/orbit-backend/internal/middleware"
	"github.com/nickkcj/orbit-backend/internal/service"
)

// ListRoles handles GET /roles
func (h *Handler) ListRoles(c echo.Context) error {
	tenant := middleware.GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "tenant required"})
	}

	roles, err := h.services.Role.ListByTenant(c.Request().Context(), tenant.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list roles"})
	}

	return c.JSON(http.StatusOK, roles)
}

// GetRole handles GET /roles/:id
func (h *Handler) GetRole(c echo.Context) error {
	tenant := middleware.GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "tenant required"})
	}

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid role ID"})
	}

	role, err := h.services.Role.GetByID(c.Request().Context(), roleID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "role not found"})
	}

	// Verify role belongs to tenant
	if role.TenantID != tenant.ID {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "role not found"})
	}

	return c.JSON(http.StatusOK, role)
}

// CreateRole handles POST /roles
func (h *Handler) CreateRole(c echo.Context) error {
	tenant := middleware.GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "tenant required"})
	}

	var req service.CreateRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.Slug == "" || req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "slug and name are required"})
	}

	role, err := h.services.Role.Create(c.Request().Context(), tenant.ID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, role)
}

// UpdateRole handles PUT /roles/:id
func (h *Handler) UpdateRole(c echo.Context) error {
	tenant := middleware.GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "tenant required"})
	}

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid role ID"})
	}

	// Verify role belongs to tenant
	existingRole, err := h.services.Role.GetByID(c.Request().Context(), roleID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "role not found"})
	}

	if existingRole.TenantID != tenant.ID {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "role not found"})
	}

	var req service.UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	role, err := h.services.Role.Update(c.Request().Context(), roleID, &req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, role)
}

// DeleteRole handles DELETE /roles/:id
func (h *Handler) DeleteRole(c echo.Context) error {
	tenant := middleware.GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "tenant required"})
	}

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid role ID"})
	}

	// Verify role belongs to tenant
	existingRole, err := h.services.Role.GetByID(c.Request().Context(), roleID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "role not found"})
	}

	if existingRole.TenantID != tenant.ID {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "role not found"})
	}

	if err := h.services.Role.Delete(c.Request().Context(), roleID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

// ListPermissions handles GET /permissions
func (h *Handler) ListPermissions(c echo.Context) error {
	permissions, err := h.services.Role.ListAllPermissions(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list permissions"})
	}

	return c.JSON(http.StatusOK, permissions)
}

// SetRolePermissions handles PUT /roles/:id/permissions
func (h *Handler) SetRolePermissions(c echo.Context) error {
	tenant := middleware.GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "tenant required"})
	}

	roleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid role ID"})
	}

	// Verify role belongs to tenant
	existingRole, err := h.services.Role.GetByID(c.Request().Context(), roleID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "role not found"})
	}

	if existingRole.TenantID != tenant.ID {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "role not found"})
	}

	if existingRole.IsSystem {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "cannot modify system role permissions"})
	}

	var req struct {
		Permissions []string `json:"permissions"`
	}
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if err := h.services.Role.SetPermissions(c.Request().Context(), roleID, req.Permissions); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	// Return updated role
	role, err := h.services.Role.GetByID(c.Request().Context(), roleID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get updated role"})
	}

	return c.JSON(http.StatusOK, role)
}
