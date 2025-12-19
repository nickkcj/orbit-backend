package handler

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required"`
}

func (h *Handler) CreateUser(c echo.Context) error {
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "invalid request body",
		})
	}

	user, err := h.services.User.Create(c.Request().Context(), req.Email, req.Password, req.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, user)
}

func (h *Handler) GetUserByEmail(c echo.Context) error {
	email := c.QueryParam("email")
	if email == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "email query param required",
		})
	}

	user, err := h.services.User.GetByEmail(c.Request().Context(), email)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{
			Error: "user not found",
		})
	}

	return c.JSON(http.StatusOK, user)
}
