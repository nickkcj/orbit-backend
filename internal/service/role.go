package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/nickkcj/orbit-backend/internal/database"
)

type RoleService struct {
	db *database.Queries
}

func NewRoleService(db *database.Queries) *RoleService {
	return &RoleService{db: db}
}

// RoleResponse represents a role with its permissions
type RoleResponse struct {
	ID          uuid.UUID    `json:"id"`
	TenantID    uuid.UUID    `json:"tenant_id"`
	Slug        string       `json:"slug"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Priority    int32        `json:"priority"`
	IsDefault   bool         `json:"is_default"`
	IsSystem    bool         `json:"is_system"`
	Permissions []Permission `json:"permissions,omitempty"`
}

// Permission represents a permission
type Permission struct {
	ID          uuid.UUID `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Category    string    `json:"category"`
}

// CreateDefaultRoles creates the default roles for a new tenant
// Returns the owner role (highest priority)
func (s *RoleService) CreateDefaultRoles(ctx context.Context, tenantID uuid.UUID) (*database.Role, error) {
	// Get all available permissions
	allPermissions, err := s.db.ListPermissions(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	// Build permission maps by category for easier assignment
	permByCode := make(map[string]uuid.UUID)
	for _, p := range allPermissions {
		permByCode[p.Code] = p.ID
	}

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
	adminRole, err := s.db.CreateRole(ctx, database.CreateRoleParams{
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
	memberRole, err := s.db.CreateRole(ctx, database.CreateRoleParams{
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

	// Owner gets ALL permissions
	for _, p := range allPermissions {
		s.db.AddRolePermission(ctx, database.AddRolePermissionParams{
			RoleID:       ownerRole.ID,
			PermissionID: p.ID,
		})
	}

	// Admin gets management permissions (no delete permissions)
	adminPermissions := []string{
		"posts.view", "posts.create", "posts.edit",
		"members.view", "members.invite", "members.manage",
		"courses.view", "courses.create", "courses.edit",
		"enrollments.view", "enrollments.enroll", "enrollments.manage",
		"settings.view", "settings.edit",
		"analytics.view",
	}
	for _, code := range adminPermissions {
		if permID, ok := permByCode[code]; ok {
			s.db.AddRolePermission(ctx, database.AddRolePermissionParams{
				RoleID:       adminRole.ID,
				PermissionID: permID,
			})
		}
	}

	// Member gets basic view and interaction permissions
	memberPermissions := []string{
		"posts.view", "posts.create",
		"members.view",
		"courses.view",
		"enrollments.view", "enrollments.enroll",
	}
	for _, code := range memberPermissions {
		if permID, ok := permByCode[code]; ok {
			s.db.AddRolePermission(ctx, database.AddRolePermissionParams{
				RoleID:       memberRole.ID,
				PermissionID: permID,
			})
		}
	}

	return &ownerRole, nil
}

func (s *RoleService) GetByID(ctx context.Context, id uuid.UUID) (*RoleResponse, error) {
	role, err := s.db.GetRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	permissions, err := s.db.GetRolePermissions(ctx, id)
	if err != nil {
		return nil, err
	}

	perms := make([]Permission, len(permissions))
	for i, p := range permissions {
		perms[i] = Permission{
			ID:          p.ID,
			Code:        p.Code,
			Name:        p.Name,
			Description: p.Description.String,
			Category:    p.Category,
		}
	}

	return &RoleResponse{
		ID:          role.ID,
		TenantID:    role.TenantID,
		Slug:        role.Slug,
		Name:        role.Name,
		Description: role.Description.String,
		Priority:    role.Priority,
		IsDefault:   role.IsDefault,
		IsSystem:    role.IsSystem,
		Permissions: perms,
	}, nil
}

func (s *RoleService) GetBySlug(ctx context.Context, tenantID uuid.UUID, slug string) (database.Role, error) {
	return s.db.GetRoleBySlug(ctx, database.GetRoleBySlugParams{
		TenantID: tenantID,
		Slug:     slug,
	})
}

func (s *RoleService) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]RoleResponse, error) {
	roles, err := s.db.ListRolesByTenant(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	result := make([]RoleResponse, len(roles))
	for i, role := range roles {
		permissions, err := s.db.GetRolePermissions(ctx, role.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get permissions for role %s: %w", role.Slug, err)
		}

		perms := make([]Permission, len(permissions))
		for j, p := range permissions {
			perms[j] = Permission{
				ID:          p.ID,
				Code:        p.Code,
				Name:        p.Name,
				Description: p.Description.String,
				Category:    p.Category,
			}
		}

		result[i] = RoleResponse{
			ID:          role.ID,
			TenantID:    role.TenantID,
			Slug:        role.Slug,
			Name:        role.Name,
			Description: role.Description.String,
			Priority:    role.Priority,
			IsDefault:   role.IsDefault,
			IsSystem:    role.IsSystem,
			Permissions: perms,
		}
	}

	return result, nil
}

// CreateRoleRequest represents a request to create a role
type CreateRoleRequest struct {
	Slug        string   `json:"slug"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Priority    int32    `json:"priority"`
	IsDefault   bool     `json:"is_default"`
	Permissions []string `json:"permissions"` // permission codes
}

func (s *RoleService) Create(ctx context.Context, tenantID uuid.UUID, req *CreateRoleRequest) (*RoleResponse, error) {
	role, err := s.db.CreateRole(ctx, database.CreateRoleParams{
		TenantID:    tenantID,
		Slug:        req.Slug,
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		Priority:    req.Priority,
		IsDefault:   req.IsDefault,
	})
	if err != nil {
		return nil, err
	}

	// Add permissions
	if len(req.Permissions) > 0 {
		if err := s.SetPermissions(ctx, role.ID, req.Permissions); err != nil {
			return nil, err
		}
	}

	return s.GetByID(ctx, role.ID)
}

// UpdateRoleRequest represents a request to update a role
type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    int32  `json:"priority"`
}

func (s *RoleService) Update(ctx context.Context, id uuid.UUID, req *UpdateRoleRequest) (*RoleResponse, error) {
	role, err := s.db.GetRoleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role.IsSystem {
		return nil, fmt.Errorf("cannot modify system role")
	}

	_, err = s.db.UpdateRole(ctx, database.UpdateRoleParams{
		ID:          id,
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		Priority:    req.Priority,
	})
	if err != nil {
		return nil, err
	}

	return s.GetByID(ctx, id)
}

func (s *RoleService) Delete(ctx context.Context, id uuid.UUID) error {
	role, err := s.db.GetRoleByID(ctx, id)
	if err != nil {
		return err
	}

	if role.IsSystem {
		return fmt.Errorf("cannot delete system role")
	}

	return s.db.DeleteRole(ctx, id)
}

// ListAllPermissions returns all available permissions
func (s *RoleService) ListAllPermissions(ctx context.Context) ([]Permission, error) {
	perms, err := s.db.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]Permission, len(perms))
	for i, p := range perms {
		result[i] = Permission{
			ID:          p.ID,
			Code:        p.Code,
			Name:        p.Name,
			Description: p.Description.String,
			Category:    p.Category,
		}
	}

	return result, nil
}

// SetPermissions sets the permissions for a role (replaces all existing)
func (s *RoleService) SetPermissions(ctx context.Context, roleID uuid.UUID, permissionCodes []string) error {
	// Get all permissions to map codes to IDs
	allPerms, err := s.db.ListPermissions(ctx)
	if err != nil {
		return err
	}

	permMap := make(map[string]uuid.UUID)
	for _, p := range allPerms {
		permMap[p.Code] = p.ID
	}

	// Get current permissions to remove
	currentPerms, err := s.db.GetRolePermissions(ctx, roleID)
	if err != nil {
		return err
	}

	// Remove all current permissions
	for _, p := range currentPerms {
		s.db.RemoveRolePermission(ctx, database.RemoveRolePermissionParams{
			RoleID:       roleID,
			PermissionID: p.ID,
		})
	}

	// Add new permissions
	for _, code := range permissionCodes {
		if permID, ok := permMap[code]; ok {
			s.db.AddRolePermission(ctx, database.AddRolePermissionParams{
				RoleID:       roleID,
				PermissionID: permID,
			})
		}
	}

	return nil
}

// AddPermission adds a single permission to a role
func (s *RoleService) AddPermission(ctx context.Context, roleID uuid.UUID, permissionCode string) error {
	perms, err := s.db.ListPermissions(ctx)
	if err != nil {
		return err
	}

	for _, p := range perms {
		if p.Code == permissionCode {
			return s.db.AddRolePermission(ctx, database.AddRolePermissionParams{
				RoleID:       roleID,
				PermissionID: p.ID,
			})
		}
	}

	return fmt.Errorf("permission not found: %s", permissionCode)
}

// RemovePermission removes a single permission from a role
func (s *RoleService) RemovePermission(ctx context.Context, roleID uuid.UUID, permissionCode string) error {
	perms, err := s.db.ListPermissions(ctx)
	if err != nil {
		return err
	}

	for _, p := range perms {
		if p.Code == permissionCode {
			return s.db.RemoveRolePermission(ctx, database.RemoveRolePermissionParams{
				RoleID:       roleID,
				PermissionID: p.ID,
			})
		}
	}

	return fmt.Errorf("permission not found: %s", permissionCode)
}
