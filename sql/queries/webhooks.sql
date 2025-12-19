-- name: CreateWebhookEvent :one
INSERT INTO webhook_events (provider, event_type, payload, idempotency_key)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetWebhookEventByKey :one
SELECT * FROM webhook_events WHERE provider = $1 AND idempotency_key = $2;

-- name: MarkWebhookProcessed :exec
UPDATE webhook_events
SET status = 'processed', processed_at = NOW()
WHERE id = $1;

-- name: MarkWebhookFailed :exec
UPDATE webhook_events
SET status = 'failed', error_message = $2, processed_at = NOW()
WHERE id = $1;

-- name: ListPendingWebhooks :many
SELECT * FROM webhook_events
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT $1;
