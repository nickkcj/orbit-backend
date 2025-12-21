package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nickkcj/orbit-backend/internal/database"
)

var (
	ErrAlreadyLiked = errors.New("already liked")
	ErrNotLiked     = errors.New("not liked")
)

type LikeService struct {
	db *database.Queries
}

func NewLikeService(db *database.Queries) *LikeService {
	return &LikeService{db: db}
}

// LikePost adds a like to a post
func (s *LikeService) LikePost(ctx context.Context, tenantID, userID, postID uuid.UUID) error {
	// Check if already liked
	_, err := s.db.GetPostLike(ctx, database.GetPostLikeParams{
		UserID: userID,
		PostID: uuid.NullUUID{UUID: postID, Valid: true},
	})
	if err == nil {
		return ErrAlreadyLiked
	}

	_, err = s.db.CreatePostLike(ctx, database.CreatePostLikeParams{
		TenantID: tenantID,
		UserID:   userID,
		PostID:   uuid.NullUUID{UUID: postID, Valid: true},
	})
	return err
}

// UnlikePost removes a like from a post
func (s *LikeService) UnlikePost(ctx context.Context, tenantID, userID, postID uuid.UUID) error {
	// Check if liked
	_, err := s.db.GetPostLike(ctx, database.GetPostLikeParams{
		UserID: userID,
		PostID: uuid.NullUUID{UUID: postID, Valid: true},
	})
	if err != nil {
		return ErrNotLiked
	}

	return s.db.DeletePostLike(ctx, database.DeletePostLikeParams{
		TenantID: tenantID,
		UserID:   userID,
		PostID:   uuid.NullUUID{UUID: postID, Valid: true},
	})
}

// HasUserLikedPost checks if user liked a post
func (s *LikeService) HasUserLikedPost(ctx context.Context, userID, postID uuid.UUID) (bool, error) {
	_, err := s.db.GetPostLike(ctx, database.GetPostLikeParams{
		UserID: userID,
		PostID: uuid.NullUUID{UUID: postID, Valid: true},
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

// LikeComment adds a like to a comment
func (s *LikeService) LikeComment(ctx context.Context, tenantID, userID, commentID uuid.UUID) error {
	// Check if already liked
	_, err := s.db.GetCommentLike(ctx, database.GetCommentLikeParams{
		UserID:    userID,
		CommentID: uuid.NullUUID{UUID: commentID, Valid: true},
	})
	if err == nil {
		return ErrAlreadyLiked
	}

	_, err = s.db.CreateCommentLike(ctx, database.CreateCommentLikeParams{
		TenantID:  tenantID,
		UserID:    userID,
		CommentID: uuid.NullUUID{UUID: commentID, Valid: true},
	})
	return err
}

// UnlikeComment removes a like from a comment
func (s *LikeService) UnlikeComment(ctx context.Context, tenantID, userID, commentID uuid.UUID) error {
	// Check if liked
	_, err := s.db.GetCommentLike(ctx, database.GetCommentLikeParams{
		UserID:    userID,
		CommentID: uuid.NullUUID{UUID: commentID, Valid: true},
	})
	if err != nil {
		return ErrNotLiked
	}

	return s.db.DeleteCommentLike(ctx, database.DeleteCommentLikeParams{
		TenantID:  tenantID,
		UserID:    userID,
		CommentID: uuid.NullUUID{UUID: commentID, Valid: true},
	})
}

// HasUserLikedComment checks if user liked a comment
func (s *LikeService) HasUserLikedComment(ctx context.Context, userID, commentID uuid.UUID) (bool, error) {
	_, err := s.db.GetCommentLike(ctx, database.GetCommentLikeParams{
		UserID:    userID,
		CommentID: uuid.NullUUID{UUID: commentID, Valid: true},
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

// GetUserLikedPostIDs returns post IDs that user has liked
func (s *LikeService) GetUserLikedPostIDs(ctx context.Context, tenantID, userID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := s.db.GetUserPostLikes(ctx, database.GetUserPostLikesParams{
		UserID:   userID,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		if row.Valid {
			result = append(result, row.UUID)
		}
	}
	return result, nil
}

// GetUserLikedCommentIDs returns comment IDs that user has liked
func (s *LikeService) GetUserLikedCommentIDs(ctx context.Context, tenantID, userID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := s.db.GetUserCommentLikes(ctx, database.GetUserCommentLikesParams{
		UserID:   userID,
		TenantID: tenantID,
	})
	if err != nil {
		return nil, err
	}

	result := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		if row.Valid {
			result = append(result, row.UUID)
		}
	}
	return result, nil
}
