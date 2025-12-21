-- name: GetTenantStats :one
SELECT
    (SELECT COUNT(*) FROM tenant_members tm WHERE tm.tenant_id = $1 AND tm.status = 'active') as total_members,
    (SELECT COUNT(*) FROM posts p WHERE p.tenant_id = $1 AND p.status = 'published') as total_posts,
    (SELECT COUNT(*) FROM comments c JOIN posts p ON c.post_id = p.id WHERE p.tenant_id = $1) as total_comments,
    (SELECT COALESCE(SUM(p2.view_count), 0) FROM posts p2 WHERE p2.tenant_id = $1) as total_views;

-- name: GetMembersGrowth :many
SELECT
    DATE(joined_at) as date,
    COUNT(*) as count
FROM tenant_members
WHERE tenant_id = $1
    AND joined_at >= $2
    AND status = 'active'
GROUP BY DATE(joined_at)
ORDER BY date;

-- name: GetPostsGrowth :many
SELECT
    DATE(created_at) as date,
    COUNT(*) as count
FROM posts
WHERE tenant_id = $1
    AND created_at >= $2
    AND status = 'published'
GROUP BY DATE(created_at)
ORDER BY date;

-- name: GetTopPosts :many
SELECT
    p.id,
    p.title,
    p.view_count,
    p.like_count,
    (SELECT COUNT(*) FROM comments WHERE post_id = p.id) as comment_count,
    p.created_at
FROM posts p
WHERE p.tenant_id = $1 AND p.status = 'published'
ORDER BY p.view_count DESC
LIMIT $2;

-- name: GetRecentMembers :many
SELECT
    tm.user_id,
    tm.display_name,
    u.name as user_name,
    u.avatar_url,
    tm.joined_at
FROM tenant_members tm
JOIN users u ON tm.user_id = u.id
WHERE tm.tenant_id = $1 AND tm.status = 'active'
ORDER BY tm.joined_at DESC
LIMIT $2;
