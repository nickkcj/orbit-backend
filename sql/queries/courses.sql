-- ============================================================================
-- COURSES
-- ============================================================================

-- name: CreateCourse :one
INSERT INTO courses (tenant_id, author_id, title, slug, description, thumbnail_url, status)
VALUES ($1, $2, $3, $4, $5, $6, 'draft')
RETURNING *;

-- name: GetCourseByID :one
SELECT * FROM courses WHERE id = $1;

-- name: GetCourseBySlug :one
SELECT * FROM courses WHERE tenant_id = $1 AND slug = $2;

-- name: ListCoursesByTenant :many
SELECT
    c.*,
    u.name as author_name,
    u.avatar_url as author_avatar
FROM courses c
JOIN users u ON c.author_id = u.id
WHERE c.tenant_id = $1 AND c.status = 'published'
ORDER BY c.published_at DESC
LIMIT $2 OFFSET $3;

-- name: ListAllCoursesByTenant :many
SELECT
    c.*,
    u.name as author_name,
    u.avatar_url as author_avatar
FROM courses c
JOIN users u ON c.author_id = u.id
WHERE c.tenant_id = $1
ORDER BY c.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListCoursesByAuthor :many
SELECT * FROM courses
WHERE tenant_id = $1 AND author_id = $2
ORDER BY created_at DESC;

-- name: UpdateCourse :one
UPDATE courses
SET title = $2, description = $3, thumbnail_url = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateCourseSlug :one
UPDATE courses
SET slug = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: PublishCourse :one
UPDATE courses
SET status = 'published', published_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UnpublishCourse :one
UPDATE courses
SET status = 'draft', updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ArchiveCourse :exec
UPDATE courses SET status = 'archived', updated_at = NOW() WHERE id = $1;

-- name: DeleteCourse :exec
DELETE FROM courses WHERE id = $1;

-- name: GetCourseWithDetails :one
SELECT
    c.*,
    u.name as author_name,
    u.avatar_url as author_avatar
FROM courses c
JOIN users u ON c.author_id = u.id
WHERE c.id = $1;

-- name: CountCoursesByTenant :one
SELECT COUNT(*) as count FROM courses WHERE tenant_id = $1 AND status = 'published';

-- name: CountAllCoursesByTenant :one
SELECT COUNT(*) as count FROM courses WHERE tenant_id = $1;
