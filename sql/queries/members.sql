-- name: AddMember :one
INSERT INTO tenant_members (tenant_id, user_id, role_id, display_name)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetMember :one
SELECT * FROM tenant_members WHERE tenant_id = $1 AND user_id = $2;

-- name: GetMemberWithRole :one
SELECT
    tm.*,
    r.slug as role_slug,
    r.name as role_name,
    r.priority as role_priority
FROM tenant_members tm
JOIN roles r ON tm.role_id = r.id
WHERE tm.tenant_id = $1 AND tm.user_id = $2;

-- name: ListMembersByTenant :many
SELECT
    tm.*,
    u.email,
    u.name as user_name,
    u.avatar_url as user_avatar,
    r.slug as role_slug,
    r.name as role_name
FROM tenant_members tm
JOIN users u ON tm.user_id = u.id
JOIN roles r ON tm.role_id = r.id
WHERE tm.tenant_id = $1 AND tm.status = 'active'
ORDER BY tm.joined_at DESC;

-- name: ListTenantsByUser :many
SELECT
    t.*,
    tm.role_id,
    tm.display_name,
    tm.joined_at,
    r.slug as role_slug
FROM tenants t
JOIN tenant_members tm ON t.id = tm.tenant_id
JOIN roles r ON tm.role_id = r.id
WHERE tm.user_id = $1 AND tm.status = 'active' AND t.status = 'active'
ORDER BY tm.joined_at DESC;

-- name: UpdateMemberRole :one
UPDATE tenant_members
SET role_id = $3, updated_at = NOW()
WHERE tenant_id = $1 AND user_id = $2
RETURNING *;

-- name: UpdateMemberStatus :exec
UPDATE tenant_members
SET status = $3, updated_at = NOW()
WHERE tenant_id = $1 AND user_id = $2;

-- name: RemoveMember :exec
DELETE FROM tenant_members WHERE tenant_id = $1 AND user_id = $2;

-- name: CountMembersByTenant :one
SELECT COUNT(*) FROM tenant_members WHERE tenant_id = $1 AND status = 'active';
