-- +goose Up
-- ============================================================================
-- ORBIT BACKEND - Enrollments & Progress Schema (V2)
-- User enrollment in courses and lesson progress tracking
-- ============================================================================

-- ============================================================================
-- ENROLLMENTS: User enrollment in courses
-- ============================================================================

CREATE TABLE course_enrollments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'completed', 'dropped')),

    -- Progress tracking
    progress_percentage INT NOT NULL DEFAULT 0 CHECK (progress_percentage >= 0 AND progress_percentage <= 100),
    completed_lessons_count INT NOT NULL DEFAULT 0,
    total_lessons_count INT NOT NULL DEFAULT 0,

    -- Resume functionality
    last_lesson_id UUID REFERENCES lessons(id) ON DELETE SET NULL,
    last_accessed_at TIMESTAMPTZ,

    -- Timestamps
    enrolled_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- One enrollment per user per course
    UNIQUE(user_id, course_id)
);

CREATE INDEX idx_enrollments_tenant ON course_enrollments(tenant_id);
CREATE INDEX idx_enrollments_user ON course_enrollments(user_id);
CREATE INDEX idx_enrollments_course ON course_enrollments(course_id);
CREATE INDEX idx_enrollments_status ON course_enrollments(status);
CREATE INDEX idx_enrollments_user_active ON course_enrollments(user_id, status) WHERE status = 'active';
CREATE INDEX idx_enrollments_last_accessed ON course_enrollments(user_id, last_accessed_at DESC) WHERE status = 'active';

-- ============================================================================
-- LESSON PROGRESS: Track user progress on individual lessons
-- ============================================================================

CREATE TABLE lesson_progress (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    enrollment_id UUID NOT NULL REFERENCES course_enrollments(id) ON DELETE CASCADE,
    lesson_id UUID NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,

    -- Progress status
    status VARCHAR(20) NOT NULL DEFAULT 'not_started' CHECK (status IN ('not_started', 'in_progress', 'completed')),

    -- Video progress (for video lessons)
    watch_duration_seconds INT NOT NULL DEFAULT 0,
    video_total_seconds INT,

    -- Timestamps
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- One progress record per lesson per enrollment
    UNIQUE(enrollment_id, lesson_id)
);

CREATE INDEX idx_lesson_progress_tenant ON lesson_progress(tenant_id);
CREATE INDEX idx_lesson_progress_enrollment ON lesson_progress(enrollment_id);
CREATE INDEX idx_lesson_progress_lesson ON lesson_progress(lesson_id);
CREATE INDEX idx_lesson_progress_status ON lesson_progress(enrollment_id, status);

-- ============================================================================
-- TRIGGERS: Auto-update timestamps
-- ============================================================================

CREATE TRIGGER update_enrollments_updated_at BEFORE UPDATE ON course_enrollments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_lesson_progress_updated_at BEFORE UPDATE ON lesson_progress FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- TRIGGER: Initialize total_lessons_count on enrollment
-- ============================================================================

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION initialize_enrollment_lesson_count()
RETURNS TRIGGER AS $$
BEGIN
    SELECT lesson_count INTO NEW.total_lessons_count
    FROM courses
    WHERE id = NEW.course_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER initialize_enrollment_lesson_count_trigger
    BEFORE INSERT ON course_enrollments
    FOR EACH ROW EXECUTE FUNCTION initialize_enrollment_lesson_count();

-- ============================================================================
-- TRIGGER: Update enrollment progress when lesson is completed
-- ============================================================================

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_enrollment_progress()
RETURNS TRIGGER AS $$
DECLARE
    v_enrollment course_enrollments%ROWTYPE;
    v_new_completed_count INT;
    v_new_percentage INT;
BEGIN
    -- Get the enrollment
    SELECT * INTO v_enrollment FROM course_enrollments WHERE id = NEW.enrollment_id;

    IF TG_OP = 'INSERT' OR TG_OP = 'UPDATE' THEN
        -- Count completed lessons for this enrollment
        SELECT COUNT(*) INTO v_new_completed_count
        FROM lesson_progress
        WHERE enrollment_id = NEW.enrollment_id AND status = 'completed';

        -- Calculate percentage
        IF v_enrollment.total_lessons_count > 0 THEN
            v_new_percentage := ROUND((v_new_completed_count::DECIMAL / v_enrollment.total_lessons_count) * 100);
        ELSE
            v_new_percentage := 0;
        END IF;

        -- Update enrollment
        UPDATE course_enrollments
        SET
            completed_lessons_count = v_new_completed_count,
            progress_percentage = v_new_percentage,
            status = CASE
                WHEN v_new_percentage >= 100 THEN 'completed'
                ELSE status
            END,
            completed_at = CASE
                WHEN v_new_percentage >= 100 AND completed_at IS NULL THEN NOW()
                ELSE completed_at
            END,
            last_lesson_id = NEW.lesson_id,
            last_accessed_at = NOW()
        WHERE id = NEW.enrollment_id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER update_enrollment_progress_trigger
    AFTER INSERT OR UPDATE ON lesson_progress
    FOR EACH ROW EXECUTE FUNCTION update_enrollment_progress();

-- ============================================================================
-- TRIGGER: Recalculate progress when lesson is deleted from progress
-- ============================================================================

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION recalculate_enrollment_on_progress_delete()
RETURNS TRIGGER AS $$
DECLARE
    v_enrollment course_enrollments%ROWTYPE;
    v_new_completed_count INT;
    v_new_percentage INT;
BEGIN
    -- Get the enrollment
    SELECT * INTO v_enrollment FROM course_enrollments WHERE id = OLD.enrollment_id;

    -- Count completed lessons for this enrollment
    SELECT COUNT(*) INTO v_new_completed_count
    FROM lesson_progress
    WHERE enrollment_id = OLD.enrollment_id AND status = 'completed';

    -- Calculate percentage
    IF v_enrollment.total_lessons_count > 0 THEN
        v_new_percentage := ROUND((v_new_completed_count::DECIMAL / v_enrollment.total_lessons_count) * 100);
    ELSE
        v_new_percentage := 0;
    END IF;

    -- Update enrollment
    UPDATE course_enrollments
    SET
        completed_lessons_count = v_new_completed_count,
        progress_percentage = v_new_percentage,
        status = CASE
            WHEN v_new_percentage < 100 AND status = 'completed' THEN 'active'
            ELSE status
        END,
        completed_at = CASE
            WHEN v_new_percentage < 100 THEN NULL
            ELSE completed_at
        END
    WHERE id = OLD.enrollment_id;

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER recalculate_enrollment_on_progress_delete_trigger
    AFTER DELETE ON lesson_progress
    FOR EACH ROW EXECUTE FUNCTION recalculate_enrollment_on_progress_delete();

-- ============================================================================
-- PERMISSIONS: Enrollment-related permissions
-- ============================================================================

INSERT INTO permissions (code, name, description, category) VALUES
    ('enrollments.view', 'Ver matrículas', 'Visualizar próprias matrículas', 'enrollments'),
    ('enrollments.enroll', 'Matricular-se', 'Matricular-se em cursos', 'enrollments'),
    ('enrollments.manage', 'Gerenciar matrículas', 'Gerenciar matrículas de todos os usuários', 'enrollments');

-- ============================================================================
-- ROLE PERMISSIONS: Add enrollment permissions to existing roles
-- ============================================================================

-- Owner: All enrollment permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.slug = 'owner' AND r.is_system = TRUE
AND p.code IN ('enrollments.view', 'enrollments.enroll', 'enrollments.manage')
ON CONFLICT DO NOTHING;

-- Admin: All enrollment permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.slug = 'admin' AND r.is_system = TRUE
AND p.code IN ('enrollments.view', 'enrollments.enroll', 'enrollments.manage')
ON CONFLICT DO NOTHING;

-- Member: View and enroll only
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.slug = 'member' AND r.is_system = TRUE
AND p.code IN ('enrollments.view', 'enrollments.enroll')
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TRIGGER IF EXISTS recalculate_enrollment_on_progress_delete_trigger ON lesson_progress;
DROP FUNCTION IF EXISTS recalculate_enrollment_on_progress_delete();
DROP TRIGGER IF EXISTS update_enrollment_progress_trigger ON lesson_progress;
DROP FUNCTION IF EXISTS update_enrollment_progress();
DROP TRIGGER IF EXISTS initialize_enrollment_lesson_count_trigger ON course_enrollments;
DROP FUNCTION IF EXISTS initialize_enrollment_lesson_count();
DROP TRIGGER IF EXISTS update_lesson_progress_updated_at ON lesson_progress;
DROP TRIGGER IF EXISTS update_enrollments_updated_at ON course_enrollments;
DROP TABLE IF EXISTS lesson_progress;
DROP TABLE IF EXISTS course_enrollments;
DELETE FROM role_permissions WHERE permission_id IN (SELECT id FROM permissions WHERE code LIKE 'enrollments.%');
DELETE FROM permissions WHERE code LIKE 'enrollments.%';
