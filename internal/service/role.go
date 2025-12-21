package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/nickkcj/orbit-backend/internal/database"
)

type RoleService struct {
	db *database.Queries
}

func NewRoleService(db *database.Queries) *RoleService {
	return &RoleService{db: db}
}

// CreateDefaultRoles creates the default roles for a new tenant
// Returns the owner role (highest priority)
func (s *RoleService) CreateDefaultRoles(ctx context.Context, tenantID uuid.UUID) (*database.Role, error) {
	// Create Owner role (highest priority)
	ownerRole, err := s.db.CreateRole(ctx, database.CreateRoleParams{
		TenantID:    tenantID,
		Slug:        "owner",
		Name:        "Owner",
		Description: sql.NullString{String: "Community owner with full access", Valid: true},
		Priority:    100,
		IsDefault:   false,
	})
	if err != nil {
		return nil, err
	}

	// Create Admin role
	_, err = s.db.CreateRole(ctx, database.CreateRoleParams{
		TenantID:    tenantID,
		Slug:        "admin",
		Name:        "Admin",
		Description: sql.NullString{String: "Administrator with management access", Valid: true},
		Priority:    50,
		IsDefault:   false,
	})
	if err != nil {
		return &ownerRole, err
	}

	// Create Member role (default for new members)
	_, err = s.db.CreateRole(ctx, database.CreateRoleParams{
		TenantID:    tenantID,
		Slug:        "member",
		Name:        "Member",
		Description: sql.NullString{String: "Regular community member", Valid: true},
		Priority:    10,
		IsDefault:   true,
	})
	if err != nil {
		return &ownerRole, err
	}

	return &ownerRole, nil
}

func (s *RoleService) GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (database.Role, error) {
	return s.db.GetRoleBySlug(ctx, database.GetRoleBySlugParams{
		TenantID: tenantID,
		Slug:     slug,
	})
}

func (s *RoleService) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]database.Role, error) {
	return s.db.ListRolesByTenant(ctx, tenantID)
}
