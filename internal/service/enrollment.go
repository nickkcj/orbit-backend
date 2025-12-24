package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"

	"github.com/nickkcj/orbit-backend/internal/database"
)

var (
	ErrAlreadyEnrolled  = errors.New("user is already enrolled in this course")
	ErrNotEnrolled      = errors.New("user is not enrolled in this course")
	ErrCourseNotFound   = errors.New("course not found")
	ErrLessonNotFound   = errors.New("lesson not found")
	ErrAccessDenied     = errors.New("access denied to this lesson")
	ErrEnrollmentFailed = errors.New("failed to create enrollment")
)

type EnrollmentService struct {
	db *database.Queries
}

func NewEnrollmentService(db *database.Queries) *EnrollmentService {
	return &EnrollmentService{db: db}
}

// ============================================================================
// ENROLLMENT OPERATIONS
// ============================================================================

type EnrollInput struct {
	TenantID uuid.UUID
	UserID   uuid.UUID
	CourseID uuid.UUID
}

func (s *EnrollmentService) Enroll(ctx context.Context, input EnrollInput) (database.CourseEnrollment, error) {
	// Check if already enrolled
	enrolled, err := s.db.IsUserEnrolled(ctx, database.IsUserEnrolledParams{
		UserID:   input.UserID,
		CourseID: input.CourseID,
	})
	if err != nil {
		return database.CourseEnrollment{}, err
	}
	if enrolled {
		return database.CourseEnrollment{}, ErrAlreadyEnrolled
	}

	// Create enrollment
	enrollment, err := s.db.CreateEnrollment(ctx, database.CreateEnrollmentParams{
		TenantID: input.TenantID,
		UserID:   input.UserID,
		CourseID: input.CourseID,
	})
	if err != nil {
		return database.CourseEnrollment{}, err
	}

	return enrollment, nil
}

func (s *EnrollmentService) IsEnrolled(ctx context.Context, userID, courseID uuid.UUID) (bool, error) {
	return s.db.IsUserEnrolled(ctx, database.IsUserEnrolledParams{
		UserID:   userID,
		CourseID: courseID,
	})
}

func (s *EnrollmentService) GetEnrollment(ctx context.Context, userID, courseID uuid.UUID) (database.CourseEnrollment, error) {
	return s.db.GetEnrollmentByUserAndCourse(ctx, database.GetEnrollmentByUserAndCourseParams{
		UserID:   userID,
		CourseID: courseID,
	})
}

func (s *EnrollmentService) GetEnrollmentByID(ctx context.Context, id uuid.UUID) (database.CourseEnrollment, error) {
	return s.db.GetEnrollmentByID(ctx, id)
}

func (s *EnrollmentService) GetEnrollmentWithProgress(ctx context.Context, id uuid.UUID) (database.GetEnrollmentWithProgressRow, error) {
	return s.db.GetEnrollmentWithProgress(ctx, id)
}

func (s *EnrollmentService) GetEnrollmentWithProgressByUserAndCourse(ctx context.Context, userID, courseID uuid.UUID) (database.GetEnrollmentWithProgressByUserAndCourseRow, error) {
	return s.db.GetEnrollmentWithProgressByUserAndCourse(ctx, database.GetEnrollmentWithProgressByUserAndCourseParams{
		UserID:   userID,
		CourseID: courseID,
	})
}

func (s *EnrollmentService) ListEnrollmentsByUser(ctx context.Context, userID, tenantID uuid.UUID, limit, offset int32) ([]database.ListEnrollmentsByUserRow, error) {
	return s.db.ListEnrollmentsByUser(ctx, database.ListEnrollmentsByUserParams{
		UserID:   userID,
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	})
}

func (s *EnrollmentService) CountEnrollmentsByUser(ctx context.Context, userID, tenantID uuid.UUID) (int64, error) {
	return s.db.CountEnrollmentsByUser(ctx, database.CountEnrollmentsByUserParams{
		UserID:   userID,
		TenantID: tenantID,
	})
}

func (s *EnrollmentService) ListEnrollmentsByCourse(ctx context.Context, courseID uuid.UUID, limit, offset int32) ([]database.ListEnrollmentsByCourseRow, error) {
	return s.db.ListEnrollmentsByCourse(ctx, database.ListEnrollmentsByCourseParams{
		CourseID: courseID,
		Limit:    limit,
		Offset:   offset,
	})
}

func (s *EnrollmentService) CountEnrollmentsByCourse(ctx context.Context, courseID uuid.UUID) (int64, error) {
	return s.db.CountEnrollmentsByCourse(ctx, courseID)
}

func (s *EnrollmentService) GetContinueLearning(ctx context.Context, userID, tenantID uuid.UUID, limit int32) ([]database.GetContinueLearningCoursesRow, error) {
	return s.db.GetContinueLearningCourses(ctx, database.GetContinueLearningCoursesParams{
		UserID:   userID,
		TenantID: tenantID,
		Limit:    limit,
	})
}

func (s *EnrollmentService) GetCompletedCourses(ctx context.Context, userID, tenantID uuid.UUID, limit, offset int32) ([]database.GetCompletedCoursesRow, error) {
	return s.db.GetCompletedCourses(ctx, database.GetCompletedCoursesParams{
		UserID:   userID,
		TenantID: tenantID,
		Limit:    limit,
		Offset:   offset,
	})
}

func (s *EnrollmentService) DropEnrollment(ctx context.Context, enrollmentID uuid.UUID) error {
	return s.db.DropEnrollment(ctx, enrollmentID)
}

func (s *EnrollmentService) DeleteEnrollment(ctx context.Context, enrollmentID uuid.UUID) error {
	return s.db.DeleteEnrollment(ctx, enrollmentID)
}

// ============================================================================
// LESSON PROGRESS OPERATIONS
// ============================================================================

func (s *EnrollmentService) MarkLessonComplete(ctx context.Context, tenantID, enrollmentID, lessonID uuid.UUID) (database.LessonProgress, error) {
	return s.db.MarkLessonComplete(ctx, database.MarkLessonCompleteParams{
		TenantID:     tenantID,
		EnrollmentID: enrollmentID,
		LessonID:     lessonID,
		Column4:      nil, // started_at will use NOW()
	})
}

func (s *EnrollmentService) MarkLessonIncomplete(ctx context.Context, enrollmentID, lessonID uuid.UUID) (database.LessonProgress, error) {
	return s.db.MarkLessonIncomplete(ctx, database.MarkLessonIncompleteParams{
		EnrollmentID: enrollmentID,
		LessonID:     lessonID,
	})
}

func (s *EnrollmentService) UpdateVideoProgress(ctx context.Context, tenantID, enrollmentID, lessonID uuid.UUID, watchDuration int32, videoTotal *int32) (database.LessonProgress, error) {
	videoTotalNull := sql.NullInt32{}
	if videoTotal != nil {
		videoTotalNull = sql.NullInt32{Int32: *videoTotal, Valid: true}
	}

	return s.db.UpdateVideoProgress(ctx, database.UpdateVideoProgressParams{
		TenantID:             tenantID,
		EnrollmentID:         enrollmentID,
		LessonID:             lessonID,
		WatchDurationSeconds: watchDuration,
		VideoTotalSeconds:    videoTotalNull,
	})
}

func (s *EnrollmentService) GetLessonProgress(ctx context.Context, enrollmentID, lessonID uuid.UUID) (database.LessonProgress, error) {
	return s.db.GetLessonProgress(ctx, database.GetLessonProgressParams{
		EnrollmentID: enrollmentID,
		LessonID:     lessonID,
	})
}

func (s *EnrollmentService) ListLessonProgressByEnrollment(ctx context.Context, enrollmentID uuid.UUID) ([]database.LessonProgress, error) {
	return s.db.ListLessonProgressByEnrollment(ctx, enrollmentID)
}

func (s *EnrollmentService) GetLessonWithProgressAndCourse(ctx context.Context, lessonID, enrollmentID uuid.UUID) (database.GetLessonWithProgressAndCourseRow, error) {
	return s.db.GetLessonWithProgressAndCourse(ctx, database.GetLessonWithProgressAndCourseParams{
		ID:           lessonID,
		EnrollmentID: enrollmentID,
	})
}

func (s *EnrollmentService) ListLessonsWithProgress(ctx context.Context, courseID, enrollmentID uuid.UUID) ([]database.ListLessonsWithProgressRow, error) {
	return s.db.ListLessonsWithProgress(ctx, database.ListLessonsWithProgressParams{
		CourseID:     courseID,
		EnrollmentID: enrollmentID,
	})
}

func (s *EnrollmentService) GetNextIncompleteLesson(ctx context.Context, courseID, enrollmentID uuid.UUID) (database.GetNextIncompleteLessonRow, error) {
	return s.db.GetNextIncompleteLesson(ctx, database.GetNextIncompleteLessonParams{
		CourseID:     courseID,
		EnrollmentID: enrollmentID,
	})
}

func (s *EnrollmentService) GetFirstLesson(ctx context.Context, courseID uuid.UUID) (database.GetFirstLessonRow, error) {
	return s.db.GetFirstLesson(ctx, courseID)
}

func (s *EnrollmentService) GetPreviousLesson(ctx context.Context, courseID uuid.UUID, modulePosition, lessonPosition int32) (database.GetPreviousLessonRow, error) {
	return s.db.GetPreviousLesson(ctx, database.GetPreviousLessonParams{
		CourseID:   courseID,
		Position:   modulePosition,
		Position_2: lessonPosition,
	})
}

func (s *EnrollmentService) GetNextLesson(ctx context.Context, courseID uuid.UUID, modulePosition, lessonPosition int32) (database.GetNextLessonRow, error) {
	return s.db.GetNextLesson(ctx, database.GetNextLessonParams{
		CourseID:   courseID,
		Position:   modulePosition,
		Position_2: lessonPosition,
	})
}

// ============================================================================
// ACCESS CONTROL
// ============================================================================

// CanAccessLesson checks if a user can access a lesson
// Returns true if user is enrolled OR lesson is marked as free preview
func (s *EnrollmentService) CanAccessLesson(ctx context.Context, userID, courseID, lessonID uuid.UUID) (bool, error) {
	// Check if user is enrolled
	enrolled, err := s.IsEnrolled(ctx, userID, courseID)
	if err != nil {
		return false, err
	}
	if enrolled {
		return true, nil
	}

	// Check if lesson is free preview
	lesson, err := s.db.GetLessonByID(ctx, lessonID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrLessonNotFound
		}
		return false, err
	}

	return lesson.IsFreePreview, nil
}

// UpdateLastAccessed updates the last accessed lesson for an enrollment
func (s *EnrollmentService) UpdateLastAccessed(ctx context.Context, enrollmentID, lessonID uuid.UUID) error {
	return s.db.UpdateEnrollmentLastAccessed(ctx, database.UpdateEnrollmentLastAccessedParams{
		ID:           enrollmentID,
		LastLessonID: uuid.NullUUID{UUID: lessonID, Valid: true},
	})
}

// ============================================================================
// COURSE PLAYER DATA
// ============================================================================

type CoursePlayerData struct {
	Enrollment database.GetEnrollmentWithProgressRow `json:"enrollment"`
	Modules    []ModuleWithLessonsProgress           `json:"modules"`
	NextLesson *database.GetNextIncompleteLessonRow  `json:"next_lesson,omitempty"`
}

type ModuleWithLessonsProgress struct {
	ModuleID    uuid.UUID                            `json:"module_id"`
	ModuleTitle string                               `json:"module_title"`
	Position    int32                                `json:"position"`
	Lessons     []database.ListLessonsWithProgressRow `json:"lessons"`
}

func (s *EnrollmentService) GetCoursePlayerData(ctx context.Context, enrollmentID uuid.UUID) (*CoursePlayerData, error) {
	// Get enrollment with progress
	enrollment, err := s.db.GetEnrollmentWithProgress(ctx, enrollmentID)
	if err != nil {
		return nil, err
	}

	// Get lessons with progress
	lessons, err := s.db.ListLessonsWithProgress(ctx, database.ListLessonsWithProgressParams{
		CourseID:     enrollment.CourseID,
		EnrollmentID: enrollmentID,
	})
	if err != nil {
		return nil, err
	}

	// Group lessons by module
	moduleMap := make(map[uuid.UUID]*ModuleWithLessonsProgress)
	var moduleOrder []uuid.UUID

	for _, lesson := range lessons {
		if _, exists := moduleMap[lesson.ModuleID]; !exists {
			moduleMap[lesson.ModuleID] = &ModuleWithLessonsProgress{
				ModuleID:    lesson.ModuleID,
				ModuleTitle: lesson.ModuleTitle,
				Position:    lesson.ModulePosition,
				Lessons:     []database.ListLessonsWithProgressRow{},
			}
			moduleOrder = append(moduleOrder, lesson.ModuleID)
		}
		moduleMap[lesson.ModuleID].Lessons = append(moduleMap[lesson.ModuleID].Lessons, lesson)
	}

	// Build ordered modules slice
	modules := make([]ModuleWithLessonsProgress, len(moduleOrder))
	for i, id := range moduleOrder {
		modules[i] = *moduleMap[id]
	}

	// Get next incomplete lesson
	var nextLesson *database.GetNextIncompleteLessonRow
	next, err := s.db.GetNextIncompleteLesson(ctx, database.GetNextIncompleteLessonParams{
		CourseID:     enrollment.CourseID,
		EnrollmentID: enrollmentID,
	})
	if err == nil {
		nextLesson = &next
	}

	return &CoursePlayerData{
		Enrollment: enrollment,
		Modules:    modules,
		NextLesson: nextLesson,
	}, nil
}
