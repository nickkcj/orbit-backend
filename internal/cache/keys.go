package cache

import (
	"fmt"

	"github.com/google/uuid"
)

// Cache key prefixes
const (
	PrefixTenant     = "tenant"
	PrefixPermission = "perms"
	PrefixPosts      = "posts"
	PrefixMember     = "member"
)

// Cache TTLs
const (
	TTLTenant      = 5 * 60  // 5 minutes in seconds
	TTLPermissions = 5 * 60  // 5 minutes in seconds
	TTLPosts       = 1 * 60  // 1 minute in seconds
	TTLMember      = 5 * 60  // 5 minutes in seconds
)

// TenantBySlugKey returns the cache key for tenant by slug
func TenantBySlugKey(slug string) string {
	return fmt.Sprintf("%s:slug:%s", PrefixTenant, slug)
}

// TenantByIDKey returns the cache key for tenant by ID
func TenantByIDKey(id uuid.UUID) string {
	return fmt.Sprintf("%s:id:%s", PrefixTenant, id)
}

// UserPermissionsKey returns the cache key for user permissions
func UserPermissionsKey(tenantID, userID uuid.UUID) string {
	return fmt.Sprintf("%s:%s:%s", PrefixPermission, tenantID, userID)
}

// PostListKey returns the cache key for post list
func PostListKey(tenantID uuid.UUID, offset, limit int32) string {
	return fmt.Sprintf("%s:list:%s:%d:%d", PrefixPosts, tenantID, offset, limit)
}

// MemberKey returns the cache key for member data
func MemberKey(tenantID, userID uuid.UUID) string {
	return fmt.Sprintf("%s:%s:%s", PrefixMember, tenantID, userID)
}

// Pattern builders for bulk invalidation

// TenantPattern returns pattern to invalidate all tenant cache
func TenantPattern(tenantID uuid.UUID) string {
	return fmt.Sprintf("%s:*:%s", PrefixTenant, tenantID)
}

// PermissionsTenantPattern returns pattern to invalidate all permissions for a tenant
func PermissionsTenantPattern(tenantID uuid.UUID) string {
	return fmt.Sprintf("%s:%s:*", PrefixPermission, tenantID)
}

// PostsTenantPattern returns pattern to invalidate all posts for a tenant
func PostsTenantPattern(tenantID uuid.UUID) string {
	return fmt.Sprintf("%s:*:%s:*", PrefixPosts, tenantID)
}
