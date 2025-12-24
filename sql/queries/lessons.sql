-- ============================================================================
-- LESSONS
-- ============================================================================

-- name: CreateLesson :one
INSERT INTO lessons (tenant_id, module_id, title, description, content, content_format, video_id, position, duration_minutes, is_free_preview)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetLessonByID :one
SELECT * FROM lessons WHERE id = $1;

-- name: GetLessonWithVideo :one
SELECT
    l.*,
    v.title as video_title,
    v.playback_url as video_playback_url,
    v.thumbnail_url as video_thumbnail_url,
    v.duration_seconds as video_duration,
    v.status as video_status
FROM lessons l
LEFT JOIN videos v ON l.video_id = v.id
WHERE l.id = $1;

-- name: ListLessonsByModule :many
SELECT
    l.*,
    v.title as video_title,
    v.thumbnail_url as video_thumbnail_url,
    v.duration_seconds as video_duration,
    v.status as video_status
FROM lessons l
LEFT JOIN videos v ON l.video_id = v.id
WHERE l.module_id = $1
ORDER BY l.position ASC;

-- name: ListLessonsByCourse :many
SELECT
    l.*,
    m.title as module_title,
    m.position as module_position,
    v.thumbnail_url as video_thumbnail_url,
    v.duration_seconds as video_duration
FROM lessons l
JOIN modules m ON l.module_id = m.id
LEFT JOIN videos v ON l.video_id = v.id
WHERE m.course_id = $1
ORDER BY m.position ASC, l.position ASC;

-- name: UpdateLesson :one
UPDATE lessons
SET title = $2, description = $3, content = $4, content_format = $5, video_id = $6, duration_minutes = $7, is_free_preview = $8, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateLessonPosition :exec
UPDATE lessons SET position = $2, updated_at = NOW() WHERE id = $1;

-- name: DeleteLesson :exec
DELETE FROM lessons WHERE id = $1;

-- name: GetMaxLessonPosition :one
SELECT COALESCE(MAX(position), -1)::int as max_position FROM lessons WHERE module_id = $1;

-- name: ReorderLessonsAfterDelete :exec
UPDATE lessons
SET position = position - 1, updated_at = NOW()
WHERE module_id = $1 AND position > $2;

-- name: ShiftLessonPositionsUp :exec
UPDATE lessons
SET position = position + 1, updated_at = NOW()
WHERE module_id = $1 AND position >= $2;

-- name: ShiftLessonPositionsDown :exec
UPDATE lessons
SET position = position - 1, updated_at = NOW()
WHERE module_id = $1 AND position > $2 AND position <= $3;

-- name: AttachVideoToLesson :exec
UPDATE lessons SET video_id = $2, updated_at = NOW() WHERE id = $1;

-- name: DetachVideoFromLesson :exec
UPDATE lessons SET video_id = NULL, updated_at = NOW() WHERE id = $1;

-- name: CountLessonsByModule :one
SELECT COUNT(*) as count FROM lessons WHERE module_id = $1;

-- name: CountLessonsByCourse :one
SELECT COUNT(*) as count
FROM lessons l
JOIN modules m ON l.module_id = m.id
WHERE m.course_id = $1;
