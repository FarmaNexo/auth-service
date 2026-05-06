-- migrations/000004_seed_legal_documents.down.sql

UPDATE auth.user_consents
SET document_id = NULL,
    content_hash_at_acceptance = NULL;

DELETE FROM auth.legal_documents
WHERE document_type_id IN (
    SELECT id FROM auth.legal_document_types WHERE code IN ('terms', 'privacy', 'marketing')
)
AND version = '1.0.0' AND locale = 'es-PE';

DELETE FROM auth.legal_document_types WHERE code IN ('terms', 'privacy', 'marketing');
