package handler

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/nickkcj/orbit-backend/internal/service"
	"github.com/nickkcj/orbit-backend/internal/worker/tasks"
)

type CreateCommentRequest struct {
	PostID   string `json:"post_id" validate:"required"`
	ParentID string `json:"parent_id"`
	Content  string `json:"content" validate:"required"`
}

type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required"`
}

func (h *Handler) CreateComment(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	var req CreateCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	postID, err := uuid.Parse(req.PostID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid post_id"})
	}

	input := service.CreateCommentInput{
		TenantID: tenant.ID,
		PostID:   postID,
		AuthorID: user.ID,
		Content:  req.Content,
	}

	var parentID *uuid.UUID
	if req.ParentID != "" {
		pid, err := uuid.Parse(req.ParentID)
		if err == nil {
			parentID = &pid
			input.ParentID = &pid
		}
	}

	comment, err := h.services.Comment.Create(c.Request().Context(), input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	// Enqueue notification task asynchronously
	go func() {
		ctx := context.Background() // Use fresh context, not request context

		// Get the post to notify the author
		post, err := h.services.Post.GetByID(ctx, postID)
		if err != nil {
			return
		}

		var payload tasks.NotificationPayload

		if parentID != nil {
			// This is a reply - notify parent comment author
			parentComment, err := h.services.Comment.GetByID(ctx, *parentID)
			if err == nil && parentComment.AuthorID != user.ID {
				payload = tasks.NotificationPayload{
					Type:        "reply",
					TenantID:    tenant.ID,
					RecipientID: parentComment.AuthorID,
					PostID:      &postID,
					PostTitle:   post.Title,
					CommentID:   &comment.ID,
					AuthorName:  user.Name,
				}
			}
		} else {
			// This is a top-level comment - notify post author
			if post.AuthorID != user.ID {
				payload = tasks.NotificationPayload{
					Type:        "comment",
					TenantID:    tenant.ID,
					RecipientID: post.AuthorID,
					PostID:      &postID,
					PostTitle:   post.Title,
					CommentID:   &comment.ID,
					AuthorName:  user.Name,
				}
			}
		}

		// Enqueue the notification task
		if payload.Type != "" && h.taskClient != nil {
			task, err := tasks.NewSendNotificationTask(payload)
			if err == nil {
				if _, err := h.taskClient.Enqueue(task); err != nil {
					log.Printf("Failed to enqueue notification task: %v", err)
				}
			}
		}
	}()

	return c.JSON(http.StatusCreated, comment)
}

func (h *Handler) GetComment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid comment id"})
	}

	comment, err := h.services.Comment.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "comment not found"})
	}

	return c.JSON(http.StatusOK, comment)
}

func (h *Handler) ListComments(c echo.Context) error {
	postID, err := uuid.Parse(c.Param("postId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid post_id"})
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	comments, err := h.services.Comment.ListByPost(c.Request().Context(), postID, int32(limit), int32(offset))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list comments"})
	}

	return c.JSON(http.StatusOK, comments)
}

func (h *Handler) ListReplies(c echo.Context) error {
	parentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid comment id"})
	}

	replies, err := h.services.Comment.ListReplies(c.Request().Context(), parentID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list replies"})
	}

	return c.JSON(http.StatusOK, replies)
}

func (h *Handler) UpdateComment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid comment id"})
	}

	var req UpdateCommentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	comment, err := h.services.Comment.Update(c.Request().Context(), id, req.Content)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, comment)
}

func (h *Handler) DeleteComment(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid comment id"})
	}

	if err := h.services.Comment.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}
