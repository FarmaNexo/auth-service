-- migrations/000005_seed_marketing_legal_document.up.sql
-- Seed v1.0.0 de la Política de Comunicaciones de Marketing.
-- Necesario porque RegisterUserHandler busca la versión vigente al crear
-- el consent (aceptado o rechazado) y document_id es FK a legal_documents.

INSERT INTO auth.legal_documents (
    document_type_id, version, locale, jurisdiction,
    title, summary, content_markdown, content_hash,
    changelog, change_severity, status,
    effective_date, published_at, approved_at, metadata
)
SELECT
    t.id, '1.0.0', 'es-PE', 'PE',
    'Política de Comunicaciones de Marketing',
    'Documento que describe el alcance del consentimiento opcional para recibir comunicaciones de marketing. Su otorgamiento o rechazo no afecta el acceso a los servicios principales.',
    $LEGALDOC$
# Política de Comunicaciones de Marketing

**FARMANEXO S.A.C.**

*Versión 1.0 — Vigente desde Enero 2025*

---

## 1. Alcance

Este documento describe el alcance del consentimiento opcional para recibir **comunicaciones de marketing** por parte de Farmanexo S.A.C.

El consentimiento de marketing es **completamente opcional** y su otorgamiento o rechazo no afecta el acceso del usuario a los servicios principales de la Plataforma.

## 2. Tipos de Comunicaciones

Si el usuario otorga este consentimiento, Farmanexo podrá enviarle:

- Novedades sobre productos farmacéuticos disponibles en la Plataforma.
- Promociones, descuentos y ofertas exclusivas de farmacias asociadas.
- Información sobre nuevas funcionalidades de la Plataforma.
- Encuestas de satisfacción y estudios de mercado relacionados con el sector farmacéutico.
- Boletines informativos sobre salud, farmacovigilancia y autocuidado.

## 3. Canales

Las comunicaciones se enviarán a través de los siguientes canales:

- Correo electrónico (al email registrado en la cuenta).
- Notificaciones push (si la app móvil está instalada).
- SMS (solo cuando el usuario lo haya autorizado expresamente).

## 4. Frecuencia

Farmanexo procurará una frecuencia razonable de comunicaciones, evitando saturar al usuario. La frecuencia estimada es de **no más de cuatro (4) comunicaciones por mes**.

## 5. Revocación del Consentimiento

El usuario puede **revocar este consentimiento en cualquier momento**, sin necesidad de justificación, mediante:

- El enlace de "darse de baja" (unsubscribe) presente en cada correo electrónico.
- La sección "Privacidad" de su perfil en la Plataforma.
- Solicitud directa al correo: `legal@farmanexo.com.pe`

La revocación se hará efectiva en un plazo máximo de **cinco (5) días hábiles**.

## 6. Base Legal

Este consentimiento se otorga conforme a:

- Ley N° 29733 — Ley de Protección de Datos Personales.
- Ley N° 28493 — Ley que Regula el Uso del Correo Electrónico Comercial No Solicitado.
- Reglamento de la Ley de Protección de Datos Personales (D.S. N° 003-2013-JUS).

## 7. Datos de Contacto

- Correo: legal@farmanexo.com.pe
- Atención: lunes a viernes, 9:00 a.m. — 6:00 p.m. (hora de Lima)

---

*Farmanexo S.A.C. — Lima, Perú — farmanexo.com.pe*
$LEGALDOC$,
    '320752226542275e792a235b86f55548c57262b1215d0cdfa25b3f4bba8e2ad9',
    'Versión inicial publicada — Enero 2025.',
    'major', 'published',
    '2025-01-01', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP,
    '{"published_by": "system_seed", "source": "migration_000005"}'::jsonb
FROM auth.legal_document_types t
WHERE t.code = 'marketing'
ON CONFLICT (document_type_id, version, locale) DO NOTHING;

-- Backfill consents marketing que aún no tienen document_id
UPDATE auth.user_consents uc
SET document_id = ld.id, content_hash_at_acceptance = ld.content_hash
FROM auth.legal_documents ld
JOIN auth.legal_document_types lt ON lt.id = ld.document_type_id
WHERE uc.consent_type = 'marketing_communications'
  AND uc.document_id IS NULL
  AND lt.code = 'marketing'
  AND ld.locale = 'es-PE'
  AND ld.version = '1.0.0';
