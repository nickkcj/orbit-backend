package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nickkcj/orbit-backend/internal/cache"
	"github.com/nickkcj/orbit-backend/internal/database"
)

// CachedPermissions holds cached permission data for a user
type CachedPermissions struct {
	Permissions  map[string]bool `json:"permissions"`
	RoleSlug     string          `json:"role_slug"`
	RolePriority int32           `json:"role_priority"`
	CachedAt     time.Time       `json:"cached_at"`
}

// PermissionService handles permission checking with caching
type PermissionService struct {
	db    *database.Queries
	cache cache.Cache
}

// NewPermissionService creates a new permission service
func NewPermissionService(db *database.Queries, c cache.Cache) *PermissionService {
	return &PermissionService{
		db:    db,
		cache: c,
	}
}

// GetUserPermissions returns cached permissions or fetches from DB
func (s *PermissionService) GetUserPermissions(ctx context.Context, tenantID, userID uuid.UUID) (*CachedPermissions, error) {
	cacheKey := cache.UserPermissionsKey(tenantID, userID)

	// Try cache first
	var cached CachedPermissions
	if s.cache != nil {
		err := s.cache.Get(ctx, cacheKey, &cached)
		if err == nil {
			return &cached, nil
		}
	}

	// Fetch from database
	member, err := s.db.GetMemberWithRole(ctx, database.GetMemberWithRoleParams{
		TenantID: tenantID,
		UserID:   userID,
	})
	if err != nil {
		return nil, err
	}

	permissions, err := s.db.GetRolePermissions(ctx, member.RoleID)
	if err != nil {
		return nil, err
	}

	// Build permission map
	permMap := make(map[string]bool)
	for _, p := range permissions {
		permMap[p.Code] = true
	}

	result := &CachedPermissions{
		Permissions:  permMap,
		RoleSlug:     member.RoleSlug,
		RolePriority: member.RolePriority,
		CachedAt:     time.Now(),
	}

	// Cache with 5-minute TTL
	if s.cache != nil {
		s.cache.Set(ctx, cacheKey, result, time.Duration(cache.TTLPermissions)*time.Second)
	}

	return result, nil
}

// HasPermission checks if a user has a specific permission
func (s *PermissionService) HasPermission(ctx context.Context, tenantID, userID uuid.UUID, code string) (bool, error) {
	perms, err := s.GetUserPermissions(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}

	return perms.Permissions[code], nil
}

// HasAnyPermission checks if a user has any of the specified permissions
func (s *PermissionService) HasAnyPermission(ctx context.Context, tenantID, userID uuid.UUID, codes ...string) (bool, error) {
	perms, err := s.GetUserPermissions(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}

	for _, code := range codes {
		if perms.Permissions[code] {
			return true, nil
		}
	}

	return false, nil
}

// HasAllPermissions checks if a user has all of the specified permissions
func (s *PermissionService) HasAllPermissions(ctx context.Context, tenantID, userID uuid.UUID, codes ...string) (bool, error) {
	perms, err := s.GetUserPermissions(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}

	for _, code := range codes {
		if !perms.Permissions[code] {
			return false, nil
		}
	}

	return true, nil
}

// IsOwnerOrAdmin checks if the user has owner or admin role
func (s *PermissionService) IsOwnerOrAdmin(ctx context.Context, tenantID, userID uuid.UUID) (bool, error) {
	perms, err := s.GetUserPermissions(ctx, tenantID, userID)
	if err != nil {
		return false, err
	}

	return perms.RoleSlug == "owner" || perms.RoleSlug == "admin", nil
}

// InvalidateUserPermissions clears cached permissions for a user
func (s *PermissionService) InvalidateUserPermissions(ctx context.Context, tenantID, userID uuid.UUID) error {
	if s.cache == nil {
		return nil
	}
	return s.cache.Delete(ctx, cache.UserPermissionsKey(tenantID, userID))
}

// InvalidateTenantPermissions clears all cached permissions for a tenant
func (s *PermissionService) InvalidateTenantPermissions(ctx context.Context, tenantID uuid.UUID) error {
	if s.cache == nil {
		return nil
	}
	return s.cache.DeletePattern(ctx, cache.PermissionsTenantPattern(tenantID))
}
