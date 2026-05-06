-- migrations/000003_add_legal_documents.down.sql

DROP TRIGGER IF EXISTS trg_legal_documents_updated_at ON auth.legal_documents;
DROP TRIGGER IF EXISTS trg_legal_document_types_updated_at ON auth.legal_document_types;
DROP FUNCTION IF EXISTS auth.legal_set_updated_at();

ALTER TABLE auth.user_consents
    DROP COLUMN IF EXISTS document_id,
    DROP COLUMN IF EXISTS content_hash_at_acceptance;

DROP TABLE IF EXISTS auth.legal_documents CASCADE;
DROP TABLE IF EXISTS auth.legal_document_types CASCADE;
DROP TYPE IF EXISTS auth.legal_document_status;
