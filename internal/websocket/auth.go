package websocket

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nickkcj/orbit-backend/internal/service"
)

var (
	ErrMissingToken   = errors.New("missing authentication token")
	ErrInvalidToken   = errors.New("invalid or expired token")
	ErrMissingTenant  = errors.New("missing tenant identifier")
	ErrInvalidTenant  = errors.New("invalid tenant")
	ErrInactiveTenant = errors.New("tenant is not active")
	ErrInactiveUser   = errors.New("user account is not active")
)

// AuthResult contains the authenticated user and tenant info
type AuthResult struct {
	UserID   uuid.UUID
	TenantID uuid.UUID
}

// Authenticator handles WebSocket authentication
type Authenticator struct {
	authService   *service.AuthService
	tenantService *service.TenantService
}

// NewAuthenticator creates a new WebSocket authenticator
func NewAuthenticator(authService *service.AuthService, tenantService *service.TenantService) *Authenticator {
	return &Authenticator{
		authService:   authService,
		tenantService: tenantService,
	}
}

// Authenticate validates token and tenant for WebSocket connection
// Token can come from: query param (?token=xxx) or cookie
// Tenant can come from: query param (?tenant=slug) or X-Tenant-Slug header
func (a *Authenticator) Authenticate(ctx context.Context, token, tenantSlug string) (*AuthResult, error) {
	if token == "" {
		return nil, ErrMissingToken
	}
	if tenantSlug == "" {
		return nil, ErrMissingTenant
	}

	// Validate JWT token
	claims, err := a.authService.ValidateToken(token)
	if err != nil {
		return nil, ErrInvalidToken
	}

	// Verify user is active
	user, err := a.authService.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrInvalidToken
	}
	if user.Status != "active" {
		return nil, ErrInactiveUser
	}

	// Validate tenant
	tenant, err := a.tenantService.GetBySlug(ctx, tenantSlug)
	if err != nil {
		return nil, ErrInvalidTenant
	}
	if tenant.Status != "active" {
		return nil, ErrInactiveTenant
	}

	return &AuthResult{
		UserID:   user.ID,
		TenantID: tenant.ID,
	}, nil
}
