package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// StreamService handles Cloudflare Stream API operations
type StreamService struct {
	accountID    string
	apiToken     string
	signingKey   string
	webhookSecret string
	httpClient   *http.Client
}

// StreamConfig holds configuration for the Stream service
type StreamConfig struct {
	AccountID     string
	APIToken      string
	SigningKey    string
	WebhookSecret string
}

// StreamVideo represents a video in Cloudflare Stream
type StreamVideo struct {
	UID           string           `json:"uid"`
	Status        StreamStatus     `json:"status"`
	Meta          StreamMeta       `json:"meta"`
	Playback      StreamPlayback   `json:"playback"`
	Thumbnail     string           `json:"thumbnail"`
	Duration      float64          `json:"duration"`
	Size          int64            `json:"size"`
	ReadyToStream bool             `json:"readyToStream"`
	Created       time.Time        `json:"created"`
}

// StreamStatus represents the processing status
type StreamStatus struct {
	State           string `json:"state"` // pendingupload, uploading, queued, inprogress, ready, error
	PctComplete     int    `json:"pctComplete,omitempty"`
	ErrorReasonCode string `json:"errorReasonCode,omitempty"`
	ErrorReasonText string `json:"errorReasonText,omitempty"`
}

// StreamMeta contains video metadata
type StreamMeta struct {
	Name string `json:"name"`
}

// StreamPlayback contains playback URLs
type StreamPlayback struct {
	HLS  string `json:"hls"`
	DASH string `json:"dash"`
}

// DirectUploadResponse contains the response from creating a direct upload
type DirectUploadResponse struct {
	UID       string `json:"uid"`
	UploadURL string `json:"uploadURL"`
}

// StreamWebhookPayload represents the webhook payload from Cloudflare
type StreamWebhookPayload struct {
	UID           string       `json:"uid"`
	ReadyToStream bool         `json:"readyToStream"`
	Status        StreamStatus `json:"status"`
	Meta          StreamMeta   `json:"meta"`
	Duration      float64      `json:"duration"`
	Size          int64        `json:"size"`
	Playback      StreamPlayback `json:"playback"`
	Thumbnail     string       `json:"thumbnail"`
}

// NewStreamService creates a new Stream service
func NewStreamService(cfg *StreamConfig) (*StreamService, error) {
	if cfg.AccountID == "" || cfg.APIToken == "" {
		return nil, fmt.Errorf("cloudflare account ID and API token are required")
	}

	return &StreamService{
		accountID:    cfg.AccountID,
		apiToken:     cfg.APIToken,
		signingKey:   cfg.SigningKey,
		webhookSecret: cfg.WebhookSecret,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// CreateDirectUpload creates a direct upload URL for client-side uploads
func (s *StreamService) CreateDirectUpload(ctx context.Context, maxDurationSeconds int, meta map[string]string) (*DirectUploadResponse, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/stream/direct_upload", s.accountID)

	payload := map[string]interface{}{
		"maxDurationSeconds": maxDurationSeconds,
	}
	if meta != nil {
		payload["meta"] = meta
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cloudflare API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Success bool `json:"success"`
		Result  DirectUploadResponse `json:"result"`
		Errors  []struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		if len(result.Errors) > 0 {
			return nil, fmt.Errorf("cloudflare API error: %s", result.Errors[0].Message)
		}
		return nil, fmt.Errorf("cloudflare API error: unknown error")
	}

	return &result.Result, nil
}

// GetVideo retrieves video details from Cloudflare Stream
func (s *StreamService) GetVideo(ctx context.Context, videoUID string) (*StreamVideo, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/stream/%s", s.accountID, videoUID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cloudflare API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Success bool        `json:"success"`
		Result  StreamVideo `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Result, nil
}

// DeleteVideo deletes a video from Cloudflare Stream
func (s *StreamService) DeleteVideo(ctx context.Context, videoUID string) error {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/stream/%s", s.accountID, videoUID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("cloudflare API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	return nil
}

// GenerateSignedToken creates a signed token for video playback
func (s *StreamService) GenerateSignedToken(videoUID string, expiresIn time.Duration) (string, error) {
	if s.signingKey == "" {
		return "", fmt.Errorf("signing key not configured")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"sub": videoUID,
		"kid": s.signingKey,
		"exp": now.Add(expiresIn).Unix(),
		"nbf": now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	// The signing key should be in PEM format
	signedToken, err := token.SignedString([]byte(s.signingKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return signedToken, nil
}

// VerifyWebhookSignature verifies the signature of a webhook request
func (s *StreamService) VerifyWebhookSignature(body []byte, signature string) bool {
	if s.webhookSecret == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(s.webhookSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}

// ParseWebhookPayload parses the webhook body into a StreamWebhookPayload
func (s *StreamService) ParseWebhookPayload(body []byte) (*StreamWebhookPayload, error) {
	var payload StreamWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse webhook payload: %w", err)
	}
	return &payload, nil
}

// CopyFromURL copies a video from a URL to Cloudflare Stream
func (s *StreamService) CopyFromURL(ctx context.Context, videoURL string, meta map[string]string) (*StreamVideo, error) {
	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/stream/copy", s.accountID)

	payload := map[string]interface{}{
		"url": videoURL,
	}
	if meta != nil {
		payload["meta"] = meta
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cloudflare API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Success bool        `json:"success"`
		Result  StreamVideo `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result.Result, nil
}
