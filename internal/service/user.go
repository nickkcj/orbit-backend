package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/nickkcj/orbit-backend/internal/database"
)

type UserService struct {
	db *database.Queries
}

func NewUserService(db *database.Queries) *UserService {
	return &UserService{db: db}
}

func (s *UserService) Create(ctx context.Context, email, password, name string) (database.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return database.User{}, err
	}

	return s.db.CreateUser(ctx, database.CreateUserParams{
		Email:        email,
		PasswordHash: string(hashedPassword),
		Name:         name,
		AvatarUrl:    sql.NullString{},
	})
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (database.User, error) {
	return s.db.GetUserByEmail(ctx, email)
}

func (s *UserService) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func (s *UserService) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) (database.User, error) {
	user, err := s.db.GetUserByID(ctx, userID)
	if err != nil {
		return database.User{}, err
	}

	avatar := sql.NullString{String: avatarURL, Valid: avatarURL != ""}

	return s.db.UpdateUser(ctx, database.UpdateUserParams{
		ID:        user.ID,
		Name:      user.Name,
		AvatarUrl: avatar,
	})
}
