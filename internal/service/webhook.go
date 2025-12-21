package service

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nickkcj/orbit-backend/internal/database"
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
	IdempotencyKey string          `json:"idempotency_key,omitempty"`
}

// LogEvent logs a webhook event to the database
func (s *WebhookService) LogEvent(ctx context.Context, provider, eventType string, payload []byte) (*WebhookEvent, error) {
	// Generate idempotency key from payload hash
	idempotencyKey := uuid.NewSHA1(uuid.NameSpaceOID, payload).String()

	// Check if already processed
	existing, err := s.db.GetWebhookEventByKey(ctx, database.GetWebhookEventByKeyParams{
		Provider:       provider,
		IdempotencyKey: sql.NullString{String: idempotencyKey, Valid: true},
	})
	if err == nil {
		// Already exists
		idemKey := ""
		if existing.IdempotencyKey.Valid {
			idemKey = existing.IdempotencyKey.String
		}
		return &WebhookEvent{
			ID:             existing.ID,
			Provider:       existing.Provider,
			EventType:      existing.EventType,
			Payload:        existing.Payload,
			Status:         existing.Status,
			IdempotencyKey: idemKey,
		}, nil
	}

	// Create new event
	event, err := s.db.CreateWebhookEvent(ctx, database.CreateWebhookEventParams{
		Provider:       provider,
		EventType:      eventType,
		Payload:        payload,
		IdempotencyKey: sql.NullString{String: idempotencyKey, Valid: true},
	})
	if err != nil {
		return nil, err
	}

	idemKey := ""
	if event.IdempotencyKey.Valid {
		idemKey = event.IdempotencyKey.String
	}

	return &WebhookEvent{
		ID:             event.ID,
		Provider:       event.Provider,
		EventType:      event.EventType,
		Payload:        event.Payload,
		Status:         event.Status,
		IdempotencyKey: idemKey,
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
		ErrorMessage: sql.NullString{String: errorMsg, Valid: errorMsg != ""},
	})
}
