package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/nickkcj/orbit-backend/internal/service"
)

// ============================================================================
// REQUEST STRUCTS
// ============================================================================

type EnrollRequest struct {
	CourseID string `json:"course_id" validate:"required"`
}

type UpdateVideoProgressRequest struct {
	WatchDurationSeconds int32  `json:"watch_duration_seconds" validate:"required"`
	VideoTotalSeconds    *int32 `json:"video_total_seconds,omitempty"`
}

// ============================================================================
// ENROLLMENT HANDLERS
// ============================================================================

// EnrollInCourse enrolls the current user in a course
func (h *Handler) EnrollInCourse(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	user := GetUserFromContext(c)

	if tenant == nil || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	var req EnrollRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	courseID, err := uuid.Parse(req.CourseID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid course id"})
	}

	// Verify course exists and belongs to tenant
	course, err := h.services.Course.GetByID(c.Request().Context(), courseID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "course not found"})
	}

	if course.TenantID != tenant.ID {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "course not found"})
	}

	// Verify course is published
	if course.Status != "published" {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "cannot enroll in unpublished course"})
	}

	enrollment, err := h.services.Enrollment.Enroll(c.Request().Context(), service.EnrollInput{
		TenantID: tenant.ID,
		UserID:   user.ID,
		CourseID: courseID,
	})
	if err != nil {
		if errors.Is(err, service.ErrAlreadyEnrolled) {
			return c.JSON(http.StatusConflict, ErrorResponse{Error: "already enrolled in this course"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}

	return c.JSON(http.StatusCreated, enrollment)
}

// GetMyEnrollments lists all enrollments for the current user
func (h *Handler) GetMyEnrollments(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	user := GetUserFromContext(c)

	if tenant == nil || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	enrollments, err := h.services.Enrollment.ListEnrollmentsByUser(c.Request().Context(), user.ID, tenant.ID, int32(limit), int32(offset))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list enrollments"})
	}

	count, err := h.services.Enrollment.CountEnrollmentsByUser(c.Request().Context(), user.ID, tenant.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to count enrollments"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"enrollments": enrollments,
		"total":       count,
		"limit":       limit,
		"offset":      offset,
	})
}

// GetContinueLearning returns courses the user has started but not completed
func (h *Handler) GetContinueLearning(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	user := GetUserFromContext(c)

	if tenant == nil || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 10 {
		limit = 5
	}

	courses, err := h.services.Enrollment.GetContinueLearning(c.Request().Context(), user.ID, tenant.ID, int32(limit))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get continue learning courses"})
	}

	return c.JSON(http.StatusOK, courses)
}

// GetCourseEnrollmentStatus gets the enrollment status for a specific course
func (h *Handler) GetCourseEnrollmentStatus(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	user := GetUserFromContext(c)

	if tenant == nil || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid course id"})
	}

	// Check if enrolled
	enrolled, err := h.services.Enrollment.IsEnrolled(c.Request().Context(), user.ID, courseID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to check enrollment"})
	}

	if !enrolled {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"enrolled": false,
		})
	}

	// Get enrollment details
	enrollment, err := h.services.Enrollment.GetEnrollmentWithProgressByUserAndCourse(c.Request().Context(), user.ID, courseID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get enrollment"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"enrolled":   true,
		"enrollment": enrollment,
	})
}

// DropCourseEnrollment drops the user's enrollment in a course
func (h *Handler) DropCourseEnrollment(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	user := GetUserFromContext(c)

	if tenant == nil || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid course id"})
	}

	enrollment, err := h.services.Enrollment.GetEnrollment(c.Request().Context(), user.ID, courseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "not enrolled in this course"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get enrollment"})
	}

	if err := h.services.Enrollment.DropEnrollment(c.Request().Context(), enrollment.ID); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to drop enrollment"})
	}

	return c.NoContent(http.StatusNoContent)
}

// GetCourseProgress returns detailed progress for a course
func (h *Handler) GetCourseProgress(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	user := GetUserFromContext(c)

	if tenant == nil || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid course id"})
	}

	enrollment, err := h.services.Enrollment.GetEnrollment(c.Request().Context(), user.ID, courseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "not enrolled in this course"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get enrollment"})
	}

	playerData, err := h.services.Enrollment.GetCoursePlayerData(c.Request().Context(), enrollment.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get course progress"})
	}

	return c.JSON(http.StatusOK, playerData)
}

// ============================================================================
// LESSON PLAYER HANDLERS
// ============================================================================

// GetLessonForPlayer gets a lesson with progress info for the player
func (h *Handler) GetLessonForPlayer(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	user := GetUserFromContext(c)

	if tenant == nil || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	lessonID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid lesson id"})
	}

	// Get lesson to find course
	lesson, err := h.services.Course.GetLessonByID(c.Request().Context(), lessonID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "lesson not found"})
	}

	// Get module to find course
	module, err := h.services.Course.GetModuleByID(c.Request().Context(), lesson.ModuleID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "module not found"})
	}

	// Check access
	canAccess, err := h.services.Enrollment.CanAccessLesson(c.Request().Context(), user.ID, module.CourseID, lessonID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to check access"})
	}

	if !canAccess {
		return c.JSON(http.StatusForbidden, ErrorResponse{Error: "access denied - enrollment required"})
	}

	// Get enrollment for progress data
	enrollment, err := h.services.Enrollment.GetEnrollment(c.Request().Context(), user.ID, module.CourseID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get enrollment"})
	}

	// If enrolled, get lesson with progress
	if err == nil {
		lessonWithProgress, err := h.services.Enrollment.GetLessonWithProgressAndCourse(c.Request().Context(), lessonID, enrollment.ID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get lesson"})
		}

		// Update last accessed
		_ = h.services.Enrollment.UpdateLastAccessed(c.Request().Context(), enrollment.ID, lessonID)

		// Get previous/next lessons
		prevLesson, _ := h.services.Enrollment.GetPreviousLesson(c.Request().Context(), module.CourseID, lessonWithProgress.ModulePosition, lessonWithProgress.Position)
		nextLesson, _ := h.services.Enrollment.GetNextLesson(c.Request().Context(), module.CourseID, lessonWithProgress.ModulePosition, lessonWithProgress.Position)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"lesson":          lessonWithProgress,
			"enrollment_id":   enrollment.ID,
			"previous_lesson": prevLesson,
			"next_lesson":     nextLesson,
		})
	}

	// Free preview - return lesson without progress
	lessonWithVideo, err := h.services.Course.GetLessonWithVideo(c.Request().Context(), lessonID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "lesson not found"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"lesson":       lessonWithVideo,
		"is_preview":   true,
		"course_id":    module.CourseID,
		"course_title": "",
	})
}

// MarkLessonComplete marks a lesson as completed
func (h *Handler) MarkLessonComplete(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	user := GetUserFromContext(c)

	if tenant == nil || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	lessonID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid lesson id"})
	}

	// Get lesson to find course
	lesson, err := h.services.Course.GetLessonByID(c.Request().Context(), lessonID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "lesson not found"})
	}

	module, err := h.services.Course.GetModuleByID(c.Request().Context(), lesson.ModuleID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "module not found"})
	}

	// Get enrollment
	enrollment, err := h.services.Enrollment.GetEnrollment(c.Request().Context(), user.ID, module.CourseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "not enrolled in this course"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get enrollment"})
	}

	progress, err := h.services.Enrollment.MarkLessonComplete(c.Request().Context(), tenant.ID, enrollment.ID, lessonID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to mark lesson complete"})
	}

	// Get updated enrollment to return progress
	updatedEnrollment, _ := h.services.Enrollment.GetEnrollmentByID(c.Request().Context(), enrollment.ID)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"progress":            progress,
		"enrollment_progress": updatedEnrollment.ProgressPercentage,
		"completed_count":     updatedEnrollment.CompletedLessonsCount,
		"total_count":         updatedEnrollment.TotalLessonsCount,
	})
}

// UnmarkLessonComplete marks a lesson as incomplete
func (h *Handler) UnmarkLessonComplete(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	user := GetUserFromContext(c)

	if tenant == nil || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	lessonID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid lesson id"})
	}

	// Get lesson to find course
	lesson, err := h.services.Course.GetLessonByID(c.Request().Context(), lessonID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "lesson not found"})
	}

	module, err := h.services.Course.GetModuleByID(c.Request().Context(), lesson.ModuleID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "module not found"})
	}

	// Get enrollment
	enrollment, err := h.services.Enrollment.GetEnrollment(c.Request().Context(), user.ID, module.CourseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "not enrolled in this course"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get enrollment"})
	}

	progress, err := h.services.Enrollment.MarkLessonIncomplete(c.Request().Context(), enrollment.ID, lessonID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrorResponse{Error: "no progress found for this lesson"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to unmark lesson complete"})
	}

	// Get updated enrollment to return progress
	updatedEnrollment, _ := h.services.Enrollment.GetEnrollmentByID(c.Request().Context(), enrollment.ID)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"progress":            progress,
		"enrollment_progress": updatedEnrollment.ProgressPercentage,
		"completed_count":     updatedEnrollment.CompletedLessonsCount,
		"total_count":         updatedEnrollment.TotalLessonsCount,
	})
}

// UpdateLessonVideoProgress updates video watch progress
func (h *Handler) UpdateLessonVideoProgress(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	user := GetUserFromContext(c)

	if tenant == nil || user == nil {
		return c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "unauthorized"})
	}

	lessonID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid lesson id"})
	}

	var req UpdateVideoProgressRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
	}

	// Get lesson to find course
	lesson, err := h.services.Course.GetLessonByID(c.Request().Context(), lessonID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "lesson not found"})
	}

	module, err := h.services.Course.GetModuleByID(c.Request().Context(), lesson.ModuleID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "module not found"})
	}

	// Get enrollment
	enrollment, err := h.services.Enrollment.GetEnrollment(c.Request().Context(), user.ID, module.CourseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return c.JSON(http.StatusForbidden, ErrorResponse{Error: "not enrolled in this course"})
		}
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to get enrollment"})
	}

	progress, err := h.services.Enrollment.UpdateVideoProgress(
		c.Request().Context(),
		tenant.ID,
		enrollment.ID,
		lessonID,
		req.WatchDurationSeconds,
		req.VideoTotalSeconds,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update video progress"})
	}

	return c.JSON(http.StatusOK, progress)
}

// ============================================================================
// ADMIN HANDLERS
// ============================================================================

// ListCourseEnrollments lists all enrollments for a course (admin)
func (h *Handler) ListCourseEnrollments(c echo.Context) error {
	tenant := GetTenantFromContext(c)
	if tenant == nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "tenant context required"})
	}

	courseID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid course id"})
	}

	// Verify course belongs to tenant
	course, err := h.services.Course.GetByID(c.Request().Context(), courseID)
	if err != nil {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "course not found"})
	}

	if course.TenantID != tenant.ID {
		return c.JSON(http.StatusNotFound, ErrorResponse{Error: "course not found"})
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, _ := strconv.Atoi(c.QueryParam("offset"))
	if offset < 0 {
		offset = 0
	}

	enrollments, err := h.services.Enrollment.ListEnrollmentsByCourse(c.Request().Context(), courseID, int32(limit), int32(offset))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to list enrollments"})
	}

	count, err := h.services.Enrollment.CountEnrollmentsByCourse(c.Request().Context(), courseID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to count enrollments"})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"enrollments": enrollments,
		"total":       count,
		"limit":       limit,
		"offset":      offset,
	})
}
