package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/nickkcj/orbit-backend/internal/database"
)

type CommentService struct {
	db *database.Queries
}

func NewCommentService(db *database.Queries) *CommentService {
	return &CommentService{db: db}
}

type CreateCommentInput struct {
	TenantID uuid.UUID
	PostID   uuid.UUID
	AuthorID uuid.UUID
	ParentID *uuid.UUID
	Content  string
}

func (s *CommentService) Create(ctx context.Context, input CreateCommentInput) (database.Comment, error) {
	parentID := uuid.NullUUID{}
	if input.ParentID != nil {
		parentID = uuid.NullUUID{UUID: *input.ParentID, Valid: true}
	}

	return s.db.CreateComment(ctx, database.CreateCommentParams{
		TenantID: input.TenantID,
		PostID:   input.PostID,
		AuthorID: input.AuthorID,
		ParentID: parentID,
		Content:  input.Content,
	})
}

func (s *CommentService) GetByID(ctx context.Context, id uuid.UUID) (database.Comment, error) {
	return s.db.GetCommentByID(ctx, id)
}

func (s *CommentService) ListByPost(ctx context.Context, postID uuid.UUID, limit, offset int32) ([]database.ListCommentsByPostRow, error) {
	return s.db.ListCommentsByPost(ctx, database.ListCommentsByPostParams{
		PostID: postID,
		Limit:  limit,
		Offset: offset,
	})
}

func (s *CommentService) ListReplies(ctx context.Context, parentID uuid.UUID) ([]database.ListRepliesRow, error) {
	return s.db.ListReplies(ctx, uuid.NullUUID{UUID: parentID, Valid: true})
}

func (s *CommentService) Update(ctx context.Context, id uuid.UUID, content string) (database.Comment, error) {
	return s.db.UpdateComment(ctx, database.UpdateCommentParams{
		ID:      id,
		Content: content,
	})
}

func (s *CommentService) Hide(ctx context.Context, id uuid.UUID) error {
	return s.db.HideComment(ctx, id)
}

func (s *CommentService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.DeleteComment(ctx, id)
}
