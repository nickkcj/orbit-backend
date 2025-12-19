-- name: CreateComment :one
INSERT INTO comments (tenant_id, post_id, author_id, parent_id, content)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetCommentByID :one
SELECT * FROM comments WHERE id = $1;

-- name: ListCommentsByPost :many
SELECT
    c.*,
    u.name as author_name,
    u.avatar_url as author_avatar
FROM comments c
JOIN users u ON c.author_id = u.id
WHERE c.post_id = $1 AND c.parent_id IS NULL AND c.status = 'visible'
ORDER BY c.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListReplies :many
SELECT
    c.*,
    u.name as author_name,
    u.avatar_url as author_avatar
FROM comments c
JOIN users u ON c.author_id = u.id
WHERE c.parent_id = $1 AND c.status = 'visible'
ORDER BY c.created_at ASC;

-- name: UpdateComment :one
UPDATE comments
SET content = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: HideComment :exec
UPDATE comments SET status = 'hidden', updated_at = NOW() WHERE id = $1;

-- name: DeleteComment :exec
UPDATE comments SET status = 'deleted', updated_at = NOW() WHERE id = $1;

-- name: IncrementCommentLikes :exec
UPDATE comments SET like_count = like_count + 1 WHERE id = $1;

-- name: DecrementCommentLikes :exec
UPDATE comments SET like_count = like_count - 1 WHERE id = $1 AND like_count > 0;
