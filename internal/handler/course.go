package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/nickkcj/orbit-backend/internal/service"
)

// ============================================================================
// REQUEST STRUCTS
// ============================================================================

type CreateCourseRequest struct {
	Title        string `json:"title" validate:"required"`
	Description  string `json:"description"`
	ThumbnailURL string `json:"thumbnail_url"`
}

type UpdateCourseRequest struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	ThumbnailURL string `json:"thumbnail_url"`
}

type CreateModuleRequest struct {
	Title       string `json:"title" validate:"required"`
	Description string `json:"description"`
}

type UpdateModuleRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type ReorderRequest struct {
	Position int32 `json:"position" validate:"required"`
}

type CreateLessonRequest struct {
	Title           string `json:"title" validate:"required"`
	Description     string `json:"description"`
	Content         string `json:"content"`
	ContentFormat   string `json:"content_format"`
	VideoID         string `json:"video_id"`
	DurationMinutes int32  `json:"duration_minutes"`
	IsFreePreview   bool   `json:"is_free_preview"`
}

type UpdateLessonRequest struct {
	Title           string `json:"title"`
	Description     string `json:"description"`
	Content         string `json:"content"`
	ContentFormat   string `json:"content_format"`
	VideoID         string `json:"video_id"`
	DurationMinutes int32  `json:"duration_minutes"`
	IsFreePreview   bool   `json:"is_free_preview"`
}

// ============================================================================
// COURSE HANDLERS
// ============================================================================

func (h *Handler) CreateCourse(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	user := GetUserFromContext(c)
	if user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	var req CreateCourseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "title is required"})
	}

	course, err := h.services.Course.Create(c.Request().Context(), service.CreateCourseInput{
		TenantID:     tenant.ID,
		AuthorID:     user.ID,
		Title:        req.Title,
		Description:  req.Description,
		ThumbnailURL: req.ThumbnailURL,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, course)
}

func (h *Handler) GetCourse(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid course id"})
	}

	course, err := h.services.Course.GetWithDetails(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "course not found"})
	}

	// Verify tenant ownership
	if course.TenantID != tenant.ID {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "course not found"})
	}

	return c.JSON(http.StatusOK, course)
}

func (h *Handler) GetCourseStructure(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid course id"})
	}

	structure, err := h.services.Course.GetCourseStructure(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "course not found"})
	}

	// Verify tenant ownership
	if structure.Course.TenantID != tenant.ID {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "course not found"})
	}

	return c.JSON(http.StatusOK, structure)
}

func (h *Handler) ListCourses(c echo.Context) error {
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

	// Check if requesting all courses (admin view)
	status := c.QueryParam("status")
	if status == "all" {
		courses, err := h.services.Course.ListAllByTenant(c.Request().Context(), tenant.ID, int32(limit), int32(offset))
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list courses"})
		}
		return c.JSON(http.StatusOK, courses)
	}

	// Default: published courses only
	courses, err := h.services.Course.ListByTenant(c.Request().Context(), tenant.ID, int32(limit), int32(offset))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list courses"})
	}

	return c.JSON(http.StatusOK, courses)
}

func (h *Handler) UpdateCourse(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	user := GetUserFromContext(c)

	if tenant == nil || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid course id"})
	}

	// Verify course exists and belongs to tenant
	course, err := h.services.Course.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "course not found"})
	}

	if course.TenantID != tenant.ID {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "course not found"})
	}

	// Check ownership for courses.edit_own permission
	if course.AuthorID != user.ID {
		hasFullEdit, err := h.services.Permission.HasPermission(c.Request().Context(), tenant.ID, user.ID, "courses.edit")
		if err != nil || !hasFullEdit {
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "permission denied"})
		}
	}

	var req UpdateCourseRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	updated, err := h.services.Course.Update(c.Request().Context(), id, service.UpdateCourseInput{
		Title:        req.Title,
		Description:  req.Description,
		ThumbnailURL: req.ThumbnailURL,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, updated)
}

func (h *Handler) PublishCourse(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid course id"})
	}

	course, err := h.services.Course.Publish(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, course)
}

func (h *Handler) UnpublishCourse(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid course id"})
	}

	course, err := h.services.Course.Unpublish(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, course)
}

func (h *Handler) DeleteCourse(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	user := GetUserFromContext(c)

	if tenant == nil || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid course id"})
	}

	course, err := h.services.Course.GetByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "course not found"})
	}

	if course.TenantID != tenant.ID {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "course not found"})
	}

	// Check ownership for courses.delete_own permission
	if course.AuthorID != user.ID {
		hasFullDelete, err := h.services.Permission.HasPermission(c.Request().Context(), tenant.ID, user.ID, "courses.delete")
		if err != nil || !hasFullDelete {
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "permission denied"})
		}
	}

	if err := h.services.Course.Delete(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

// ============================================================================
// MODULE HANDLERS
// ============================================================================

func (h *Handler) CreateModule(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid course id"})
	}

	var req CreateModuleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "title is required"})
	}

	module, err := h.services.Course.CreateModule(c.Request().Context(), service.CreateModuleInput{
		TenantID:    tenant.ID,
		CourseID:    courseID,
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, module)
}

func (h *Handler) GetModule(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid module id"})
	}

	module, err := h.services.Course.GetModuleByID(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "module not found"})
	}

	return c.JSON(http.StatusOK, module)
}

func (h *Handler) ListModules(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid course id"})
	}

	modules, err := h.services.Course.ListModulesByCourse(c.Request().Context(), courseID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list modules"})
	}

	return c.JSON(http.StatusOK, modules)
}

func (h *Handler) UpdateModule(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid module id"})
	}

	var req UpdateModuleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	module, err := h.services.Course.UpdateModule(c.Request().Context(), id, service.UpdateModuleInput{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, module)
}

func (h *Handler) ReorderModule(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid module id"})
	}

	var req ReorderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if err := h.services.Course.ReorderModule(c.Request().Context(), id, req.Position); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) DeleteModule(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid module id"})
	}

	if err := h.services.Course.DeleteModule(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

// ============================================================================
// LESSON HANDLERS
// ============================================================================

func (h *Handler) CreateLesson(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	moduleID, err := uuid.Parse(c.Param("moduleId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid module id"})
	}

	var req CreateLessonRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if req.Title == "" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "title is required"})
	}

	input := service.CreateLessonInput{
		TenantID:        tenant.ID,
		ModuleID:        moduleID,
		Title:           req.Title,
		Description:     req.Description,
		Content:         req.Content,
		ContentFormat:   req.ContentFormat,
		DurationMinutes: req.DurationMinutes,
		IsFreePreview:   req.IsFreePreview,
	}

	if req.VideoID != "" {
		videoID, err := uuid.Parse(req.VideoID)
		if err == nil {
			input.VideoID = &videoID
		}
	}

	lesson, err := h.services.Course.CreateLesson(c.Request().Context(), input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, lesson)
}

func (h *Handler) GetLesson(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid lesson id"})
	}

	lesson, err := h.services.Course.GetLessonWithVideo(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "lesson not found"})
	}

	return c.JSON(http.StatusOK, lesson)
}

func (h *Handler) ListLessons(c echo.Context) error {
	moduleID, err := uuid.Parse(c.Param("moduleId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid module id"})
	}

	lessons, err := h.services.Course.ListLessonsByModule(c.Request().Context(), moduleID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list lessons"})
	}

	return c.JSON(http.StatusOK, lessons)
}

func (h *Handler) UpdateLesson(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid lesson id"})
	}

	var req UpdateLessonRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	input := service.UpdateLessonInput{
		Title:           req.Title,
		Description:     req.Description,
		Content:         req.Content,
		ContentFormat:   req.ContentFormat,
		DurationMinutes: req.DurationMinutes,
		IsFreePreview:   req.IsFreePreview,
	}

	if req.VideoID != "" {
		videoID, err := uuid.Parse(req.VideoID)
		if err == nil {
			input.VideoID = &videoID
		}
	}

	lesson, err := h.services.Course.UpdateLesson(c.Request().Context(), id, input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusOK, lesson)
}

func (h *Handler) ReorderLesson(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid lesson id"})
	}

	var req ReorderRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	if err := h.services.Course.ReorderLesson(c.Request().Context(), id, req.Position); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.NoContent(http.StatusOK)
}

func (h *Handler) DeleteLesson(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid lesson id"})
	}

	if err := h.services.Course.DeleteLesson(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}
