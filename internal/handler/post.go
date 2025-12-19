package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/nickkcj/orbit-backend/internal/service"
)

type CreatePostRequest struct {
	CategoryID    string `json:"category_id"`
	Title         string `json:"title" validate:"required"`
	Content       string `json:"content"`
	ContentFormat string `json:"content_format"`
	Excerpt       string `json:"excerpt"`
	CoverImageURL string `json:"cover_image_url"`
}

type UpdatePostRequest struct {
	Title         string `json:"title"`
	Content       string `json:"content"`
	Excerpt       string `json:"excerpt"`
	CoverImageURL string `json:"cover_image_url"`
	CategoryID    string `json:"category_id"`
}

func (h *Handler) CreatePost(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	var req CreatePostRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	// Get author from auth context
	user := GetUserFromContext(c)
	authorID := user.ID

	input := service.CreatePostInput{
		TenantID:      tenant.ID,
		AuthorID:      authorID,
		Title:         req.Title,
		Content:       req.Content,
		ContentFormat: req.ContentFormat,
		Excerpt:       req.Excerpt,
		CoverImageURL: req.CoverImageURL,
	}

	if req.CategoryID != "" {
		catID, err := uuid.Parse(req.CategoryID)
		if err == nil {
			input.CategoryID = &catID
		}
	}

	post, err := h.services.Post.Create(c.Request().Context(), input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, post)
}

func (h *Handler) GetPost(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid post id"})
	}

	post, err := h.services.Post.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "post not found"})
	}

	// Increment view count
	_ = h.services.Post.IncrementViews(c.Request().Context(), id)

	return c.JSON(http.StatusOK, post)
}

func (h *Handler) ListPosts(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	posts, err := h.services.Post.ListByTenant(c.Request().Context(), tenant.ID, int32(limit), int32(offset))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list posts"})
	}

	return c.JSON(http.StatusOK, posts)
}

func (h *Handler) UpdatePost(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid post id"})
	}

	var req UpdatePostRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	input := service.UpdatePostInput{
		Title:         req.Title,
		Content:       req.Content,
		Excerpt:       req.Excerpt,
		CoverImageURL: req.CoverImageURL,
	}

	if req.CategoryID != "" {
		catID, err := uuid.Parse(req.CategoryID)
		if err == nil {
			input.CategoryID = &catID
		}
	}

	post, err := h.services.Post.Update(c.Request().Context(), id, input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, post)
}

func (h *Handler) PublishPost(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid post id"})
	}

	post, err := h.services.Post.Publish(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, post)
}

func (h *Handler) DeletePost(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid post id"})
	}

	if err := h.services.Post.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}
