-- name: CreateUser :one
INSERT INTO users (email, password_hash, name, avatar_url)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users
SET name = $2, avatar_url = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1;

-- name: VerifyUserEmail :exec
UPDATE users SET email_verified_at = NOW(), updated_at = NOW() WHERE id = $1;

-- name: DeleteUser :exec
UPDATE users SET status = 'deleted', updated_at = NOW() WHERE id = $1;
