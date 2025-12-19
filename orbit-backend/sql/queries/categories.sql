-- name: CreateCategory :one
INSERT INTO categories (tenant_id, slug, name, description, icon, position)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetCategoryByID :one
SELECT * FROM categories WHERE id = $1;

-- name: GetCategoryBySlug :one
SELECT * FROM categories WHERE tenant_id = $1 AND slug = $2;

-- name: ListCategoriesByTenant :many
SELECT * FROM categories
WHERE tenant_id = $1 AND is_visible = true
ORDER BY position ASC;

-- name: UpdateCategory :one
UPDATE categories
SET name = $2, description = $3, icon = $4, position = $5, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = $1;

-- name: HideCategory :exec
UPDATE categories SET is_visible = false, updated_at = NOW() WHERE id = $1;
