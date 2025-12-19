package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/nickkcj/orbit-backend/internal/database"
)

type MemberService struct {
	db *database.Queries
}

func NewMemberService(db *database.Queries) *MemberService {
	return &MemberService{db: db}
}

func (s *MemberService) Add(ctx context.Context, tenantID, userID, roleID uuid.UUID, displayName string) (database.TenantMember, error) {
	return s.db.AddMember(ctx, database.AddMemberParams{
		TenantID:    tenantID,
		UserID:      userID,
		RoleID:      roleID,
		DisplayName: sql.NullString{String: displayName, Valid: displayName != ""},
	})
}

func (s *MemberService) AddWithDefaultRole(ctx context.Context, tenantID, userID uuid.UUID, displayName string) (database.TenantMember, error) {
	// Get default role for this tenant
	role, err := s.db.GetDefaultRole(ctx, tenantID)
	if err != nil {
		return database.TenantMember{}, err
	}

	return s.Add(ctx, tenantID, userID, role.ID, displayName)
}

func (s *MemberService) Get(ctx context.Context, tenantID, userID uuid.UUID) (database.TenantMember, error) {
	return s.db.GetMember(ctx, database.GetMemberParams{
		TenantID: tenantID,
		UserID:   userID,
	})
}

func (s *MemberService) GetWithRole(ctx context.Context, tenantID, userID uuid.UUID) (database.GetMemberWithRoleRow, error) {
	return s.db.GetMemberWithRole(ctx, database.GetMemberWithRoleParams{
		TenantID: tenantID,
		UserID:   userID,
	})
}

func (s *MemberService) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]database.ListMembersByTenantRow, error) {
	return s.db.ListMembersByTenant(ctx, tenantID)
}

func (s *MemberService) ListTenantsByUser(ctx context.Context, userID uuid.UUID) ([]database.ListTenantsByUserRow, error) {
	return s.db.ListTenantsByUser(ctx, userID)
}

func (s *MemberService) UpdateRole(ctx context.Context, tenantID, userID, roleID uuid.UUID) (database.TenantMember, error) {
	return s.db.UpdateMemberRole(ctx, database.UpdateMemberRoleParams{
		TenantID: tenantID,
		UserID:   userID,
		RoleID:   roleID,
	})
}

func (s *MemberService) UpdateStatus(ctx context.Context, tenantID, userID uuid.UUID, status string) error {
	return s.db.UpdateMemberStatus(ctx, database.UpdateMemberStatusParams{
		TenantID: tenantID,
		UserID:   userID,
		Status:   status,
	})
}

func (s *MemberService) Remove(ctx context.Context, tenantID, userID uuid.UUID) error {
	return s.db.RemoveMember(ctx, database.RemoveMemberParams{
		TenantID: tenantID,
		UserID:   userID,
	})
}

func (s *MemberService) Count(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	return s.db.CountMembersByTenant(ctx, tenantID)
}
