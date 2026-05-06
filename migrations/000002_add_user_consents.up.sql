-- migrations/000002_add_user_consents.up.sql
-- Registro de consentimientos otorgados por usuarios (cumplimiento LPDP Ley 29733)

CREATE TABLE IF NOT EXISTS auth.user_consents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,

    -- Tipo de consentimiento: terms_of_service, privacy_policy, marketing_communications
    consent_type VARCHAR(64) NOT NULL,

    -- Versión del documento aceptado (permite tracking histórico ante cambios de TyC)
    document_version VARCHAR(32) NOT NULL,

    -- Aceptación explícita del usuario (true) o rechazo explícito (false, p.ej. marketing)
    accepted BOOLEAN NOT NULL,

    -- Momento exacto de la acción (evidencia legal)
    accepted_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,

    -- Metadata para evidencia legal
    ip_address INET,
    user_agent TEXT,

    -- Revocación posterior (NULL = sigue vigente)
    withdrawn_at TIMESTAMPTZ,

    -- Audit
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,

    -- FK a users con CASCADE para cumplir ARCO (eliminar cuenta → eliminar consentimientos)
    CONSTRAINT fk_user_consents_user
        FOREIGN KEY (user_id)
        REFERENCES auth.users(id)
        ON DELETE CASCADE
);

-- Índices
CREATE INDEX idx_user_consents_user_id ON auth.user_consents(user_id);
CREATE INDEX idx_user_consents_type ON auth.user_consents(consent_type);
CREATE INDEX idx_user_consents_accepted_at ON auth.user_consents(accepted_at DESC);
CREATE INDEX idx_user_consents_active
    ON auth.user_consents(user_id, consent_type, accepted_at DESC)
    WHERE withdrawn_at IS NULL;

-- Comentarios
COMMENT ON TABLE auth.user_consents IS 'Registro de consentimientos otorgados por usuarios (LPDP Ley 29733)';
COMMENT ON COLUMN auth.user_consents.consent_type IS 'Tipo: terms_of_service, privacy_policy, marketing_communications';
COMMENT ON COLUMN auth.user_consents.document_version IS 'Versión del documento aceptado (ej. 2026-04-20)';
COMMENT ON COLUMN auth.user_consents.accepted IS 'true = aceptó, false = rechazó explícitamente';
COMMENT ON COLUMN auth.user_consents.ip_address IS 'IP del usuario al momento de aceptar (evidencia legal)';
COMMENT ON COLUMN auth.user_consents.user_agent IS 'User-Agent del navegador (evidencia legal)';
COMMENT ON COLUMN auth.user_consents.withdrawn_at IS 'Fecha de revocación (NULL = activo)';
