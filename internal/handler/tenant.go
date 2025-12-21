package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nickkcj/orbit-backend/internal/service"
)

type CreateTenantRequest struct {
	Slug        string `json:"slug" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type UpdateTenantSettingsRequest struct {
	Theme *service.ThemeSettings `json:"theme"`
}

type UpdateTenantLogoRequest struct {
	LogoURL string `json:"logo_url" validate:"required"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (h *Handler) GetTenantBySlug(c echo.Context) error {
	slug := c.Param("slug")

	tenant, err := h.services.Tenant.GetBySlug(c.Request().Context(), slug)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "tenant not found",
		})
	}

	return c.JSON(http.StatusOK, tenant)
}

func (h *Handler) CreateTenant(c echo.Context) error {
	// Get authenticated user
	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	var req CreateTenantRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "invalid request body",
		})
	}

	ctx := c.Request().Context()

	// Create the tenant
	tenant, err := h.services.Tenant.Create(ctx, req.Slug, req.Name, req.Description)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
	}

	// Create default roles for the tenant
	ownerRole, err := h.services.Role.CreateDefaultRoles(ctx, tenant.ID)
	if err != nil {
		// Tenant created but roles failed - log but continue
		// The owner role is the first one returned
	}

	// Add the creator as owner member
	if ownerRole != nil {
		_, err = h.services.Member.Add(ctx, tenant.ID, user.ID, ownerRole.ID, user.Name)
		if err != nil {
			// Log error but don't fail - tenant is created
		}
	}

	return c.JSON(http.StatusCreated, tenant)
}

func (h *Handler) ListTenants(c echo.Context) error {
	tenants, err := h.services.Tenant.List(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "failed to list tenants",
		})
	}

	return c.JSON(http.StatusOK, tenants)
}

func (h *Handler) UpdateTenantSettings(c echo.Context) error {
	// Get tenant from context (set by tenant middleware)
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	// Get authenticated user
	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	// Check if user is owner or admin
	ctx := c.Request().Context()
	isOwnerOrAdmin, err := h.services.Member.IsOwnerOrAdmin(ctx, tenant.ID, user.ID)
	if err != nil || !isOwnerOrAdmin {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: "only owners and admins can update settings"})
	}

	var req UpdateTenantSettingsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	settings := service.TenantSettings{
		Theme: req.Theme,
	}

	updatedTenant, err := h.services.Tenant.UpdateSettings(ctx, tenant.ID, settings)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update settings"})
	}

	return c.JSON(http.StatusOK, updatedTenant)
}

func (h *Handler) UpdateTenantLogo(c echo.Context) error {
	// Get tenant from context
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	// Get authenticated user
	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	// Check if user is owner or admin
	ctx := c.Request().Context()
	isOwnerOrAdmin, err := h.services.Member.IsOwnerOrAdmin(ctx, tenant.ID, user.ID)
	if err != nil || !isOwnerOrAdmin {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: "only owners and admins can update logo"})
	}

	var req UpdateTenantLogoRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	updatedTenant, err := h.services.Tenant.UpdateLogo(ctx, tenant.ID, req.LogoURL)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update logo"})
	}

	return c.JSON(http.StatusOK, updatedTenant)
}
