package service

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/gosimple/slug"

	"github.com/nickkcj/orbit-backend/internal/database"
)

type CourseService struct {
	db *database.Queries
}

func NewCourseService(db *database.Queries) *CourseService {
	return &CourseService{db: db}
}

// ============================================================================
// COURSE OPERATIONS
// ============================================================================

type CreateCourseInput struct {
	TenantID     uuid.UUID
	AuthorID     uuid.UUID
	Title        string
	Description  string
	ThumbnailURL string
}

func (s *CourseService) Create(ctx context.Context, input CreateCourseInput) (database.Course, error) {
	courseSlug := slug.Make(input.Title)

	return s.db.CreateCourse(ctx, database.CreateCourseParams{
		TenantID:     input.TenantID,
		AuthorID:     input.AuthorID,
		Title:        input.Title,
		Slug:         courseSlug,
		Description:  sql.NullString{String: input.Description, Valid: input.Description != ""},
		ThumbnailUrl: sql.NullString{String: input.ThumbnailURL, Valid: input.ThumbnailURL != ""},
	})
}

func (s *CourseService) GetByID(ctx context.Context, id uuid.UUID) (database.Course, error) {
	return s.db.GetCourseByID(ctx, id)
}

func (s *CourseService) GetBySlug(ctx context.Context, tenantID uuid.UUID, courseSlug string) (database.Course, error) {
	return s.db.GetCourseBySlug(ctx, database.GetCourseBySlugParams{
		TenantID: tenantID,
		Slug:     courseSlug,
	})
}

func (s *CourseService) GetWithDetails(ctx context.Context, id uuid.UUID) (database.GetCourseWithDetailsRow, error) {
	return s.db.GetCourseWithDetails(ctx, id)
}

func (s *CourseService) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]database.ListCoursesByTenantRow, error) {
	return s.db.ListCoursesByTenant(ctx, database.ListCoursesByTenantParams{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	})
}

func (s *CourseService) ListAllByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]database.ListAllCoursesByTenantRow, error) {
	return s.db.ListAllCoursesByTenant(ctx, database.ListAllCoursesByTenantParams{
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	})
}

func (s *CourseService) ListByAuthor(ctx context.Context, tenantID, authorID uuid.UUID) ([]database.Course, error) {
	return s.db.ListCoursesByAuthor(ctx, database.ListCoursesByAuthorParams{
		TenantID: tenantID,
		AuthorID: authorID,
	})
}

type UpdateCourseInput struct {
	Title        string
	Description  string
	ThumbnailURL string
}

func (s *CourseService) Update(ctx context.Context, id uuid.UUID, input UpdateCourseInput) (database.Course, error) {
	return s.db.UpdateCourse(ctx, database.UpdateCourseParams{
		ID:           id,
		Title:        input.Title,
		Description:  sql.NullString{String: input.Description, Valid: input.Description != ""},
		ThumbnailUrl: sql.NullString{String: input.ThumbnailURL, Valid: input.ThumbnailURL != ""},
	})
}

func (s *CourseService) Publish(ctx context.Context, id uuid.UUID) (database.Course, error) {
	return s.db.PublishCourse(ctx, id)
}

func (s *CourseService) Unpublish(ctx context.Context, id uuid.UUID) (database.Course, error) {
	return s.db.UnpublishCourse(ctx, id)
}

func (s *CourseService) Archive(ctx context.Context, id uuid.UUID) error {
	return s.db.ArchiveCourse(ctx, id)
}

func (s *CourseService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.DeleteCourse(ctx, id)
}

// ============================================================================
// MODULE OPERATIONS
// ============================================================================

type CreateModuleInput struct {
	TenantID    uuid.UUID
	CourseID    uuid.UUID
	Title       string
	Description string
}

func (s *CourseService) CreateModule(ctx context.Context, input CreateModuleInput) (database.Module, error) {
	// Get next position
	maxPos, err := s.db.GetMaxModulePosition(ctx, input.CourseID)
	if err != nil {
		return database.Module{}, err
	}

	return s.db.CreateModule(ctx, database.CreateModuleParams{
		TenantID:    input.TenantID,
		CourseID:    input.CourseID,
		Title:       input.Title,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		Position:    maxPos + 1,
	})
}

func (s *CourseService) GetModuleByID(ctx context.Context, id uuid.UUID) (database.Module, error) {
	return s.db.GetModuleByID(ctx, id)
}

func (s *CourseService) ListModulesByCourse(ctx context.Context, courseID uuid.UUID) ([]database.Module, error) {
	return s.db.ListModulesByCourse(ctx, courseID)
}

type UpdateModuleInput struct {
	Title       string
	Description string
}

func (s *CourseService) UpdateModule(ctx context.Context, id uuid.UUID, input UpdateModuleInput) (database.Module, error) {
	return s.db.UpdateModule(ctx, database.UpdateModuleParams{
		ID:          id,
		Title:       input.Title,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
	})
}

func (s *CourseService) ReorderModule(ctx context.Context, moduleID uuid.UUID, newPosition int32) error {
	module, err := s.db.GetModuleByID(ctx, moduleID)
	if err != nil {
		return err
	}

	oldPosition := module.Position

	if newPosition == oldPosition {
		return nil
	}

	if newPosition < oldPosition {
		// Moving up: shift modules between newPosition and oldPosition down
		if err := s.db.ShiftModulePositionsUp(ctx, database.ShiftModulePositionsUpParams{
			CourseID: module.CourseID,
			Position: newPosition,
		}); err != nil {
			return err
		}
	} else {
		// Moving down: shift modules between oldPosition and newPosition up
		if err := s.db.ShiftModulePositionsDown(ctx, database.ShiftModulePositionsDownParams{
			CourseID:   module.CourseID,
			Position:   oldPosition,
			Position_2: newPosition,
		}); err != nil {
			return err
		}
	}

	return s.db.UpdateModulePosition(ctx, database.UpdateModulePositionParams{
		ID:       moduleID,
		Position: newPosition,
	})
}

func (s *CourseService) DeleteModule(ctx context.Context, id uuid.UUID) error {
	module, err := s.db.GetModuleByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.db.DeleteModule(ctx, id); err != nil {
		return err
	}

	// Reorder remaining modules
	return s.db.ReorderModulesAfterDelete(ctx, database.ReorderModulesAfterDeleteParams{
		CourseID: module.CourseID,
		Position: module.Position,
	})
}

// ============================================================================
// LESSON OPERATIONS
// ============================================================================

type CreateLessonInput struct {
	TenantID        uuid.UUID
	ModuleID        uuid.UUID
	Title           string
	Description     string
	Content         string
	ContentFormat   string
	VideoID         *uuid.UUID
	DurationMinutes int32
	IsFreePreview   bool
}

func (s *CourseService) CreateLesson(ctx context.Context, input CreateLessonInput) (database.Lesson, error) {
	// Get next position
	maxPos, err := s.db.GetMaxLessonPosition(ctx, input.ModuleID)
	if err != nil {
		return database.Lesson{}, err
	}

	contentFormat := input.ContentFormat
	if contentFormat == "" {
		contentFormat = "markdown"
	}

	videoID := uuid.NullUUID{}
	if input.VideoID != nil {
		videoID = uuid.NullUUID{UUID: *input.VideoID, Valid: true}
	}

	return s.db.CreateLesson(ctx, database.CreateLessonParams{
		TenantID:        input.TenantID,
		ModuleID:        input.ModuleID,
		Title:           input.Title,
		Description:     sql.NullString{String: input.Description, Valid: input.Description != ""},
		Content:         sql.NullString{String: input.Content, Valid: input.Content != ""},
		ContentFormat:   contentFormat,
		VideoID:         videoID,
		Position:        maxPos + 1,
		DurationMinutes: sql.NullInt32{Int32: input.DurationMinutes, Valid: input.DurationMinutes > 0},
		IsFreePreview:   input.IsFreePreview,
	})
}

func (s *CourseService) GetLessonByID(ctx context.Context, id uuid.UUID) (database.Lesson, error) {
	return s.db.GetLessonByID(ctx, id)
}

func (s *CourseService) GetLessonWithVideo(ctx context.Context, id uuid.UUID) (database.GetLessonWithVideoRow, error) {
	return s.db.GetLessonWithVideo(ctx, id)
}

func (s *CourseService) ListLessonsByModule(ctx context.Context, moduleID uuid.UUID) ([]database.ListLessonsByModuleRow, error) {
	return s.db.ListLessonsByModule(ctx, moduleID)
}

func (s *CourseService) ListLessonsByCourse(ctx context.Context, courseID uuid.UUID) ([]database.ListLessonsByCourseRow, error) {
	return s.db.ListLessonsByCourse(ctx, courseID)
}

type UpdateLessonInput struct {
	Title           string
	Description     string
	Content         string
	ContentFormat   string
	VideoID         *uuid.UUID
	DurationMinutes int32
	IsFreePreview   bool
}

func (s *CourseService) UpdateLesson(ctx context.Context, id uuid.UUID, input UpdateLessonInput) (database.Lesson, error) {
	videoID := uuid.NullUUID{}
	if input.VideoID != nil {
		videoID = uuid.NullUUID{UUID: *input.VideoID, Valid: true}
	}

	contentFormat := input.ContentFormat
	if contentFormat == "" {
		contentFormat = "markdown"
	}

	return s.db.UpdateLesson(ctx, database.UpdateLessonParams{
		ID:              id,
		Title:           input.Title,
		Description:     sql.NullString{String: input.Description, Valid: input.Description != ""},
		Content:         sql.NullString{String: input.Content, Valid: input.Content != ""},
		ContentFormat:   contentFormat,
		VideoID:         videoID,
		DurationMinutes: sql.NullInt32{Int32: input.DurationMinutes, Valid: input.DurationMinutes > 0},
		IsFreePreview:   input.IsFreePreview,
	})
}

func (s *CourseService) ReorderLesson(ctx context.Context, lessonID uuid.UUID, newPosition int32) error {
	lesson, err := s.db.GetLessonByID(ctx, lessonID)
	if err != nil {
		return err
	}

	oldPosition := lesson.Position

	if newPosition == oldPosition {
		return nil
	}

	if newPosition < oldPosition {
		// Moving up
		if err := s.db.ShiftLessonPositionsUp(ctx, database.ShiftLessonPositionsUpParams{
			ModuleID: lesson.ModuleID,
			Position: newPosition,
		}); err != nil {
			return err
		}
	} else {
		// Moving down
		if err := s.db.ShiftLessonPositionsDown(ctx, database.ShiftLessonPositionsDownParams{
			ModuleID:   lesson.ModuleID,
			Position:   oldPosition,
			Position_2: newPosition,
		}); err != nil {
			return err
		}
	}

	return s.db.UpdateLessonPosition(ctx, database.UpdateLessonPositionParams{
		ID:       lessonID,
		Position: newPosition,
	})
}

func (s *CourseService) DeleteLesson(ctx context.Context, id uuid.UUID) error {
	lesson, err := s.db.GetLessonByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.db.DeleteLesson(ctx, id); err != nil {
		return err
	}

	// Reorder remaining lessons
	return s.db.ReorderLessonsAfterDelete(ctx, database.ReorderLessonsAfterDeleteParams{
		ModuleID: lesson.ModuleID,
		Position: lesson.Position,
	})
}

func (s *CourseService) AttachVideoToLesson(ctx context.Context, lessonID, videoID uuid.UUID) error {
	return s.db.AttachVideoToLesson(ctx, database.AttachVideoToLessonParams{
		ID:      lessonID,
		VideoID: uuid.NullUUID{UUID: videoID, Valid: true},
	})
}

func (s *CourseService) DetachVideoFromLesson(ctx context.Context, lessonID uuid.UUID) error {
	return s.db.DetachVideoFromLesson(ctx, lessonID)
}

// ============================================================================
// FULL COURSE STRUCTURE
// ============================================================================

type CourseStructure struct {
	Course  database.GetCourseWithDetailsRow `json:"course"`
	Modules []ModuleWithLessons              `json:"modules"`
}

type ModuleWithLessons struct {
	Module  database.Module                  `json:"module"`
	Lessons []database.ListLessonsByModuleRow `json:"lessons"`
}

func (s *CourseService) GetCourseStructure(ctx context.Context, courseID uuid.UUID) (*CourseStructure, error) {
	course, err := s.db.GetCourseWithDetails(ctx, courseID)
	if err != nil {
		return nil, err
	}

	modules, err := s.db.ListModulesByCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}

	structure := &CourseStructure{
		Course:  course,
		Modules: make([]ModuleWithLessons, len(modules)),
	}

	for i, module := range modules {
		lessons, err := s.db.ListLessonsByModule(ctx, module.ID)
		if err != nil {
			return nil, err
		}
		structure.Modules[i] = ModuleWithLessons{
			Module:  module,
			Lessons: lessons,
		}
	}

	return structure, nil
}
