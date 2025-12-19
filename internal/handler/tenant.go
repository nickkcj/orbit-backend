package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type CreateTenantRequest struct {
	Slug        string `json:"slug" validate:"required"`
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
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
	var req CreateTenantRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "invalid request body",
		})
	}

	tenant, err := h.services.Tenant.Create(c.Request().Context(), req.Slug, req.Name, req.Description)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
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
