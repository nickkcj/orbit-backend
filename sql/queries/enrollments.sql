-- ============================================================================
-- ENROLLMENTS
-- ============================================================================

-- name: CreateEnrollment :one
INSERT INTO course_enrollments (tenant_id, user_id, course_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetEnrollmentByID :one
SELECT * FROM course_enrollments WHERE id = $1;

-- name: GetEnrollmentByUserAndCourse :one
SELECT * FROM course_enrollments
WHERE user_id = $1 AND course_id = $2;

-- name: IsUserEnrolled :one
SELECT EXISTS (
    SELECT 1 FROM course_enrollments
    WHERE user_id = $1 AND course_id = $2 AND status IN ('active', 'completed')
) as enrolled;

-- name: ListEnrollmentsByUser :many
SELECT
    e.*,
    c.title as course_title,
    c.slug as course_slug,
    c.thumbnail_url as course_thumbnail,
    c.lesson_count as course_total_lessons,
    u.name as author_name,
    u.avatar_url as author_avatar
FROM course_enrollments e
JOIN courses c ON e.course_id = c.id
JOIN users u ON c.author_id = u.id
WHERE e.user_id = $1 AND e.tenant_id = $2 AND e.status = 'active'
ORDER BY e.last_accessed_at DESC NULLS LAST, e.enrolled_at DESC
LIMIT $3 OFFSET $4;

-- name: CountEnrollmentsByUser :one
SELECT COUNT(*) as count FROM course_enrollments
WHERE user_id = $1 AND tenant_id = $2 AND status = 'active';

-- name: ListEnrollmentsByCourse :many
SELECT
    e.*,
    u.name as user_name,
    u.email as user_email,
    u.avatar_url as user_avatar
FROM course_enrollments e
JOIN users u ON e.user_id = u.id
WHERE e.course_id = $1
ORDER BY e.enrolled_at DESC
LIMIT $2 OFFSET $3;

-- name: CountEnrollmentsByCourse :one
SELECT COUNT(*) as count FROM course_enrollments WHERE course_id = $1;

-- name: GetContinueLearningCourses :many
SELECT
    e.*,
    c.title as course_title,
    c.slug as course_slug,
    c.thumbnail_url as course_thumbnail,
    c.lesson_count as course_total_lessons,
    c.module_count as course_module_count,
    u.name as author_name,
    u.avatar_url as author_avatar,
    l.id as lesson_id,
    l.title as lesson_title,
    m.title as last_module_title
FROM course_enrollments e
JOIN courses c ON e.course_id = c.id
JOIN users u ON c.author_id = u.id
LEFT JOIN lessons l ON e.last_lesson_id = l.id
LEFT JOIN modules m ON l.module_id = m.id
WHERE e.user_id = $1 AND e.tenant_id = $2 AND e.status = 'active' AND e.progress_percentage < 100
ORDER BY e.last_accessed_at DESC NULLS LAST, e.enrolled_at DESC
LIMIT $3;

-- name: GetCompletedCourses :many
SELECT
    e.*,
    c.title as course_title,
    c.slug as course_slug,
    c.thumbnail_url as course_thumbnail,
    u.name as author_name
FROM course_enrollments e
JOIN courses c ON e.course_id = c.id
JOIN users u ON c.author_id = u.id
WHERE e.user_id = $1 AND e.tenant_id = $2 AND e.status = 'completed'
ORDER BY e.completed_at DESC
LIMIT $3 OFFSET $4;

-- name: UpdateEnrollmentLastAccessed :exec
UPDATE course_enrollments
SET last_accessed_at = NOW(), last_lesson_id = $2, updated_at = NOW()
WHERE id = $1;

-- name: DropEnrollment :exec
UPDATE course_enrollments
SET status = 'dropped', updated_at = NOW()
WHERE id = $1;

-- name: DeleteEnrollment :exec
DELETE FROM course_enrollments WHERE id = $1;

-- name: GetEnrollmentWithProgress :one
SELECT
    e.*,
    c.title as course_title,
    c.slug as course_slug,
    c.thumbnail_url as course_thumbnail,
    c.lesson_count as course_total_lessons,
    c.module_count as course_module_count,
    c.description as course_description,
    u.name as author_name,
    u.avatar_url as author_avatar
FROM course_enrollments e
JOIN courses c ON e.course_id = c.id
JOIN users u ON c.author_id = u.id
WHERE e.id = $1;

-- name: GetEnrollmentWithProgressByUserAndCourse :one
SELECT
    e.*,
    c.title as course_title,
    c.slug as course_slug,
    c.thumbnail_url as course_thumbnail,
    c.lesson_count as course_total_lessons,
    c.module_count as course_module_count,
    c.description as course_description,
    u.name as author_name,
    u.avatar_url as author_avatar
FROM course_enrollments e
JOIN courses c ON e.course_id = c.id
JOIN users u ON c.author_id = u.id
WHERE e.user_id = $1 AND e.course_id = $2;
