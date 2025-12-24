package tasks

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
)

// NotificationPayload contains data for sending notifications
type NotificationPayload struct {
	Type        string    `json:"type"` // "comment", "reply", "welcome", "mention", "new_post"
	TenantID    uuid.UUID `json:"tenant_id"`
	RecipientID uuid.UUID `json:"recipient_id"`

	// For comment/reply notifications
	PostID    *uuid.UUID `json:"post_id,omitempty"`
	PostTitle string     `json:"post_title,omitempty"`
	CommentID *uuid.UUID `json:"comment_id,omitempty"`
	AuthorID  *uuid.UUID `json:"author_id,omitempty"`
	AuthorName string    `json:"author_name,omitempty"`

	// For welcome notifications
	CommunityName string `json:"community_name,omitempty"`
}

// NewSendNotificationTask creates a new notification task
func NewSendNotificationTask(payload NotificationPayload) (*asynq.Task, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return asynq.NewTask(
		TypeSendNotification,
		data,
		asynq.Queue(QueueDefault),
		asynq.MaxRetry(3),
		asynq.Timeout(30*time.Second),
		asynq.Retention(24*time.Hour),
	), nil
}
