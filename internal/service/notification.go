package service

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/nickkcj/orbit-backend/internal/database"
	"github.com/sqlc-dev/pqtype"
)

type NotificationService struct {
	db *database.Queries
}

func NewNotificationService(db *database.Queries) *NotificationService {
	return &NotificationService{db: db}
}

type NotificationType string

const (
	NotificationTypeComment  NotificationType = "comment"
	NotificationTypeReply    NotificationType = "reply"
	NotificationTypeMention  NotificationType = "mention"
	NotificationTypeNewPost  NotificationType = "new_post"
	NotificationTypeWelcome  NotificationType = "welcome"
)

type NotificationData struct {
	PostID      *uuid.UUID `json:"post_id,omitempty"`
	PostTitle   string     `json:"post_title,omitempty"`
	CommentID   *uuid.UUID `json:"comment_id,omitempty"`
	AuthorID    *uuid.UUID `json:"author_id,omitempty"`
	AuthorName  string     `json:"author_name,omitempty"`
}

type CreateNotificationInput struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
	Type     NotificationType
	Title    string
	Message  string
	Data     NotificationData
}

func (s *NotificationService) Create(ctx context.Context, input CreateNotificationInput) (database.Notification, error) {
	dataJSON, err := json.Marshal(input.Data)
	if err != nil {
		dataJSON = []byte("{}")
	}

	return s.db.CreateNotification(ctx, database.CreateNotificationParams{
		TenantID: input.TenantID,
		UserID:   input.UserID,
		Type:     string(input.Type),
		Title:    input.Title,
		Message:  sql.NullString{String: input.Message, Valid: input.Message != ""},
		Data:     pqtype.NullRawMessage{RawMessage: dataJSON, Valid: true},
	})
}

func (s *NotificationService) List(ctx context.Context, tenantID, userID uuid.UUID, limit, offset int32) ([]database.Notification, error) {
	return s.db.ListNotificationsByUser(ctx, database.ListNotificationsByUserParams{
		TenantID: tenantID,
		UserID:   userID,
		Limit:    limit,
		Offset:   offset,
	})
}

func (s *NotificationService) ListUnread(ctx context.Context, tenantID, userID uuid.UUID) ([]database.Notification, error) {
	return s.db.ListUnreadNotifications(ctx, database.ListUnreadNotificationsParams{
		TenantID: tenantID,
		UserID:   userID,
	})
}

func (s *NotificationService) CountUnread(ctx context.Context, tenantID, userID uuid.UUID) (int64, error) {
	return s.db.CountUnreadNotifications(ctx, database.CountUnreadNotificationsParams{
		TenantID: tenantID,
		UserID:   userID,
	})
}

func (s *NotificationService) MarkRead(ctx context.Context, notificationID, userID uuid.UUID) error {
	return s.db.MarkNotificationRead(ctx, database.MarkNotificationReadParams{
		ID:     notificationID,
		UserID: userID,
	})
}

func (s *NotificationService) MarkAllRead(ctx context.Context, tenantID, userID uuid.UUID) error {
	return s.db.MarkAllNotificationsRead(ctx, database.MarkAllNotificationsReadParams{
		TenantID: tenantID,
		UserID:   userID,
	})
}

func (s *NotificationService) Delete(ctx context.Context, notificationID, userID uuid.UUID) error {
	return s.db.DeleteNotification(ctx, database.DeleteNotificationParams{
		ID:     notificationID,
		UserID: userID,
	})
}

// Helper methods for creating specific notification types

func (s *NotificationService) NotifyComment(ctx context.Context, tenantID, postAuthorID uuid.UUID, commenterName, postTitle string, postID, commentID uuid.UUID) error {
	_, err := s.Create(ctx, CreateNotificationInput{
		TenantID: tenantID,
		UserID:   postAuthorID,
		Type:     NotificationTypeComment,
		Title:    "Novo comentário",
		Message:  commenterName + " comentou no seu post \"" + postTitle + "\"",
		Data: NotificationData{
			PostID:     &postID,
			PostTitle:  postTitle,
			CommentID:  &commentID,
			AuthorName: commenterName,
		},
	})
	return err
}

func (s *NotificationService) NotifyReply(ctx context.Context, tenantID, commentAuthorID uuid.UUID, replierName, postTitle string, postID, commentID uuid.UUID) error {
	_, err := s.Create(ctx, CreateNotificationInput{
		TenantID: tenantID,
		UserID:   commentAuthorID,
		Type:     NotificationTypeReply,
		Title:    "Nova resposta",
		Message:  replierName + " respondeu seu comentário em \"" + postTitle + "\"",
		Data: NotificationData{
			PostID:     &postID,
			PostTitle:  postTitle,
			CommentID:  &commentID,
			AuthorName: replierName,
		},
	})
	return err
}

func (s *NotificationService) NotifyWelcome(ctx context.Context, tenantID, userID uuid.UUID, communityName string) error {
	_, err := s.Create(ctx, CreateNotificationInput{
		TenantID: tenantID,
		UserID:   userID,
		Type:     NotificationTypeWelcome,
		Title:    "Bem-vindo!",
		Message:  "Você agora faz parte da comunidade " + communityName + "!",
		Data:     NotificationData{},
	})
	return err
}
