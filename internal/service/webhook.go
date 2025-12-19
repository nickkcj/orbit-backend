package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nickkcj/orbit-backend/internal/database"
	"github.com/sqlc-dev/pqtype"
)

type WebhookService struct {
	db *database.Queries
}

func NewWebhookService(db *database.Queries) *WebhookService {
	return &WebhookService{db: db}
}

type WebhookEvent struct {
	ID             uuid.UUID       `json:"id"`
	Provider       string          `json:"provider"`
	EventType      string          `json:"event_type"`
	Payload        json.RawMessage `json:"payload"`
	Status         string          `json:"status"`
	IdempotencyKey *string         `json:"idempotency_key,omitempty"`
}

// LogEvent logs a webhook event to the database
func (s *WebhookService) LogEvent(ctx context.Context, provider, eventType string, payload []byte) (*WebhookEvent, error) {
	// Generate idempotency key from payload hash
	idempotencyKey := uuid.NewSHA1(uuid.NameSpaceOID, payload).String()

	// Check if already processed
	existing, err := s.db.GetWebhookEventByKey(ctx, database.GetWebhookEventByKeyParams{
		Provider:       provider,
		IdempotencyKey: &idempotencyKey,
	})
	if err == nil {
		// Already exists
		return &WebhookEvent{
			ID:             existing.ID,
			Provider:       existing.Provider,
			EventType:      existing.EventType,
			Payload:        existing.Payload.RawMessage,
			Status:         existing.Status,
			IdempotencyKey: existing.IdempotencyKey,
		}, nil
	}

	// Create new event
	event, err := s.db.CreateWebhookEvent(ctx, database.CreateWebhookEventParams{
		Provider:       provider,
		EventType:      eventType,
		Payload:        pqtype.NullRawMessage{RawMessage: payload, Valid: true},
		IdempotencyKey: &idempotencyKey,
	})
	if err != nil {
		return nil, err
	}

	return &WebhookEvent{
		ID:             event.ID,
		Provider:       event.Provider,
		EventType:      event.EventType,
		Payload:        event.Payload.RawMessage,
		Status:         event.Status,
		IdempotencyKey: event.IdempotencyKey,
	}, nil
}

// MarkProcessed marks a webhook event as successfully processed
func (s *WebhookService) MarkProcessed(ctx context.Context, id uuid.UUID) error {
	return s.db.MarkWebhookProcessed(ctx, id)
}

// MarkFailed marks a webhook event as failed
func (s *WebhookService) MarkFailed(ctx context.Context, id uuid.UUID, errorMsg string) error {
	return s.db.MarkWebhookFailed(ctx, database.MarkWebhookFailedParams{
		ID:           id,
		ErrorMessage: &errorMsg,
	})
}
