-- +goose Up
-- ============================================================================
-- ORBIT BACKEND - Initial Schema
-- Multi-tenant community platform
-- ============================================================================

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================================
-- CORE: TENANTS (Comunidades/Criadores)
-- ============================================================================

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Identificação
    slug VARCHAR(63) NOT NULL UNIQUE,          -- URL-friendly identifier (ex: "minha-comunidade")
    name VARCHAR(255) NOT NULL,
    description TEXT,
    logo_url VARCHAR(512),

    -- Configurações
    settings JSONB DEFAULT '{}',               -- Configurações flexíveis do tenant

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deleted')),

    -- Billing (future-proofing para Stripe)
    billing_status VARCHAR(20) NOT NULL DEFAULT 'free' CHECK (billing_status IN ('free', 'active', 'past_due', 'canceled')),
    stripe_customer_id VARCHAR(255),           -- Stripe Customer ID (cus_xxx)
    stripe_subscription_id VARCHAR(255),       -- Stripe Subscription ID (sub_xxx)
    plan_id VARCHAR(50),                       -- Identificador do plano (ex: "pro", "enterprise")

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_status ON tenants(status);
CREATE INDEX idx_tenants_billing_status ON tenants(billing_status);
CREATE INDEX idx_tenants_stripe_customer ON tenants(stripe_customer_id) WHERE stripe_customer_id IS NOT NULL;

-- ============================================================================
-- CORE: USERS (Autenticação global)
-- ============================================================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Autenticação
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,

    -- Profile global
    name VARCHAR(255) NOT NULL,
    avatar_url VARCHAR(512),

    -- Status
    email_verified_at TIMESTAMPTZ,
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'deleted')),

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_status ON users(status);

-- ============================================================================
-- RBAC: ROLES & PERMISSIONS
-- ============================================================================

-- Permissões disponíveis no sistema (definidas pela plataforma)
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Identificação
    code VARCHAR(50) NOT NULL UNIQUE,          -- Ex: "posts.create", "videos.upload", "members.manage"
    name VARCHAR(100) NOT NULL,
    description TEXT,

    -- Agrupamento para UI
    category VARCHAR(50) NOT NULL,             -- Ex: "content", "members", "settings"

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Roles customizáveis por tenant
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Identificação
    slug VARCHAR(50) NOT NULL,                 -- Ex: "admin", "moderator", "member"
    name VARCHAR(100) NOT NULL,
    description TEXT,

    -- Hierarquia (quanto maior, mais poder)
    priority INT NOT NULL DEFAULT 0,

    -- Flags especiais
    is_default BOOLEAN NOT NULL DEFAULT FALSE, -- Role atribuída automaticamente a novos membros
    is_system BOOLEAN NOT NULL DEFAULT FALSE,  -- Roles do sistema (owner) não podem ser deletadas

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(tenant_id, slug)
);

CREATE INDEX idx_roles_tenant ON roles(tenant_id);

-- Junction: Role <-> Permissions
CREATE TABLE role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (role_id, permission_id)
);

-- ============================================================================
-- MEMBERSHIP: Relação User <-> Tenant
-- ============================================================================

CREATE TABLE tenant_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,

    -- Profile específico do tenant (override do profile global)
    display_name VARCHAR(255),
    bio TEXT,

    -- Status do membership
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'suspended', 'banned')),

    -- Timestamps
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(tenant_id, user_id)
);

CREATE INDEX idx_tenant_members_tenant ON tenant_members(tenant_id);
CREATE INDEX idx_tenant_members_user ON tenant_members(user_id);
CREATE INDEX idx_tenant_members_role ON tenant_members(role_id);

-- ============================================================================
-- CONTENT: CATEGORIES/MODULES
-- ============================================================================

CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Identificação
    slug VARCHAR(63) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    icon VARCHAR(50),                          -- Nome do ícone (ex: "book", "video")

    -- Ordenação
    position INT NOT NULL DEFAULT 0,

    -- Visibilidade
    is_visible BOOLEAN NOT NULL DEFAULT TRUE,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(tenant_id, slug)
);

CREATE INDEX idx_categories_tenant ON categories(tenant_id);

-- ============================================================================
-- CONTENT: POSTS
-- ============================================================================

CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Conteúdo
    title VARCHAR(500) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    content TEXT,                              -- Pode ser Markdown ou JSON (editor rico)
    content_format VARCHAR(20) NOT NULL DEFAULT 'markdown' CHECK (content_format IN ('markdown', 'json', 'html')),

    -- SEO/Preview
    excerpt VARCHAR(500),
    cover_image_url VARCHAR(512),

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'archived')),
    published_at TIMESTAMPTZ,

    -- Métricas (denormalizadas para performance)
    view_count INT NOT NULL DEFAULT 0,
    like_count INT NOT NULL DEFAULT 0,
    comment_count INT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(tenant_id, slug)
);

CREATE INDEX idx_posts_tenant ON posts(tenant_id);
CREATE INDEX idx_posts_category ON posts(category_id);
CREATE INDEX idx_posts_author ON posts(author_id);
CREATE INDEX idx_posts_status ON posts(status);
CREATE INDEX idx_posts_published_at ON posts(published_at DESC);

-- ============================================================================
-- CONTENT: VIDEOS (Metadados - processamento externo)
-- ============================================================================

CREATE TABLE videos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    uploader_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Identificação
    title VARCHAR(500) NOT NULL,
    description TEXT,

    -- Storage externo (S3/R2/Cloudflare Stream)
    external_id VARCHAR(255),                  -- ID no provedor externo
    provider VARCHAR(50) NOT NULL DEFAULT 'cloudflare' CHECK (provider IN ('cloudflare', 's3', 'mux', 'bunny')),

    -- URLs (preenchidas após processamento via webhook)
    original_url VARCHAR(512),                 -- URL do arquivo original
    playback_url VARCHAR(512),                 -- URL do HLS/DASH para playback
    thumbnail_url VARCHAR(512),

    -- Metadados do vídeo
    duration_seconds INT,
    file_size_bytes BIGINT,
    resolution VARCHAR(20),                    -- Ex: "1080p", "4k"

    -- Status do processamento
    status VARCHAR(30) NOT NULL DEFAULT 'pending' CHECK (status IN (
        'pending',          -- Aguardando upload
        'uploading',        -- Upload em andamento
        'processing',       -- Processando no provedor
        'ready',            -- Pronto para reprodução
        'failed'            -- Falha no processamento
    )),
    error_message TEXT,                        -- Mensagem de erro (se status = failed)

    -- Associação opcional com post
    post_id UUID REFERENCES posts(id) ON DELETE SET NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_videos_tenant ON videos(tenant_id);
CREATE INDEX idx_videos_uploader ON videos(uploader_id);
CREATE INDEX idx_videos_status ON videos(status);
CREATE INDEX idx_videos_external ON videos(provider, external_id);
CREATE INDEX idx_videos_post ON videos(post_id);

-- ============================================================================
-- SOCIAL: COMMENTS (Threads com auto-relacionamento)
-- ============================================================================

CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Thread (auto-relacionamento para respostas)
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,

    -- Conteúdo
    content TEXT NOT NULL,

    -- Métricas (denormalizadas para performance)
    like_count INT NOT NULL DEFAULT 0,
    reply_count INT NOT NULL DEFAULT 0,

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'visible' CHECK (status IN ('visible', 'hidden', 'deleted')),

    -- Profundidade da thread (para limitar aninhamento se necessário)
    depth INT NOT NULL DEFAULT 0,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comments_tenant ON comments(tenant_id);
CREATE INDEX idx_comments_post ON comments(post_id);
CREATE INDEX idx_comments_author ON comments(author_id);
CREATE INDEX idx_comments_parent ON comments(parent_id);
CREATE INDEX idx_comments_status ON comments(status);
-- Index para buscar comentários raiz de um post ordenados por data
CREATE INDEX idx_comments_post_root ON comments(post_id, created_at DESC) WHERE parent_id IS NULL;

-- ============================================================================
-- AUDIT: WEBHOOK EVENTS LOG
-- ============================================================================

CREATE TABLE webhook_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Origem
    provider VARCHAR(50) NOT NULL,             -- Ex: "cloudflare", "stripe", "pagar_me"
    event_type VARCHAR(100) NOT NULL,          -- Ex: "video.ready", "subscription.created"

    -- Payload
    payload JSONB NOT NULL,

    -- Processamento
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processed', 'failed', 'ignored')),
    processed_at TIMESTAMPTZ,
    error_message TEXT,

    -- Idempotência
    idempotency_key VARCHAR(255),              -- Para evitar processamento duplicado

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(provider, idempotency_key)
);

CREATE INDEX idx_webhook_events_provider ON webhook_events(provider);
CREATE INDEX idx_webhook_events_status ON webhook_events(status);
CREATE INDEX idx_webhook_events_created ON webhook_events(created_at DESC);

-- ============================================================================
-- SEED: DEFAULT PERMISSIONS
-- ============================================================================

INSERT INTO permissions (code, name, description, category) VALUES
    -- Content
    ('posts.view', 'Ver posts', 'Visualizar posts publicados', 'content'),
    ('posts.create', 'Criar posts', 'Criar novos posts', 'content'),
    ('posts.edit', 'Editar posts', 'Editar qualquer post', 'content'),
    ('posts.edit_own', 'Editar próprios posts', 'Editar apenas posts próprios', 'content'),
    ('posts.delete', 'Deletar posts', 'Deletar qualquer post', 'content'),
    ('posts.delete_own', 'Deletar próprios posts', 'Deletar apenas posts próprios', 'content'),

    -- Videos
    ('videos.view', 'Ver vídeos', 'Assistir vídeos', 'content'),
    ('videos.upload', 'Upload de vídeos', 'Fazer upload de vídeos', 'content'),
    ('videos.delete', 'Deletar vídeos', 'Deletar qualquer vídeo', 'content'),
    ('videos.delete_own', 'Deletar próprios vídeos', 'Deletar apenas vídeos próprios', 'content'),

    -- Comments
    ('comments.view', 'Ver comentários', 'Visualizar comentários', 'content'),
    ('comments.create', 'Comentar', 'Criar comentários em posts', 'content'),
    ('comments.edit_own', 'Editar próprios comentários', 'Editar apenas comentários próprios', 'content'),
    ('comments.delete', 'Deletar comentários', 'Deletar qualquer comentário', 'content'),
    ('comments.delete_own', 'Deletar próprios comentários', 'Deletar apenas comentários próprios', 'content'),

    -- Members
    ('members.view', 'Ver membros', 'Visualizar lista de membros', 'members'),
    ('members.invite', 'Convidar membros', 'Convidar novos membros', 'members'),
    ('members.manage', 'Gerenciar membros', 'Alterar roles e banir membros', 'members'),
    ('members.remove', 'Remover membros', 'Remover membros da comunidade', 'members'),

    -- Moderation
    ('moderation.warn', 'Avisar membros', 'Enviar avisos a membros', 'moderation'),
    ('moderation.mute', 'Silenciar membros', 'Silenciar membros temporariamente', 'moderation'),
    ('moderation.ban', 'Banir membros', 'Banir membros permanentemente', 'moderation'),

    -- Settings
    ('settings.view', 'Ver configurações', 'Visualizar configurações da comunidade', 'settings'),
    ('settings.edit', 'Editar configurações', 'Alterar configurações da comunidade', 'settings'),
    ('roles.manage', 'Gerenciar roles', 'Criar, editar e deletar roles', 'settings'),
    ('categories.manage', 'Gerenciar categorias', 'Criar, editar e deletar categorias', 'settings');

-- ============================================================================
-- FUNCTIONS: Auto-update updated_at
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply triggers
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_tenant_members_updated_at BEFORE UPDATE ON tenant_members FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_categories_updated_at BEFORE UPDATE ON categories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_posts_updated_at BEFORE UPDATE ON posts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_videos_updated_at BEFORE UPDATE ON videos FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_comments_updated_at BEFORE UPDATE ON comments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================================================
-- FUNCTIONS: Auto-create default roles on tenant creation
-- ============================================================================

CREATE OR REPLACE FUNCTION create_default_roles()
RETURNS TRIGGER AS $$
DECLARE
    owner_role_id UUID;
    admin_role_id UUID;
    member_role_id UUID;
    perm RECORD;
BEGIN
    -- Create Owner role (all permissions, system role)
    INSERT INTO roles (tenant_id, slug, name, description, priority, is_system)
    VALUES (NEW.id, 'owner', 'Owner', 'Dono da comunidade com acesso total', 1000, TRUE)
    RETURNING id INTO owner_role_id;

    -- Create Admin role
    INSERT INTO roles (tenant_id, slug, name, description, priority, is_system)
    VALUES (NEW.id, 'admin', 'Admin', 'Administrador com amplos poderes', 500, TRUE)
    RETURNING id INTO admin_role_id;

    -- Create Member role (default)
    INSERT INTO roles (tenant_id, slug, name, description, priority, is_default, is_system)
    VALUES (NEW.id, 'member', 'Membro', 'Membro padrão da comunidade', 100, TRUE, TRUE)
    RETURNING id INTO member_role_id;

    -- Assign ALL permissions to Owner
    FOR perm IN SELECT id FROM permissions LOOP
        INSERT INTO role_permissions (role_id, permission_id) VALUES (owner_role_id, perm.id);
    END LOOP;

    -- Assign management permissions to Admin (excluding roles.manage)
    FOR perm IN SELECT id FROM permissions WHERE code NOT IN ('roles.manage', 'settings.edit') LOOP
        INSERT INTO role_permissions (role_id, permission_id) VALUES (admin_role_id, perm.id);
    END LOOP;

    -- Assign basic permissions to Member
    FOR perm IN SELECT id FROM permissions WHERE code IN (
        'posts.view', 'posts.create', 'posts.edit_own', 'posts.delete_own',
        'videos.view', 'videos.upload', 'videos.delete_own',
        'comments.view', 'comments.create', 'comments.edit_own', 'comments.delete_own',
        'members.view'
    ) LOOP
        INSERT INTO role_permissions (role_id, permission_id) VALUES (member_role_id, perm.id);
    END LOOP;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER create_tenant_default_roles
    AFTER INSERT ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION create_default_roles();

-- ============================================================================
-- FUNCTIONS: Auto-calculate comment depth and update counters
-- ============================================================================

CREATE OR REPLACE FUNCTION set_comment_depth()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.parent_id IS NULL THEN
        NEW.depth = 0;
    ELSE
        SELECT depth + 1 INTO NEW.depth FROM comments WHERE id = NEW.parent_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_comment_depth_trigger
    BEFORE INSERT ON comments
    FOR EACH ROW
    EXECUTE FUNCTION set_comment_depth();

-- Atualiza reply_count do comentário pai e comment_count do post
CREATE OR REPLACE FUNCTION update_comment_counters()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        -- Incrementa comment_count do post
        UPDATE posts SET comment_count = comment_count + 1 WHERE id = NEW.post_id;

        -- Incrementa reply_count do comentário pai (se existir)
        IF NEW.parent_id IS NOT NULL THEN
            UPDATE comments SET reply_count = reply_count + 1 WHERE id = NEW.parent_id;
        END IF;

    ELSIF TG_OP = 'DELETE' THEN
        -- Decrementa comment_count do post
        UPDATE posts SET comment_count = comment_count - 1 WHERE id = OLD.post_id;

        -- Decrementa reply_count do comentário pai (se existir)
        IF OLD.parent_id IS NOT NULL THEN
            UPDATE comments SET reply_count = reply_count - 1 WHERE id = OLD.parent_id;
        END IF;
    END IF;

    RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_comment_counters_trigger
    AFTER INSERT OR DELETE ON comments
    FOR EACH ROW
    EXECUTE FUNCTION update_comment_counters();
