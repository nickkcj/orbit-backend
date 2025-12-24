package handler

import (
	"net/http"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
)

type PresignUploadRequest struct {
	Filename    string `json:"filename" validate:"required"`
	ContentType string `json:"content_type"`
}

type PresignUploadResponse struct {
	UploadURL string `json:"upload_url"`
	FileKey   string `json:"file_key"`
	ExpiresIn int    `json:"expires_in"`
}

// PresignUpload generates a presigned URL for uploading a file to R2
func (h *Handler) PresignUpload(c echo.Context) error {
	// Check if storage is configured
	if h.services.Storage == nil {
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error: "storage service not configured",
		})
	}

	// Get tenant from context (required for tenant-scoped uploads)
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "tenant context required",
		})
	}

	// Get authenticated user
	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "authentication required",
		})
	}

	var req PresignUploadRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "invalid request body",
		})
	}

	if req.Filename == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "filename is required",
		})
	}

	// Determine content type from filename if not provided
	contentType := req.ContentType
	if contentType == "" {
		contentType = getContentTypeFromFilename(req.Filename)
	}

	// Validate content type (only allow video formats)
	if !isAllowedVideoType(contentType) {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "only video files are allowed (mp4, webm, mov, avi)",
		})
	}

	// Generate presigned URL
	result, err := h.services.Storage.GenerateUploadURL(
		c.Request().Context(),
		tenant.ID,
		req.Filename,
		contentType,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "failed to generate upload URL",
		})
	}

	return c.JSON(http.StatusOK, PresignUploadResponse{
		UploadURL: result.UploadURL,
		FileKey:   result.FileKey,
		ExpiresIn: result.ExpiresIn,
	})
}

func getContentTypeFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp4":
		return "video/mp4"
	case ".webm":
		return "video/webm"
	case ".mov":
		return "video/quicktime"
	case ".avi":
		return "video/x-msvideo"
	case ".mkv":
		return "video/x-matroska"
	default:
		return "application/octet-stream"
	}
}

func isAllowedVideoType(contentType string) bool {
	allowedTypes := []string{
		"video/mp4",
		"video/webm",
		"video/quicktime",
		"video/x-msvideo",
		"video/x-matroska",
	}

	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

func isAllowedImageType(contentType string) bool {
	allowedTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
	}

	for _, allowed := range allowedTypes {
		if contentType == allowed {
			return true
		}
	}
	return false
}

func getImageContentTypeFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

// PresignImageUpload generates a presigned URL for uploading an image (avatars, etc)
func (h *Handler) PresignImageUpload(c echo.Context) error {
	// Check if storage is configured
	if h.services.Storage == nil {
		return c.JSON(http.StatusServiceUnavailable, ErrorResponse{
			Error: "storage service not configured",
		})
	}

	// Get tenant from context (required for tenant-scoped uploads)
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "tenant context required",
		})
	}

	// Get authenticated user
	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error: "authentication required",
		})
	}

	var req PresignUploadRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "invalid request body",
		})
	}

	if req.Filename == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "filename is required",
		})
	}

	// Determine content type from filename if not provided
	contentType := req.ContentType
	if contentType == "" {
		contentType = getImageContentTypeFromFilename(req.Filename)
	}

	// Validate content type (only allow image formats)
	if !isAllowedImageType(contentType) {
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "only image files are allowed (jpg, png, gif, webp)",
		})
	}

	// Generate presigned URL for images
	result, err := h.services.Storage.GenerateImageUploadURL(
		c.Request().Context(),
		tenant.ID,
		req.Filename,
		contentType,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "failed to generate upload URL",
		})
	}

	return c.JSON(http.StatusOK, PresignUploadResponse{
		UploadURL: result.UploadURL,
		FileKey:   result.FileKey,
		ExpiresIn: result.ExpiresIn,
	})
}
