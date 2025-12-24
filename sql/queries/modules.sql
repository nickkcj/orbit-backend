-- ============================================================================
-- MODULES
-- ============================================================================

-- name: CreateModule :one
INSERT INTO modules (tenant_id, course_id, title, description, position)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetModuleByID :one
SELECT * FROM modules WHERE id = $1;

-- name: ListModulesByCourse :many
SELECT * FROM modules
WHERE course_id = $1
ORDER BY position ASC;

-- name: UpdateModule :one
UPDATE modules
SET title = $2, description = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateModulePosition :exec
UPDATE modules SET position = $2, updated_at = NOW() WHERE id = $1;

-- name: DeleteModule :exec
DELETE FROM modules WHERE id = $1;

-- name: GetMaxModulePosition :one
SELECT COALESCE(MAX(position), -1)::int as max_position FROM modules WHERE course_id = $1;

-- name: ReorderModulesAfterDelete :exec
UPDATE modules
SET position = position - 1, updated_at = NOW()
WHERE course_id = $1 AND position > $2;

-- name: ShiftModulePositionsUp :exec
UPDATE modules
SET position = position + 1, updated_at = NOW()
WHERE course_id = $1 AND position >= $2;

-- name: ShiftModulePositionsDown :exec
UPDATE modules
SET position = position - 1, updated_at = NOW()
WHERE course_id = $1 AND position > $2 AND position <= $3;

-- name: CountModulesByCourse :one
SELECT COUNT(*) as count FROM modules WHERE course_id = $1;
