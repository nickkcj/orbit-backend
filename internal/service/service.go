package service

import (
	"log"

	"github.com/nickkcj/orbit-backend/internal/database"
)

type Services struct {
	Auth         *AuthService
	Tenant       *TenantService
	User         *UserService
	Post         *PostService
	Comment      *CommentService
	Category     *CategoryService
	Member       *MemberService
	Storage      *StorageService
	Webhook      *WebhookService
	Notification *NotificationService
}

type StorageConfig struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
}

func New(db *database.Queries, jwtSecret string, storageConfig *StorageConfig) *Services {
	services := &Services{
		Auth:         NewAuthService(db, jwtSecret),
		Tenant:       NewTenantService(db),
		User:         NewUserService(db),
		Post:         NewPostService(db),
		Comment:      NewCommentService(db),
		Category:     NewCategoryService(db),
		Member:       NewMemberService(db),
		Webhook:      NewWebhookService(db),
		Notification: NewNotificationService(db),
	}

	// Initialize storage service if config provided
	if storageConfig != nil && storageConfig.AccountID != "" {
		storage, err := NewStorageService(
			storageConfig.AccountID,
			storageConfig.AccessKeyID,
			storageConfig.SecretAccessKey,
			storageConfig.BucketName,
		)
		if err != nil {
			log.Printf("Warning: Failed to initialize storage service: %v", err)
		} else {
			services.Storage = storage
			log.Println("Storage service initialized (R2)")
		}
	}

	return services
}
