-- name: CreateNotification :one
INSERT INTO notifications (tenant_id, user_id, type, title, message, data)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetNotification :one
SELECT * FROM notifications WHERE id = $1;

-- name: ListNotificationsByUser :many
SELECT * FROM notifications
WHERE tenant_id = $1 AND user_id = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: ListUnreadNotifications :many
SELECT * FROM notifications
WHERE tenant_id = $1 AND user_id = $2 AND read_at IS NULL
ORDER BY created_at DESC;

-- name: CountUnreadNotifications :one
SELECT COUNT(*) FROM notifications
WHERE tenant_id = $1 AND user_id = $2 AND read_at IS NULL;

-- name: MarkNotificationRead :exec
UPDATE notifications
SET read_at = NOW()
WHERE id = $1 AND user_id = $2;

-- name: MarkAllNotificationsRead :exec
UPDATE notifications
SET read_at = NOW()
WHERE tenant_id = $1 AND user_id = $2 AND read_at IS NULL;

-- name: DeleteNotification :exec
DELETE FROM notifications WHERE id = $1 AND user_id = $2;

-- name: DeleteOldNotifications :exec
DELETE FROM notifications
WHERE created_at < NOW() - INTERVAL '30 days' AND read_at IS NOT NULL;
