package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/gosimple/slug"

	"github.com/nickkcj/orbit-backend/internal/database"
)

type CategoryService struct {
	db *database.Queries
}

func NewCategoryService(db *database.Queries) *CategoryService {
	return &CategoryService{db: db}
}

type CreateCategoryInput struct {
	TenantID    uuid.UUID
	Name        string
	Description string
	Icon        string
	Position    int32
}

func (s *CategoryService) Create(ctx context.Context, input CreateCategoryInput) (database.Category, error) {
	categorySlug := slug.Make(input.Name)

	return s.db.CreateCategory(ctx, database.CreateCategoryParams{
		TenantID:    input.TenantID,
		Slug:        categorySlug,
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Icon:        sql.NullString{String: input.Icon, Valid: input.Icon != ""},
		Position:    input.Position,
	})
}

func (s *CategoryService) GetByID(ctx context.Context, id uuid.UUID) (database.Category, error) {
	return s.db.GetCategoryByID(ctx, id)
}

func (s *CategoryService) GetBySlug(ctx context.Context, tenantID uuid.UUID, categorySlug string) (database.Category, error) {
	return s.db.GetCategoryBySlug(ctx, database.GetCategoryBySlugParams{
		TenantID: tenantID,
		Slug:     categorySlug,
	})
}

func (s *CategoryService) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]database.Category, error) {
	return s.db.ListCategoriesByTenant(ctx, tenantID)
}

type UpdateCategoryInput struct {
	Name        string
	Description string
	Icon        string
	Position    int32
}

func (s *CategoryService) Update(ctx context.Context, id uuid.UUID, input UpdateCategoryInput) (database.Category, error) {
	return s.db.UpdateCategory(ctx, database.UpdateCategoryParams{
		ID:          id,
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Icon:        sql.NullString{String: input.Icon, Valid: input.Icon != ""},
		Position:    input.Position,
	})
}

func (s *CategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.DeleteCategory(ctx, id)
}

func (s *CategoryService) Hide(ctx context.Context, id uuid.UUID) error {
	return s.db.HideCategory(ctx, id)
}
