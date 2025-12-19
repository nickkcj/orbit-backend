package handler

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

// R2 webhook payload for video ready notification
type R2WebhookPayload struct {
	Action  string `json:"action"`
	Bucket  string `json:"bucket"`
	Object  R2Object `json:"object"`
}

type R2Object struct {
	Key          string `json:"key"`
	Size         int64  `json:"size"`
	ETag         string `json:"etag"`
	VersionID    string `json:"versionId,omitempty"`
}

// Generic webhook response
type WebhookResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// HandleR2Webhook processes webhooks from Cloudflare R2
func (h *Handler) HandleR2Webhook(c echo.Context) error {
	// Read body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, WebhookResponse{
			Success: false,
			Message: "failed to read request body",
		})
	}

	// Verify webhook signature if secret is configured
	webhookSecret := c.Request().Header.Get("X-Webhook-Secret")
	if webhookSecret != "" {
		signature := c.Request().Header.Get("X-R2-Signature")
		if !verifyR2Signature(body, signature, webhookSecret) {
			return c.JSON(http.StatusUnauthorized, WebhookResponse{
				Success: false,
				Message: "invalid signature",
			})
		}
	}

	// Parse payload
	var payload R2WebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return c.JSON(http.StatusBadRequest, WebhookResponse{
			Success: false,
			Message: "invalid payload format",
		})
	}

	// Log the webhook event
	_, err = h.services.Webhook.LogEvent(c.Request().Context(), "r2", payload.Action, body)
	if err != nil {
		// Log error but don't fail the webhook
		c.Logger().Errorf("failed to log webhook event: %v", err)
	}

	// Process based on action
	switch payload.Action {
	case "PutObject":
		// Video uploaded - could trigger processing
		return h.handleVideoUploaded(c, payload)
	case "DeleteObject":
		// Video deleted
		return h.handleVideoDeleted(c, payload)
	default:
		// Unknown action, acknowledge but don't process
		return c.JSON(http.StatusOK, WebhookResponse{
			Success: true,
			Message: "event acknowledged",
		})
	}
}

func (h *Handler) handleVideoUploaded(c echo.Context, payload R2WebhookPayload) error {
	// Extract video info from key path
	// Expected format: tenants/{tenant_id}/videos/{video_uuid}/{filename}
	// For now, just acknowledge - video is ready since R2 doesn't transcode

	return c.JSON(http.StatusOK, WebhookResponse{
		Success: true,
		Message: "video upload processed",
	})
}

func (h *Handler) handleVideoDeleted(c echo.Context, payload R2WebhookPayload) error {
	// Handle video deletion if needed
	return c.JSON(http.StatusOK, WebhookResponse{
		Success: true,
		Message: "video deletion processed",
	})
}

func verifyR2Signature(body []byte, signature, secret string) bool {
	if signature == "" || secret == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
