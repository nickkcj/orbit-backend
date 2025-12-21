-- name: CreatePostLike :one
INSERT INTO likes (tenant_id, user_id, post_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeletePostLike :exec
DELETE FROM likes
WHERE tenant_id = $1 AND user_id = $2 AND post_id = $3;

-- name: GetPostLike :one
SELECT * FROM likes
WHERE user_id = $1 AND post_id = $2;

-- name: CreateCommentLike :one
INSERT INTO likes (tenant_id, user_id, comment_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: DeleteCommentLike :exec
DELETE FROM likes
WHERE tenant_id = $1 AND user_id = $2 AND comment_id = $3;

-- name: GetCommentLike :one
SELECT * FROM likes
WHERE user_id = $1 AND comment_id = $2;

-- name: GetUserPostLikes :many
SELECT post_id FROM likes
WHERE user_id = $1 AND tenant_id = $2 AND post_id IS NOT NULL;

-- name: GetUserCommentLikes :many
SELECT comment_id FROM likes
WHERE user_id = $1 AND tenant_id = $2 AND comment_id IS NOT NULL;

-- name: GetPostLikesCount :one
SELECT COUNT(*) as count FROM likes
WHERE post_id = $1;

-- name: GetCommentLikesCount :one
SELECT COUNT(*) as count FROM likes
WHERE comment_id = $1;

-- name: CheckUserLikedPosts :many
SELECT post_id FROM likes
WHERE user_id = $1 AND post_id = ANY($2::uuid[]);

-- name: CheckUserLikedComments :many
SELECT comment_id FROM likes
WHERE user_id = $1 AND comment_id = ANY($2::uuid[]);
