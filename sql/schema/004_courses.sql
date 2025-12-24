-- +goose Up
-- ============================================================================
-- ORBIT BACKEND - Courses Schema (MVP)
-- Course Builder with Modules and Lessons
-- ============================================================================

-- ============================================================================
-- CONTENT: COURSES
-- ============================================================================

CREATE TABLE courses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Identificacao
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    description TEXT,

    -- Media
    thumbnail_url VARCHAR(512),

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    published_at TIMESTAMPTZ,

    -- Metricas (denormalizadas para performance)
    module_count INT NOT NULL DEFAULT 0,
    lesson_count INT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(tenant_id, slug)
);

CREATE INDEX idx_courses_tenant ON courses(tenant_id);
CREATE INDEX idx_courses_author ON courses(author_id);
CREATE INDEX idx_courses_status ON courses(status);
CREATE INDEX idx_courses_published_at ON courses(published_at DESC);

-- ============================================================================
-- CONTENT: MODULES (Sections within a Course)
-- ============================================================================

CREATE TABLE modules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    course_id UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,

    -- Identificacao
    title VARCHAR(500) NOT NULL,
    description TEXT,

    -- Ordenacao (para drag-and-drop)
    position INT NOT NULL DEFAULT 0,

    -- Metricas
    lesson_count INT NOT NULL DEFAULT 0,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_modules_tenant ON modules(tenant_id);
CREATE INDEX idx_modules_course ON modules(course_id);
CREATE INDEX idx_modules_position ON modules(course_id, position);

-- ============================================================================
-- CONTENT: LESSONS (Content within a Module)
-- ============================================================================

CREATE TABLE lessons (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    module_id UUID NOT NULL REFERENCES modules(id) ON DELETE CASCADE,

    -- Identificacao
    title VARCHAR(500) NOT NULL,
    description TEXT,

    -- Conteudo
    content TEXT,                              -- Rich text content (markdown/html)
    content_format VARCHAR(20) NOT NULL DEFAULT 'markdown' CHECK (content_format IN ('markdown', 'json', 'html')),

    -- Video associado (optional)
    video_id UUID REFERENCES videos(id) ON DELETE SET NULL,

    -- Ordenacao (para drag-and-drop)
    position INT NOT NULL DEFAULT 0,

    -- Duracao estimada (em minutos)
    duration_minutes INT DEFAULT 0,

    -- Flags
    is_free_preview BOOLEAN NOT NULL DEFAULT FALSE,  -- Permite preview sem enrollment

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_lessons_tenant ON lessons(tenant_id);
CREATE INDEX idx_lessons_module ON lessons(module_id);
CREATE INDEX idx_lessons_video ON lessons(video_id);
CREATE INDEX idx_lessons_position ON lessons(module_id, position);

-- ============================================================================
-- TRIGGERS: Auto-update updated_at
-- ============================================================================

CREATE TRIGGER update_courses_updated_at BEFORE UPDATE ON courses FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_modules_updated_at BEFORE UPDATE ON modules FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_lessons_updated_at BEFORE UPDATE ON lessons FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- TRIGGERS: Auto-update counters
-- ============================================================================

-- Module count on course
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_course_module_count()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        UPDATE courses SET module_count = module_count + 1 WHERE id = NEW.course_id;
    ELSIF TG_OP = 'DELETE' THEN
        UPDATE courses SET module_count = module_count - 1 WHERE id = OLD.course_id;
    END IF;
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER update_course_module_count_trigger
    AFTER INSERT OR DELETE ON modules
    FOR EACH ROW EXECUTE FUNCTION update_course_module_count();

-- Lesson count on module and course
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_lesson_counts()
RETURNS TRIGGER AS $$
DECLARE
    v_course_id UUID;
BEGIN
    IF TG_OP = 'INSERT' THEN
        SELECT course_id INTO v_course_id FROM modules WHERE id = NEW.module_id;
        UPDATE modules SET lesson_count = lesson_count + 1 WHERE id = NEW.module_id;
        UPDATE courses SET lesson_count = lesson_count + 1 WHERE id = v_course_id;
    ELSIF TG_OP = 'DELETE' THEN
        SELECT course_id INTO v_course_id FROM modules WHERE id = OLD.module_id;
        UPDATE modules SET lesson_count = lesson_count - 1 WHERE id = OLD.module_id;
        UPDATE courses SET lesson_count = lesson_count - 1 WHERE id = v_course_id;
    END IF;
    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER update_lesson_counts_trigger
    AFTER INSERT OR DELETE ON lessons
    FOR EACH ROW EXECUTE FUNCTION update_lesson_counts();

-- ============================================================================
-- SEED: Course-related permissions
-- ============================================================================

INSERT INTO permissions (code, name, description, category) VALUES
    -- Courses
    ('courses.view', 'Ver cursos', 'Visualizar cursos publicados', 'courses'),
    ('courses.create', 'Criar cursos', 'Criar novos cursos', 'courses'),
    ('courses.edit', 'Editar cursos', 'Editar qualquer curso', 'courses'),
    ('courses.edit_own', 'Editar proprios cursos', 'Editar apenas cursos proprios', 'courses'),
    ('courses.delete', 'Deletar cursos', 'Deletar qualquer curso', 'courses'),
    ('courses.delete_own', 'Deletar proprios cursos', 'Deletar apenas cursos proprios', 'courses'),
    ('courses.publish', 'Publicar cursos', 'Publicar cursos', 'courses');

-- ============================================================================
-- UPDATE: Add course permissions to existing roles
-- ============================================================================

-- Add ALL course permissions to Owner role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.slug = 'owner' AND r.is_system = TRUE
AND p.code IN ('courses.view', 'courses.create', 'courses.edit', 'courses.edit_own',
               'courses.delete', 'courses.delete_own', 'courses.publish')
ON CONFLICT DO NOTHING;

-- Add most course permissions to Admin role (excluding courses.delete)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.slug = 'admin' AND r.is_system = TRUE
AND p.code IN ('courses.view', 'courses.create', 'courses.edit', 'courses.edit_own',
               'courses.delete_own', 'courses.publish')
ON CONFLICT DO NOTHING;

-- Add basic course permissions to Member role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
CROSS JOIN permissions p
WHERE r.slug = 'member' AND r.is_system = TRUE
AND p.code IN ('courses.view')
ON CONFLICT DO NOTHING;
