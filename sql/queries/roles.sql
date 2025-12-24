-- name: GetRoleByID :one
SELECT * FROM roles WHERE id = $1;

-- name: GetRoleBySlug :one
SELECT * FROM roles WHERE tenant_id = $1 AND slug = $2;

-- name: GetDefaultRole :one
SELECT * FROM roles WHERE tenant_id = $1 AND is_default = true LIMIT 1;

-- name: ListRolesByTenant :many
SELECT * FROM roles WHERE tenant_id = $1 ORDER BY priority DESC;

-- name: CreateRole :one
INSERT INTO roles (tenant_id, slug, name, description, priority, is_default)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateRole :one
UPDATE roles
SET name = $2, description = $3, priority = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1 AND is_system = false;

-- name: GetRolePermissions :many
SELECT p.* FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
WHERE rp.role_id = $1;

-- name: AddRolePermission :exec
INSERT INTO role_permissions (role_id, permission_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveRolePermission :exec
DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2;

-- name: ListPermissions :many
SELECT * FROM permissions ORDER BY category, code;

-- name: GetMemberPermissions :many
SELECT p.code FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
JOIN tenant_members tm ON rp.role_id = tm.role_id
WHERE tm.tenant_id = $1 AND tm.user_id = $2 AND tm.status = 'active';

-- name: HasPermission :one
SELECT EXISTS (
    SELECT 1
    FROM permissions p
    JOIN role_permissions rp ON p.id = rp.permission_id
    JOIN tenant_members tm ON rp.role_id = tm.role_id
    WHERE tm.tenant_id = $1
    AND tm.user_id = $2
    AND tm.status = 'active'
    AND p.code = $3
) as has_permission;
