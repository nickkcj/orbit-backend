package service

import (
	"github.com/nickkcj/orbit-backend/internal/database"
)

type Services struct {
	Tenant *TenantService
	User   *UserService
}

func New(db *database.Queries) *Services {
	return &Services{
		Tenant: NewTenantService(db),
		User:   NewUserService(db),
	}
}
