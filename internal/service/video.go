package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nickkcj/orbit-backend/internal/database"
)

// VideoService handles video operations
type VideoService struct {
	db     *database.Queries
	stream *StreamService
}

// NewVideoService creates a new video service
func NewVideoService(db *database.Queries, stream *StreamService) *VideoService {
	return &VideoService{
		db:     db,
		stream: stream,
	}
}

// VideoUploadRequest represents a request to initiate video upload
type VideoUploadRequest struct {
	Title       string
	Description string
	TenantID    uuid.UUID
	UploaderID  uuid.UUID
}

// VideoUploadResponse contains the upload URL and video record
type VideoUploadResponse struct {
	VideoID      uuid.UUID `json:"video_id"`
	UploadURL    string    `json:"upload_url"`
	StreamUID    string    `json:"stream_uid"`
}

// InitiateUpload creates a video record and returns a direct upload URL
func (s *VideoService) InitiateUpload(ctx context.Context, req *VideoUploadRequest) (*VideoUploadResponse, error) {
	if s.stream == nil {
		return nil, fmt.Errorf("stream service not configured")
	}

	// Create direct upload URL in Cloudflare Stream
	meta := map[string]string{
		"name":      req.Title,
		"tenant_id": req.TenantID.String(),
	}

	// Max 30 minutes video
	directUpload, err := s.stream.CreateDirectUpload(ctx, 1800, meta)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload URL: %w", err)
	}

	// Create video record in database
	video, err := s.db.CreateVideo(ctx, database.CreateVideoParams{
		TenantID:    req.TenantID,
		UploaderID:  req.UploaderID,
		Title:       req.Title,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		Provider:    "cloudflare",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create video record: %w", err)
	}

	// Update video with external ID
	_, err = s.db.UpdateVideoStatus(ctx, database.UpdateVideoStatusParams{
		ID:     video.ID,
		Status: "uploading",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update video status: %w", err)
	}

	return &VideoUploadResponse{
		VideoID:   video.ID,
		UploadURL: directUpload.UploadURL,
		StreamUID: directUpload.UID,
	}, nil
}

// VideoResponse represents a video for API responses
type VideoResponse struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	UploaderID   uuid.UUID  `json:"uploader_id"`
	UploaderName string     `json:"uploader_name,omitempty"`
	Title        string     `json:"title"`
	Description  string     `json:"description,omitempty"`
	ThumbnailURL string     `json:"thumbnail_url,omitempty"`
	Duration     int        `json:"duration_seconds,omitempty"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
}

// GetByID retrieves a video by ID
func (s *VideoService) GetByID(ctx context.Context, id uuid.UUID) (*VideoResponse, error) {
	video, err := s.db.GetVideoByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &VideoResponse{
		ID:           video.ID,
		TenantID:     video.TenantID,
		UploaderID:   video.UploaderID,
		Title:        video.Title,
		Description:  video.Description.String,
		ThumbnailURL: video.ThumbnailUrl.String,
		Duration:     int(video.DurationSeconds.Int32),
		Status:       video.Status,
		CreatedAt:    video.CreatedAt,
	}, nil
}

// ListByTenant lists videos for a tenant
func (s *VideoService) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]VideoResponse, error) {
	videos, err := s.db.ListVideosByTenant(ctx, database.ListVideosByTenantParams{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return nil, err
	}

	result := make([]VideoResponse, len(videos))
	for i, v := range videos {
		result[i] = VideoResponse{
			ID:           v.ID,
			TenantID:     v.TenantID,
			UploaderID:   v.UploaderID,
			UploaderName: v.UploaderName,
			Title:        v.Title,
			Description:  v.Description.String,
			ThumbnailURL: v.ThumbnailUrl.String,
			Duration:     int(v.DurationSeconds.Int32),
			Status:       v.Status,
			CreatedAt:    v.CreatedAt,
		}
	}

	return result, nil
}

// Delete deletes a video
func (s *VideoService) Delete(ctx context.Context, id uuid.UUID) error {
	// Get video to check if we need to delete from Cloudflare
	video, err := s.db.GetVideoByID(ctx, id)
	if err != nil {
		return err
	}

	// Delete from Cloudflare if we have an external ID
	if video.ExternalID.Valid && video.ExternalID.String != "" && s.stream != nil {
		if err := s.stream.DeleteVideo(ctx, video.ExternalID.String); err != nil {
			// Log but don't fail - we still want to delete from our DB
			fmt.Printf("Warning: failed to delete video from Cloudflare: %v\n", err)
		}
	}

	// Delete from database
	return s.db.DeleteVideo(ctx, id)
}

// GeneratePlaybackToken generates a signed token for video playback
func (s *VideoService) GeneratePlaybackToken(ctx context.Context, videoID uuid.UUID) (string, error) {
	video, err := s.db.GetVideoByID(ctx, videoID)
	if err != nil {
		return "", err
	}

	if !video.ExternalID.Valid || video.ExternalID.String == "" {
		return "", fmt.Errorf("video not yet processed")
	}

	if s.stream == nil {
		return "", fmt.Errorf("stream service not configured")
	}

	// Generate token valid for 1 hour
	return s.stream.GenerateSignedToken(video.ExternalID.String, time.Hour)
}

// ProcessWebhook processes a Cloudflare Stream webhook
func (s *VideoService) ProcessWebhook(ctx context.Context, payload *StreamWebhookPayload) error {
	// Find video by external ID
	video, err := s.db.GetVideoByExternalID(ctx, database.GetVideoByExternalIDParams{
		Provider:   "cloudflare",
		ExternalID: sql.NullString{String: payload.UID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("video not found for stream UID %s: %w", payload.UID, err)
	}

	if payload.ReadyToStream {
		// Video is ready
		_, err = s.db.UpdateVideoAfterProcessing(ctx, database.UpdateVideoAfterProcessingParams{
			ID:              video.ID,
			ExternalID:      sql.NullString{String: payload.UID, Valid: true},
			PlaybackUrl:     sql.NullString{String: payload.Playback.HLS, Valid: true},
			ThumbnailUrl:    sql.NullString{String: payload.Thumbnail, Valid: true},
			DurationSeconds: sql.NullInt32{Int32: int32(payload.Duration), Valid: true},
			Resolution:      sql.NullString{}, // Can be extracted from metadata if needed
		})
		if err != nil {
			return fmt.Errorf("failed to update video: %w", err)
		}
	} else if payload.Status.State == "error" {
		// Video processing failed
		_, err = s.db.UpdateVideoStatus(ctx, database.UpdateVideoStatusParams{
			ID:           video.ID,
			Status:       "failed",
			ErrorMessage: sql.NullString{String: payload.Status.ErrorReasonText, Valid: true},
		})
		if err != nil {
			return fmt.Errorf("failed to update video status: %w", err)
		}
	}

	return nil
}

// UpdateExternalID updates the external ID after client-side upload completes
func (s *VideoService) UpdateExternalID(ctx context.Context, videoID uuid.UUID, streamUID string) error {
	_, err := s.db.UpdateVideoStatus(ctx, database.UpdateVideoStatusParams{
		ID:     videoID,
		Status: "processing",
	})
	return err
}
