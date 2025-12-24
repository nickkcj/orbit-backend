package tasks

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// WebhookPayload contains data for processing webhooks
type WebhookPayload struct {
	Provider   string          `json:"provider"` // "r2", "cloudflare_stream", etc.
	EventType  string          `json:"event_type"`
	EventID    uuid.UUID       `json:"event_id"`
	RawPayload json.RawMessage `json:"raw_payload"`
}

// NewProcessWebhookTask creates a new webhook processing task
func NewProcessWebhookTask(payload WebhookPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(
		TypeProcessWebhook,
		data,
		asynq.Queue(QueueCritical),
		asynq.MaxRetry(5),
		asynq.Timeout(2*time.Minute),
		asynq.Retention(48*time.Hour),
	), nil
}
