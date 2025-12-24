-- +goose Up
-- ============================================================================
-- ORBIT BACKEND - Likes Schema
-- Track user likes on posts and comments
-- ============================================================================

-- Likes table (unified for posts and comments)
CREATE TABLE likes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Likeable target (one must be set)
    post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    comment_id UUID REFERENCES comments(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure only one of post_id or comment_id is set
    CONSTRAINT likes_target_check CHECK (
        (post_id IS NOT NULL AND comment_id IS NULL) OR
        (post_id IS NULL AND comment_id IS NOT NULL)
    ),

    -- Unique constraint per user per target
    CONSTRAINT likes_unique_post UNIQUE (user_id, post_id),
    CONSTRAINT likes_unique_comment UNIQUE (user_id, comment_id)
);

-- Indexes
CREATE INDEX idx_likes_tenant ON likes(tenant_id);
CREATE INDEX idx_likes_user ON likes(user_id);
CREATE INDEX idx_likes_post ON likes(post_id) WHERE post_id IS NOT NULL;
CREATE INDEX idx_likes_comment ON likes(comment_id) WHERE comment_id IS NOT NULL;

-- ============================================================================
-- FUNCTIONS: Auto-update like counts
-- ============================================================================

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_like_counts()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- Increment like_count
        IF NEW.post_id IS NOT NULL THEN
            UPDATE posts SET like_count = like_count + 1 WHERE id = NEW.post_id;
        ELSIF NEW.comment_id IS NOT NULL THEN
            UPDATE comments SET like_count = like_count + 1 WHERE id = NEW.comment_id;
        END IF;

    ELSIF TG_OP = 'DELETE' THEN
        -- Decrement like_count
        IF OLD.post_id IS NOT NULL THEN
            UPDATE posts SET like_count = like_count - 1 WHERE id = OLD.post_id;
        ELSIF OLD.comment_id IS NOT NULL THEN
            UPDATE comments SET like_count = like_count - 1 WHERE id = OLD.comment_id;
        END IF;
    END IF;

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER update_like_counts_trigger
    AFTER INSERT OR DELETE ON likes
    FOR EACH ROW
    EXECUTE FUNCTION update_like_counts();
