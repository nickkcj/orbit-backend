package service

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nickkcj/orbit-backend/internal/database"
	"github.com/sqlc-dev/pqtype"
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

func (s *TenantService) GetByID(ctx context.Context, id uuid.UUID) (database.Tenant, error) {
	return s.db.GetTenantByID(ctx, id)
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

// TenantSettings represents the settings structure stored in JSONB
type TenantSettings struct {
	Theme *ThemeSettings `json:"theme,omitempty"`
}

type ThemeSettings struct {
	PrimaryColor string `json:"primaryColor,omitempty"`
	AccentColor  string `json:"accentColor,omitempty"`
	BannerURL    string `json:"bannerUrl,omitempty"`
}

func (s *TenantService) UpdateSettings(ctx context.Context, tenantID uuid.UUID, settings TenantSettings) (database.Tenant, error) {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return database.Tenant{}, err
	}

	return s.db.UpdateTenantSettings(ctx, database.UpdateTenantSettingsParams{
		ID:       tenantID,
		Settings: pqtype.NullRawMessage{RawMessage: settingsJSON, Valid: true},
	})
}

func (s *TenantService) UpdateLogo(ctx context.Context, tenantID uuid.UUID, logoURL string) (database.Tenant, error) {
	return s.db.UpdateTenantLogo(ctx, database.UpdateTenantLogoParams{
		ID:     tenantID,
		LogoUrl: sql.NullString{String: logoURL, Valid: logoURL != ""},
	})
}
