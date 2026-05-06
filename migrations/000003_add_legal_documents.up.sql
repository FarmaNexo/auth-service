-- migrations/000003_add_legal_documents.up.sql
-- Sistema de gestión de documentos legales versionados (Términos, Privacidad, Marketing, Cookies, etc.)
-- Cumplimiento LPDP (Ley 29733), evidencia de aceptación con integridad criptográfica (SHA-256).

-- ============================================================
-- 1. CATÁLOGO DE TIPOS DE DOCUMENTO
-- ============================================================
-- Tabla y NO enum: agregar nuevos tipos (cookies, data_processing, etc.) NO requiere migración.

CREATE TABLE IF NOT EXISTS auth.legal_document_types (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code            VARCHAR(50)  NOT NULL UNIQUE,
    name            VARCHAR(150) NOT NULL,
    description     TEXT,
    is_required     BOOLEAN      NOT NULL DEFAULT TRUE,
    display_order   INT          NOT NULL DEFAULT 0,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_legal_document_types_active
    ON auth.legal_document_types(is_active, display_order)
    WHERE is_active = TRUE;

COMMENT ON TABLE auth.legal_document_types IS 'Catálogo de tipos de documento legal (terms, privacy, marketing, cookies, ...)';
COMMENT ON COLUMN auth.legal_document_types.code IS 'Código estable usado por API y código (terms, privacy, marketing, cookies, data_processing)';
COMMENT ON COLUMN auth.legal_document_types.is_required IS 'Si es TRUE, el usuario debe aceptarlo para usar el servicio';
COMMENT ON COLUMN auth.legal_document_types.display_order IS 'Orden de presentación en UI (modal, registro)';

-- ============================================================
-- 2. ENUM DE STATUS DEL DOCUMENTO
-- ============================================================

DO $$ BEGIN
    CREATE TYPE auth.legal_document_status AS ENUM (
        'draft',        -- Editable, no visible al público
        'in_review',    -- En revisión legal, no visible al público
        'published',    -- Vigente, visible al público
        'archived',     -- Reemplazado por versión más nueva, accesible por ARCO
        'withdrawn'     -- Retirado por defecto legal/regulatorio (raro)
    );
EXCEPTION WHEN duplicate_object THEN null; END $$;

-- ============================================================
-- 3. DOCUMENTOS LEGALES VERSIONADOS
-- ============================================================

CREATE TABLE IF NOT EXISTS auth.legal_documents (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_type_id    UUID NOT NULL REFERENCES auth.legal_document_types(id) ON DELETE RESTRICT,

    version             VARCHAR(20)  NOT NULL,
    locale              VARCHAR(10)  NOT NULL DEFAULT 'es-PE',
    jurisdiction        VARCHAR(10)  NOT NULL DEFAULT 'PE',

    title               VARCHAR(255) NOT NULL,
    summary             TEXT,
    content_markdown    TEXT         NOT NULL,
    content_html        TEXT,
    content_hash        VARCHAR(64)  NOT NULL,

    changelog           TEXT,
    change_severity     VARCHAR(20),

    status              auth.legal_document_status NOT NULL DEFAULT 'draft',
    effective_date      DATE,
    published_at        TIMESTAMPTZ,
    archived_at         TIMESTAMPTZ,

    created_by          UUID,
    reviewed_by         UUID,
    approved_by         UUID,
    approved_at         TIMESTAMPTZ,

    metadata            JSONB        NOT NULL DEFAULT '{}'::jsonb,

    created_at          TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at          TIMESTAMPTZ,

    CONSTRAINT uq_legal_documents_type_version_locale
        UNIQUE (document_type_id, version, locale),

    CONSTRAINT chk_legal_documents_change_severity
        CHECK (change_severity IS NULL OR change_severity IN ('minor', 'major', 'critical'))
);

-- Índice principal: lookup de versión vigente por tipo + locale
CREATE INDEX idx_legal_documents_published
    ON auth.legal_documents(document_type_id, locale, published_at DESC)
    WHERE status = 'published' AND deleted_at IS NULL;

CREATE INDEX idx_legal_documents_status
    ON auth.legal_documents(status)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_legal_documents_hash
    ON auth.legal_documents(content_hash);

CREATE INDEX idx_legal_documents_type
    ON auth.legal_documents(document_type_id);

-- Solo UNA versión 'published' por (type, locale) en simultáneo
CREATE UNIQUE INDEX uq_legal_documents_one_published_per_type_locale
    ON auth.legal_documents(document_type_id, locale)
    WHERE status = 'published' AND deleted_at IS NULL;

COMMENT ON TABLE auth.legal_documents IS 'Versiones de documentos legales (markdown source, hash de integridad, workflow draft→published→archived)';
COMMENT ON COLUMN auth.legal_documents.content_hash IS 'SHA-256 del content_markdown — evidencia de integridad para auditoría legal';
COMMENT ON COLUMN auth.legal_documents.summary IS 'Resumen ejecutivo mostrado en modal de re-aceptación';
COMMENT ON COLUMN auth.legal_documents.changelog IS 'Qué cambió respecto a la versión anterior (visible al user en re-aceptación)';
COMMENT ON COLUMN auth.legal_documents.change_severity IS 'minor (typo, redacción) | major (cambio de cláusula) | critical (fuerza re-aceptación a todos los users)';
COMMENT ON COLUMN auth.legal_documents.effective_date IS 'Fecha legal de entrada en vigor (puede ser posterior a published_at: aviso previo de 15 días)';
COMMENT ON COLUMN auth.legal_documents.metadata IS 'Extensible: ID de aprobación externa, link a PDF firmado en DocuSign, etc.';

-- ============================================================
-- 4. ADAPTAR user_consents PARA APUNTAR AL DOCUMENTO EXACTO
-- ============================================================

ALTER TABLE auth.user_consents
    ADD COLUMN IF NOT EXISTS document_id UUID REFERENCES auth.legal_documents(id) ON DELETE RESTRICT,
    ADD COLUMN IF NOT EXISTS content_hash_at_acceptance VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_user_consents_document
    ON auth.user_consents(document_id);

COMMENT ON COLUMN auth.user_consents.document_id IS 'Referencia al documento legal exacto que el usuario aceptó (FK estricta para integridad)';
COMMENT ON COLUMN auth.user_consents.content_hash_at_acceptance IS 'SHA-256 del documento al momento de aceptación (defensa en profundidad: si alguien edita el documento, el hash deja de coincidir y queda evidencia)';

-- ============================================================
-- 5. TRIGGER: actualizar updated_at automáticamente
-- ============================================================

CREATE OR REPLACE FUNCTION auth.legal_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_legal_document_types_updated_at
    BEFORE UPDATE ON auth.legal_document_types
    FOR EACH ROW EXECUTE FUNCTION auth.legal_set_updated_at();

CREATE TRIGGER trg_legal_documents_updated_at
    BEFORE UPDATE ON auth.legal_documents
    FOR EACH ROW EXECUTE FUNCTION auth.legal_set_updated_at();
