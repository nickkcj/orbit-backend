package service

import (
	"github.com/nickkcj/orbit-backend/internal/database"
)

type Services struct {
	Auth     *AuthService
	Tenant   *TenantService
	User     *UserService
	Post     *PostService
	Comment  *CommentService
	Category *CategoryService
	Member   *MemberService
}

func New(db *database.Queries, jwtSecret string) *Services {
	return &Services{
		Auth:     NewAuthService(db, jwtSecret),
		Tenant:   NewTenantService(db),
		User:     NewUserService(db),
		Post:     NewPostService(db),
		Comment:  NewCommentService(db),
		Category: NewCategoryService(db),
		Member:   NewMemberService(db),
	}
}
