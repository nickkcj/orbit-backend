package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
	"github.com/nickkcj/orbit-backend/internal/service"
	"github.com/nickkcj/orbit-backend/internal/worker/tasks"
)

// NotificationHandler processes notification tasks
type NotificationHandler struct {
	notificationSvc *service.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(svc *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notificationSvc: svc}
}

// Handle processes a notification task
func (h *NotificationHandler) Handle(ctx context.Context, task *asynq.Task) error {
	var payload tasks.NotificationPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal notification payload: %w", err)
	}

	switch payload.Type {
	case "comment":
		if payload.PostID == nil || payload.CommentID == nil {
			return fmt.Errorf("missing post_id or comment_id for comment notification")
		}
		return h.notificationSvc.NotifyComment(
			ctx,
			payload.TenantID,
			payload.RecipientID,
			payload.AuthorName,
			payload.PostTitle,
			*payload.PostID,
			*payload.CommentID,
		)

	case "reply":
		if payload.PostID == nil || payload.CommentID == nil {
			return fmt.Errorf("missing post_id or comment_id for reply notification")
		}
		return h.notificationSvc.NotifyReply(
			ctx,
			payload.TenantID,
			payload.RecipientID,
			payload.AuthorName,
			payload.PostTitle,
			*payload.PostID,
			*payload.CommentID,
		)

	case "welcome":
		return h.notificationSvc.NotifyWelcome(
			ctx,
			payload.TenantID,
			payload.RecipientID,
			payload.CommunityName,
		)

	default:
		return fmt.Errorf("unknown notification type: %s", payload.Type)
	}
}
