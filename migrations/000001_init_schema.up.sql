-- migrations/000001_init_schema.up.sql
-- Migración inicial del Auth Service

-- Crear esquema dedicado para auth
CREATE SCHEMA IF NOT EXISTS auth;

-- Habilitar extensión UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Tabla users en esquema auth
CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    phone VARCHAR(50),
    
    -- Status flags
    is_active BOOLEAN DEFAULT true NOT NULL,
    is_verified BOOLEAN DEFAULT false NOT NULL,
    email_verified_at TIMESTAMPTZ,
    
    -- Role and permissions
    role VARCHAR(50) DEFAULT 'user' NOT NULL,
    
    -- Login tracking
    last_login_at TIMESTAMPTZ,
    login_count INTEGER DEFAULT 0 NOT NULL,
    
    -- Audit fields
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted_at TIMESTAMPTZ
);

-- Índices para users
CREATE UNIQUE INDEX idx_users_email ON auth.users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_deleted_at ON auth.users(deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX idx_users_role ON auth.users(role);
CREATE INDEX idx_users_is_active ON auth.users(is_active);

-- Tabla refresh_tokens en esquema auth
CREATE TABLE IF NOT EXISTS auth.refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    
    -- Revocation
    is_revoked BOOLEAN DEFAULT false NOT NULL,
    revoked_at TIMESTAMPTZ,
    
    -- Metadata
    ip_address VARCHAR(50),
    user_agent TEXT,
    
    -- Audit
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    
    -- Foreign key
    CONSTRAINT fk_refresh_tokens_user 
        FOREIGN KEY (user_id) 
        REFERENCES auth.users(id) 
        ON DELETE CASCADE
);

-- Índices para refresh_tokens
CREATE INDEX idx_refresh_tokens_user_id ON auth.refresh_tokens(user_id);
CREATE UNIQUE INDEX idx_refresh_tokens_token_hash ON auth.refresh_tokens(token_hash) WHERE is_revoked = false;
CREATE INDEX idx_refresh_tokens_expires_at ON auth.refresh_tokens(expires_at);
CREATE INDEX idx_refresh_tokens_is_revoked ON auth.refresh_tokens(is_revoked) WHERE is_revoked = false;

-- Comentarios en tablas
COMMENT ON TABLE auth.users IS 'Tabla principal de usuarios del sistema';
COMMENT ON TABLE auth.refresh_tokens IS 'Tokens de refresco para autenticación JWT';

-- Comentarios en columnas importantes
COMMENT ON COLUMN auth.users.email IS 'Email único del usuario (usado para login)';
COMMENT ON COLUMN auth.users.password_hash IS 'Hash bcrypt del password';
COMMENT ON COLUMN auth.users.role IS 'Rol del usuario: user, pharmacy_owner, admin';
COMMENT ON COLUMN auth.refresh_tokens.token_hash IS 'SHA-256 hash del refresh token';