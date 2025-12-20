package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/gosimple/slug"

	"github.com/nickkcj/orbit-backend/internal/database"
)

type PostService struct {
	db *database.Queries
}

func NewPostService(db *database.Queries) *PostService {
	return &PostService{db: db}
}

type CreatePostInput struct {
	TenantID      uuid.UUID
	AuthorID      uuid.UUID
	CategoryID    *uuid.UUID
	Title         string
	Content       string
	ContentFormat string
	Excerpt       string
	CoverImageURL string
}

func (s *PostService) Create(ctx context.Context, input CreatePostInput) (database.Post, error) {
	postSlug := slug.Make(input.Title)

	categoryID := uuid.NullUUID{}
	if input.CategoryID != nil {
		categoryID = uuid.NullUUID{UUID: *input.CategoryID, Valid: true}
	}

	contentFormat := input.ContentFormat
	if contentFormat == "" {
		contentFormat = "markdown"
	}

	return s.db.CreatePost(ctx, database.CreatePostParams{
		TenantID:      input.TenantID,
		AuthorID:      input.AuthorID,
		CategoryID:    categoryID,
		Title:         input.Title,
		Slug:          postSlug,
		Content:       sql.NullString{String: input.Content, Valid: input.Content != ""},
		ContentFormat: contentFormat,
		Excerpt:       sql.NullString{String: input.Excerpt, Valid: input.Excerpt != ""},
		CoverImageUrl: sql.NullString{String: input.CoverImageURL, Valid: input.CoverImageURL != ""},
		Status:        "draft",
	})
}

func (s *PostService) GetByID(ctx context.Context, id uuid.UUID) (database.Post, error) {
	return s.db.GetPostByID(ctx, id)
}

func (s *PostService) GetBySlug(ctx context.Context, tenantID uuid.UUID, postSlug string) (database.Post, error) {
	return s.db.GetPostBySlug(ctx, database.GetPostBySlugParams{
		TenantID: tenantID,
		Slug:     postSlug,
	})
}

func (s *PostService) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]database.ListPostsByTenantRow, error) {
	return s.db.ListPostsByTenant(ctx, database.ListPostsByTenantParams{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	})
}

type UpdatePostInput struct {
	Title         string
	Content       string
	Excerpt       string
	CoverImageURL string
	CategoryID    *uuid.UUID
}

func (s *PostService) Update(ctx context.Context, id uuid.UUID, input UpdatePostInput) (database.Post, error) {
	categoryID := uuid.NullUUID{}
	if input.CategoryID != nil {
		categoryID = uuid.NullUUID{UUID: *input.CategoryID, Valid: true}
	}

	return s.db.UpdatePost(ctx, database.UpdatePostParams{
		ID:            id,
		Title:         input.Title,
		Content:       sql.NullString{String: input.Content, Valid: input.Content != ""},
		Excerpt:       sql.NullString{String: input.Excerpt, Valid: input.Excerpt != ""},
		CoverImageUrl: sql.NullString{String: input.CoverImageURL, Valid: input.CoverImageURL != ""},
		CategoryID:    categoryID,
	})
}

func (s *PostService) Publish(ctx context.Context, id uuid.UUID) (database.Post, error) {
	return s.db.PublishPost(ctx, id)
}

func (s *PostService) Archive(ctx context.Context, id uuid.UUID) error {
	return s.db.ArchivePost(ctx, id)
}

func (s *PostService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.DeletePost(ctx, id)
}

func (s *PostService) IncrementViews(ctx context.Context, id uuid.UUID) error {
	return s.db.IncrementPostViews(ctx, id)
}

func (s *PostService) ListByAuthor(ctx context.Context, tenantID, authorID uuid.UUID) ([]database.Post, error) {
	return s.db.ListPostsByAuthor(ctx, database.ListPostsByAuthorParams{
		TenantID: tenantID,
		AuthorID: authorID,
	})
}
