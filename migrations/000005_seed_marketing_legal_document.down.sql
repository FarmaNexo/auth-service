-- migrations/000005_seed_marketing_legal_document.down.sql

UPDATE auth.user_consents
SET document_id = NULL, content_hash_at_acceptance = NULL
WHERE consent_type = 'marketing_communications'
  AND document_id IN (
    SELECT ld.id FROM auth.legal_documents ld
    JOIN auth.legal_document_types lt ON lt.id = ld.document_type_id
    WHERE lt.code = 'marketing' AND ld.version = '1.0.0' AND ld.locale = 'es-PE'
  );

DELETE FROM auth.legal_documents
WHERE document_type_id IN (SELECT id FROM auth.legal_document_types WHERE code = 'marketing')
  AND version = '1.0.0' AND locale = 'es-PE';
