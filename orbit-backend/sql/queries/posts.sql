-- name: CreatePost :one
INSERT INTO posts (tenant_id, category_id, author_id, title, slug, content, content_format, excerpt, cover_image_url, status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetPostByID :one
SELECT * FROM posts WHERE id = $1;

-- name: GetPostBySlug :one
SELECT * FROM posts WHERE tenant_id = $1 AND slug = $2;

-- name: ListPostsByTenant :many
SELECT
    p.*,
    u.name as author_name,
    u.avatar_url as author_avatar,
    c.name as category_name
FROM posts p
JOIN users u ON p.author_id = u.id
LEFT JOIN categories c ON p.category_id = c.id
WHERE p.tenant_id = $1 AND p.status = 'published'
ORDER BY p.published_at DESC
LIMIT $2 OFFSET $3;

-- name: ListPostsByCategory :many
SELECT
    p.*,
    u.name as author_name,
    u.avatar_url as author_avatar
FROM posts p
JOIN users u ON p.author_id = u.id
WHERE p.tenant_id = $1 AND p.category_id = $2 AND p.status = 'published'
ORDER BY p.published_at DESC
LIMIT $3 OFFSET $4;

-- name: ListPostsByAuthor :many
SELECT * FROM posts
WHERE tenant_id = $1 AND author_id = $2
ORDER BY created_at DESC;

-- name: UpdatePost :one
UPDATE posts
SET title = $2, content = $3, excerpt = $4, cover_image_url = $5, category_id = $6, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: PublishPost :one
UPDATE posts
SET status = 'published', published_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ArchivePost :exec
UPDATE posts SET status = 'archived', updated_at = NOW() WHERE id = $1;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = $1;

-- name: IncrementPostViews :exec
UPDATE posts SET view_count = view_count + 1 WHERE id = $1;

-- name: IncrementPostLikes :exec
UPDATE posts SET like_count = like_count + 1 WHERE id = $1;

-- name: DecrementPostLikes :exec
UPDATE posts SET like_count = like_count - 1 WHERE id = $1 AND like_count > 0;
