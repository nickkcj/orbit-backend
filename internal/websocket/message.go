package websocket

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MessageType defines the type of WebSocket message
type MessageType string

const (
	// Notifications
	MessageTypeNotificationNew MessageType = "notification:new"

	// Posts
	MessageTypePostCreated MessageType = "post:created"
	MessageTypePostUpdated MessageType = "post:updated"
	MessageTypePostDeleted MessageType = "post:deleted"

	// Comments
	MessageTypeCommentCreated MessageType = "comment:created"
	MessageTypeCommentDeleted MessageType = "comment:deleted"

	// Likes
	MessageTypeLikeUpdated MessageType = "like:updated"

	// Connection
	MessageTypeConnected MessageType = "connected"
	MessageTypePing      MessageType = "ping"
	MessageTypePong      MessageType = "pong"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType     `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

// NotificationPayload for notification messages
type NotificationPayload struct {
	ID        uuid.UUID `json:"id"`
	Type      string    `json:"notification_type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Data      any       `json:"data,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// PostCreatedPayload for new post messages
type PostCreatedPayload struct {
	ID         uuid.UUID  `json:"id"`
	Title      string     `json:"title"`
	AuthorID   uuid.UUID  `json:"author_id"`
	AuthorName string     `json:"author_name"`
	CategoryID *uuid.UUID `json:"category_id,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// CommentCreatedPayload for new comment messages
type CommentCreatedPayload struct {
	ID         uuid.UUID  `json:"id"`
	PostID     uuid.UUID  `json:"post_id"`
	AuthorID   uuid.UUID  `json:"author_id"`
	AuthorName string     `json:"author_name"`
	ParentID   *uuid.UUID `json:"parent_id,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// LikeUpdatedPayload for like count changes
type LikeUpdatedPayload struct {
	TargetType string    `json:"target_type"` // "post" or "comment"
	TargetID   uuid.UUID `json:"target_id"`
	LikeCount  int       `json:"like_count"`
}

// ConnectedPayload for connection confirmation
type ConnectedPayload struct {
	Status   string `json:"status"`
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
}

// NewMessage creates a new message with the given type and payload
func NewMessage(msgType MessageType, payload any) (*Message, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return &Message{
		Type:      msgType,
		Payload:   data,
		Timestamp: time.Now(),
	}, nil
}
