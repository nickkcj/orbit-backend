package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type NotificationResponse struct {
	ID        string      `json:"id"`
	Type      string      `json:"type"`
	Title     string      `json:"title"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	ReadAt    *string     `json:"read_at,omitempty"`
	CreatedAt string      `json:"created_at"`
}

type UnreadCountResponse struct {
	Count int64 `json:"count"`
}

func (h *Handler) ListNotifications(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	// Parse query params
	limit := int32(20)
	offset := int32(0)

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = int32(parsed)
		}
	}
	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}

	notifications, err := h.services.Notification.List(c.Request().Context(), tenant.ID, user.ID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list notifications"})
	}

	return c.JSON(http.StatusOK, notifications)
}

func (h *Handler) GetUnreadCount(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	count, err := h.services.Notification.CountUnread(c.Request().Context(), tenant.ID, user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to count notifications"})
	}

	return c.JSON(http.StatusOK, UnreadCountResponse{Count: count})
}

func (h *Handler) MarkNotificationRead(c echo.Context) error {
	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid notification id"})
	}

	if err := h.services.Notification.MarkRead(c.Request().Context(), notificationID, user.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to mark notification as read"})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) MarkAllNotificationsRead(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	if err := h.services.Notification.MarkAllRead(c.Request().Context(), tenant.ID, user.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to mark all notifications as read"})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) DeleteNotification(c echo.Context) error {
	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	notificationID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid notification id"})
	}

	if err := h.services.Notification.Delete(c.Request().Context(), notificationID, user.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete notification"})
	}

	return c.NoContent(http.StatusNoContent)
}
