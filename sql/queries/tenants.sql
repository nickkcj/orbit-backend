-- name: CreateTenant :one
INSERT INTO tenants (slug, name, description)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetTenantByID :one
SELECT * FROM tenants WHERE id = $1;

-- name: GetTenantBySlug :one
SELECT * FROM tenants WHERE slug = $1;

-- name: ListTenants :many
SELECT * FROM tenants
WHERE status = 'active'
ORDER BY created_at DESC;

-- name: UpdateTenant :one
UPDATE tenants
SET name = $2, description = $3, logo_url = $4, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateTenantBilling :one
UPDATE tenants
SET billing_status = $2, stripe_customer_id = $3, stripe_subscription_id = $4, plan_id = $5, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTenant :exec
UPDATE tenants SET status = 'deleted', updated_at = NOW() WHERE id = $1;

-- name: UpdateTenantSettings :one
UPDATE tenants
SET settings = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateTenantLogo :one
UPDATE tenants
SET logo_url = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;
