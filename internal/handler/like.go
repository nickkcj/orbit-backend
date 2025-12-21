package handler

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/nickkcj/orbit-backend/internal/service"
)

type LikeResponse struct {
	Liked     bool `json:"liked"`
	LikeCount int  `json:"like_count"`
}

// LikePost handles POST /posts/:id/like
func (h *Handler) LikePost(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid post id"})
	}

	err = h.services.Like.LikePost(c.Request().Context(), tenant.ID, user.ID, postID)
	if errors.Is(err, service.ErrAlreadyLiked) {
		return c.JSON(http.StatusConflict, ErrorResponse{Error: "already liked"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to like post"})
	}

	// Get updated post to return like count
	post, err := h.services.Post.GetByID(c.Request().Context(), tenant.ID, postID)
	if err != nil {
		return c.JSON(http.StatusOK, LikeResponse{Liked: true, LikeCount: 0})
	}

	return c.JSON(http.StatusOK, LikeResponse{
		Liked:     true,
		LikeCount: int(post.LikeCount),
	})
}

// UnlikePost handles DELETE /posts/:id/like
func (h *Handler) UnlikePost(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid post id"})
	}

	err = h.services.Like.UnlikePost(c.Request().Context(), tenant.ID, user.ID, postID)
	if errors.Is(err, service.ErrNotLiked) {
		return c.JSON(http.StatusConflict, ErrorResponse{Error: "not liked"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to unlike post"})
	}

	// Get updated post to return like count
	post, err := h.services.Post.GetByID(c.Request().Context(), tenant.ID, postID)
	if err != nil {
		return c.JSON(http.StatusOK, LikeResponse{Liked: false, LikeCount: 0})
	}

	return c.JSON(http.StatusOK, LikeResponse{
		Liked:     false,
		LikeCount: int(post.LikeCount),
	})
}

// LikeComment handles POST /comments/:id/like
func (h *Handler) LikeComment(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	commentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid comment id"})
	}

	err = h.services.Like.LikeComment(c.Request().Context(), tenant.ID, user.ID, commentID)
	if errors.Is(err, service.ErrAlreadyLiked) {
		return c.JSON(http.StatusConflict, ErrorResponse{Error: "already liked"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to like comment"})
	}

	// Get updated comment to return like count
	comment, err := h.services.Comment.GetByID(c.Request().Context(), tenant.ID, commentID)
	if err != nil {
		return c.JSON(http.StatusOK, LikeResponse{Liked: true, LikeCount: 0})
	}

	return c.JSON(http.StatusOK, LikeResponse{
		Liked:     true,
		LikeCount: int(comment.LikeCount),
	})
}

// UnlikeComment handles DELETE /comments/:id/like
func (h *Handler) UnlikeComment(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	commentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid comment id"})
	}

	err = h.services.Like.UnlikeComment(c.Request().Context(), tenant.ID, user.ID, commentID)
	if errors.Is(err, service.ErrNotLiked) {
		return c.JSON(http.StatusConflict, ErrorResponse{Error: "not liked"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to unlike comment"})
	}

	// Get updated comment to return like count
	comment, err := h.services.Comment.GetByID(c.Request().Context(), tenant.ID, commentID)
	if err != nil {
		return c.JSON(http.StatusOK, LikeResponse{Liked: false, LikeCount: 0})
	}

	return c.JSON(http.StatusOK, LikeResponse{
		Liked:     false,
		LikeCount: int(comment.LikeCount),
	})
}

// GetPostLikeStatus handles GET /posts/:id/like
func (h *Handler) GetPostLikeStatus(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "authentication required"})
	}

	postID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid post id"})
	}

	liked, err := h.services.Like.HasUserLikedPost(c.Request().Context(), user.ID, postID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to check like status"})
	}

	// Get post to return like count
	post, err := h.services.Post.GetByID(c.Request().Context(), tenant.ID, postID)
	if err != nil {
		return c.JSON(http.StatusOK, LikeResponse{Liked: liked, LikeCount: 0})
	}

	return c.JSON(http.StatusOK, LikeResponse{
		Liked:     liked,
		LikeCount: int(post.LikeCount),
	})
}
