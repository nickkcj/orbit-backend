package handler

import (
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/nickkcj/orbit-backend/internal/middleware"
	"github.com/nickkcj/orbit-backend/internal/service"
)

// InitiateVideoUpload handles POST /videos - creates upload URL
func (h *Handler) InitiateVideoUpload(c echo.Context) error {
	user := middleware.GetUserFromContext(c)
	tenant := middleware.GetTenantFromContext(c)

	if user == nil || tenant == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req struct {
		Title       string `json:"title" validate:"required"`
		Description string `json:"description"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "title is required"})
	}

	uploadResp, err := h.services.Video.InitiateUpload(c.Request().Context(), &service.VideoUploadRequest{
		Title:       req.Title,
		Description: req.Description,
		TenantID:    tenant.ID,
		UploaderID:  user.ID,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, uploadResp)
}

// ListVideos handles GET /videos
func (h *Handler) ListVideos(c echo.Context) error {
	tenant := middleware.GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "tenant required"})
	}

	limit := int32(20)
	offset := int32(0)

	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := parseIntParam(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = int32(parsed)
		}
	}

	if o := c.QueryParam("offset"); o != "" {
		if parsed, err := parseIntParam(o); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}

	videos, err := h.services.Video.ListByTenant(c.Request().Context(), tenant.ID, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list videos"})
	}

	return c.JSON(http.StatusOK, videos)
}

// GetVideo handles GET /videos/:id
func (h *Handler) GetVideo(c echo.Context) error {
	tenant := middleware.GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "tenant required"})
	}

	videoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid video ID"})
	}

	video, err := h.services.Video.GetByID(c.Request().Context(), videoID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "video not found"})
	}

	// Verify tenant ownership
	if video.TenantID != tenant.ID {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "video not found"})
	}

	return c.JSON(http.StatusOK, video)
}

// GetVideoPlaybackToken handles GET /videos/:id/token
func (h *Handler) GetVideoPlaybackToken(c echo.Context) error {
	tenant := middleware.GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "tenant required"})
	}

	videoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid video ID"})
	}

	// Verify video belongs to tenant
	video, err := h.services.Video.GetByID(c.Request().Context(), videoID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "video not found"})
	}

	if video.TenantID != tenant.ID {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "video not found"})
	}

	token, err := h.services.Video.GeneratePlaybackToken(c.Request().Context(), videoID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

// DeleteVideo handles DELETE /videos/:id
func (h *Handler) DeleteVideo(c echo.Context) error {
	user := middleware.GetUserFromContext(c)
	tenant := middleware.GetTenantFromContext(c)

	if user == nil || tenant == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	videoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid video ID"})
	}

	// Get video to check ownership
	video, err := h.services.Video.GetByID(c.Request().Context(), videoID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "video not found"})
	}

	// Verify tenant ownership
	if video.TenantID != tenant.ID {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "video not found"})
	}

	// Check if user is owner or has delete permission
	// Permission middleware already checked videos.delete or videos.delete_own
	// Here we verify ownership for videos.delete_own
	if video.UploaderID != user.ID {
		// User must have videos.delete permission (not just videos.delete_own)
		hasFullDelete, err := h.services.Permission.HasPermission(
			c.Request().Context(),
			tenant.ID,
			user.ID,
			"videos.delete",
		)
		if err != nil || !hasFullDelete {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "permission denied"})
		}
	}

	if err := h.services.Video.Delete(c.Request().Context(), videoID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete video"})
	}

	return c.NoContent(http.StatusNoContent)
}

// HandleStreamWebhook handles POST /webhooks/stream - Cloudflare Stream webhook
func (h *Handler) HandleStreamWebhook(c echo.Context) error {
	// Read body for signature verification
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed to read body"})
	}

	// Verify webhook signature
	signature := c.Request().Header.Get("Webhook-Signature")
	if h.services.Stream != nil && signature != "" {
		if !h.services.Stream.VerifyWebhookSignature(body, signature) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid signature"})
		}
	}

	// Parse payload
	if h.services.Stream == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{"error": "stream service not configured"})
	}

	payload, err := h.services.Stream.ParseWebhookPayload(body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid payload"})
	}

	// Process webhook
	if err := h.services.Video.ProcessWebhook(c.Request().Context(), payload); err != nil {
		// Log error but return success to avoid retries for non-existent videos
		c.Logger().Errorf("Failed to process stream webhook: %v", err)
	}

	return c.NoContent(http.StatusOK)
}

// ConfirmVideoUpload handles POST /videos/:id/confirm - confirms upload completion
func (h *Handler) ConfirmVideoUpload(c echo.Context) error {
	user := middleware.GetUserFromContext(c)
	tenant := middleware.GetTenantFromContext(c)

	if user == nil || tenant == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	videoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid video ID"})
	}

	var req struct {
		StreamUID string `json:"stream_uid" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	// Verify video belongs to user and tenant
	video, err := h.services.Video.GetByID(c.Request().Context(), videoID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "video not found"})
	}

	if video.TenantID != tenant.ID || video.UploaderID != user.ID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "permission denied"})
	}

	// Update video with stream UID
	if err := h.services.Video.UpdateExternalID(c.Request().Context(), videoID, req.StreamUID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to confirm upload"})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "processing"})
}

func parseIntParam(s string) (int, error) {
	var result int
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int(c-'0')
		} else {
			return 0, echo.ErrBadRequest
		}
	}
	return result, nil
}
