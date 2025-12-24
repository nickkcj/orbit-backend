package service

import (
	"log"

	"github.com/nickkcj/orbit-backend/internal/cache"
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
	Role         *RoleService
	Storage      *StorageService
	Webhook      *WebhookService
	Notification *NotificationService
	Analytics    *AnalyticsService
	Like         *LikeService
	Permission   *PermissionService
	Stream       *StreamService
	Video        *VideoService
	Course       *CourseService
	Enrollment   *EnrollmentService
}

type StorageConfig struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
}

func New(db *database.Queries, jwtSecret string, storageConfig *StorageConfig, streamConfig *StreamConfig, googleConfig *GoogleOAuthConfig, c cache.Cache) *Services {
	services := &Services{
		Auth:         NewAuthService(db, jwtSecret, googleConfig),
		Tenant:       NewTenantService(db),
		User:         NewUserService(db),
		Post:         NewPostService(db),
		Comment:      NewCommentService(db),
		Category:     NewCategoryService(db),
		Member:       NewMemberService(db),
		Role:         NewRoleService(db),
		Webhook:      NewWebhookService(db),
		Notification: NewNotificationService(db),
		Analytics:    NewAnalyticsService(db),
		Like:         NewLikeService(db),
		Permission:   NewPermissionService(db, c),
		Course:       NewCourseService(db),
		Enrollment:   NewEnrollmentService(db),
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

	// Initialize stream and video services if config provided
	if streamConfig != nil && streamConfig.AccountID != "" && streamConfig.APIToken != "" {
		stream, err := NewStreamService(streamConfig)
		if err != nil {
			log.Printf("Warning: Failed to initialize stream service: %v", err)
		} else {
			services.Stream = stream
			services.Video = NewVideoService(db, stream)
			log.Println("Stream service initialized (Cloudflare)")
		}
	} else {
		// Create video service without stream (limited functionality)
		services.Video = NewVideoService(db, nil)
	}

	return services
}
