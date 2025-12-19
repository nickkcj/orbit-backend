-- name: CreateVideo :one
INSERT INTO videos (tenant_id, uploader_id, title, description, provider, status)
VALUES ($1, $2, $3, $4, $5, 'pending')
RETURNING *;

-- name: GetVideoByID :one
SELECT * FROM videos WHERE id = $1;

-- name: GetVideoByExternalID :one
SELECT * FROM videos WHERE provider = $1 AND external_id = $2;

-- name: ListVideosByTenant :many
SELECT
    v.*,
    u.name as uploader_name
FROM videos v
JOIN users u ON v.uploader_id = u.id
WHERE v.tenant_id = $1 AND v.status = 'ready'
ORDER BY v.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListVideosByPost :many
SELECT * FROM videos WHERE post_id = $1 AND status = 'ready';

-- name: UpdateVideoStatus :one
UPDATE videos
SET status = $2, error_message = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateVideoAfterProcessing :one
UPDATE videos
SET
    external_id = $2,
    playback_url = $3,
    thumbnail_url = $4,
    duration_seconds = $5,
    resolution = $6,
    status = 'ready',
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: AttachVideoToPost :exec
UPDATE videos SET post_id = $2, updated_at = NOW() WHERE id = $1;

-- name: DeleteVideo :exec
DELETE FROM videos WHERE id = $1;
