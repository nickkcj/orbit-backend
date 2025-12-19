package service

import (
	"context"
	"database/sql"

	"github.com/nickkcj/orbit-backend/internal/database"
)

type TenantService struct {
	db *database.Queries
}

func NewTenantService(db *database.Queries) *TenantService {
	return &TenantService{db: db}
}

func (s *TenantService) GetBySlug(ctx context.Context, slug string) (database.Tenant, error) {
	return s.db.GetTenantBySlug(ctx, slug)
}

func (s *TenantService) Create(ctx context.Context, slug, name, description string) (database.Tenant, error) {
	return s.db.CreateTenant(ctx, database.CreateTenantParams{
		Slug:        slug,
		Name:        name,
		Description: sql.NullString{String: description, Valid: description != ""},
	})
}

func (s *TenantService) List(ctx context.Context) ([]database.Tenant, error) {
	return s.db.ListTenants(ctx)
}
