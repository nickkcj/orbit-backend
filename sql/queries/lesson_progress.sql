-- ============================================================================
-- LESSON PROGRESS
-- ============================================================================

-- name: CreateLessonProgress :one
INSERT INTO lesson_progress (tenant_id, enrollment_id, lesson_id, status, started_at)
VALUES ($1, $2, $3, 'in_progress', NOW())
RETURNING *;

-- name: GetLessonProgress :one
SELECT * FROM lesson_progress
WHERE enrollment_id = $1 AND lesson_id = $2;

-- name: GetLessonProgressByID :one
SELECT * FROM lesson_progress WHERE id = $1;

-- name: ListLessonProgressByEnrollment :many
SELECT * FROM lesson_progress
WHERE enrollment_id = $1
ORDER BY updated_at DESC;

-- name: MarkLessonComplete :one
INSERT INTO lesson_progress (tenant_id, enrollment_id, lesson_id, status, started_at, completed_at)
VALUES ($1, $2, $3, 'completed', COALESCE($4, NOW()), NOW())
ON CONFLICT (enrollment_id, lesson_id) DO UPDATE
SET status = 'completed', completed_at = NOW(), updated_at = NOW()
RETURNING *;

-- name: MarkLessonIncomplete :one
UPDATE lesson_progress
SET status = 'in_progress', completed_at = NULL, updated_at = NOW()
WHERE enrollment_id = $1 AND lesson_id = $2
RETURNING *;

-- name: UpdateVideoProgress :one
INSERT INTO lesson_progress (tenant_id, enrollment_id, lesson_id, status, watch_duration_seconds, video_total_seconds, started_at)
VALUES ($1, $2, $3, 'in_progress', $4, $5, NOW())
ON CONFLICT (enrollment_id, lesson_id) DO UPDATE
SET watch_duration_seconds = $4, video_total_seconds = COALESCE($5, lesson_progress.video_total_seconds), updated_at = NOW()
RETURNING *;

-- name: DeleteLessonProgress :exec
DELETE FROM lesson_progress
WHERE enrollment_id = $1 AND lesson_id = $2;

-- name: GetNextIncompleteLesson :one
SELECT
    l.id,
    l.title,
    l.position,
    m.id as module_id,
    m.title as module_title,
    m.position as module_position
FROM lessons l
JOIN modules m ON l.module_id = m.id
WHERE m.course_id = $1
AND l.id NOT IN (
    SELECT lesson_id FROM lesson_progress
    WHERE enrollment_id = $2 AND status = 'completed'
)
ORDER BY m.position ASC, l.position ASC
LIMIT 1;

-- name: GetFirstLesson :one
SELECT
    l.id,
    l.title,
    l.position,
    m.id as module_id,
    m.title as module_title,
    m.position as module_position
FROM lessons l
JOIN modules m ON l.module_id = m.id
WHERE m.course_id = $1
ORDER BY m.position ASC, l.position ASC
LIMIT 1;

-- name: GetLessonWithProgressAndCourse :one
SELECT
    l.*,
    m.id as module_id,
    m.title as module_title,
    m.position as module_position,
    m.course_id as course_id,
    c.title as course_title,
    c.slug as course_slug,
    v.title as video_title,
    v.playback_url as video_playback_url,
    v.thumbnail_url as video_thumbnail_url,
    v.duration_seconds as video_duration,
    v.status as video_status,
    lp.id as progress_id,
    lp.status as progress_status,
    lp.watch_duration_seconds as progress_watch_duration,
    lp.completed_at as progress_completed_at
FROM lessons l
JOIN modules m ON l.module_id = m.id
JOIN courses c ON m.course_id = c.id
LEFT JOIN videos v ON l.video_id = v.id
LEFT JOIN lesson_progress lp ON l.id = lp.lesson_id AND lp.enrollment_id = $2
WHERE l.id = $1;

-- name: ListLessonsWithProgress :many
SELECT
    l.id,
    l.title,
    l.position,
    l.duration_minutes,
    l.is_free_preview,
    l.video_id,
    m.id as module_id,
    m.title as module_title,
    m.position as module_position,
    v.duration_seconds as video_duration,
    v.thumbnail_url as video_thumbnail,
    COALESCE(lp.status, 'not_started') as progress_status,
    lp.completed_at as progress_completed_at
FROM lessons l
JOIN modules m ON l.module_id = m.id
LEFT JOIN videos v ON l.video_id = v.id
LEFT JOIN lesson_progress lp ON l.id = lp.lesson_id AND lp.enrollment_id = $2
WHERE m.course_id = $1
ORDER BY m.position ASC, l.position ASC;

-- name: GetPreviousLesson :one
SELECT
    l.id,
    l.title,
    m.position as module_position,
    l.position as lesson_position
FROM lessons l
JOIN modules m ON l.module_id = m.id
WHERE m.course_id = $1
AND (m.position < $2 OR (m.position = $2 AND l.position < $3))
ORDER BY m.position DESC, l.position DESC
LIMIT 1;

-- name: GetNextLesson :one
SELECT
    l.id,
    l.title,
    m.position as module_position,
    l.position as lesson_position
FROM lessons l
JOIN modules m ON l.module_id = m.id
WHERE m.course_id = $1
AND (m.position > $2 OR (m.position = $2 AND l.position > $3))
ORDER BY m.position ASC, l.position ASC
LIMIT 1;

-- name: CountCompletedLessonsByEnrollment :one
SELECT COUNT(*) as count FROM lesson_progress
WHERE enrollment_id = $1 AND status = 'completed';
