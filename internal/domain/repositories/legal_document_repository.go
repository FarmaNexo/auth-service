// internal/domain/repositories/legal_document_repository.go
package repositories

import (
	"context"

	"github.com/farmanexo/auth-service/internal/domain/entities"
	"github.com/google/uuid"
)

// LegalDocumentTypeRepository — catálogo de tipos.
type LegalDocumentTypeRepository interface {
	// FindAllActive lista los tipos activos ordenados por display_order.
	FindAllActive(ctx context.Context) ([]*entities.LegalDocumentType, error)

	// FindByCode busca un tipo por código (terms, privacy, marketing, ...).
	FindByCode(ctx context.Context, code string) (*entities.LegalDocumentType, error)

	// FindByID busca por id.
	FindByID(ctx context.Context, id uuid.UUID) (*entities.LegalDocumentType, error)
}

// LegalDocumentRepository — versiones de documentos legales.
type LegalDocumentRepository interface {
	// FindCurrentPublished retorna la única versión 'published' del tipo+locale.
	// Es la versión vigente legalmente y la que el sitio público debe servir.
	FindCurrentPublished(ctx context.Context, typeCode, locale string) (*entities.LegalDocument, error)

	// FindByTypeAndVersion retorna una versión específica (incluso archivada — para ARCO).
	FindByTypeAndVersion(ctx context.Context, typeCode, version, locale string) (*entities.LegalDocument, error)

	// FindByID retorna por id (incluye relación DocumentType cargada).
	FindByID(ctx context.Context, id uuid.UUID) (*entities.LegalDocument, error)

	// ListVersionsByType lista todas las versiones publicadas/archivadas de un tipo+locale,
	// ordenadas por published_at DESC. Excluye drafts e in_review.
	ListVersionsByType(ctx context.Context, typeCode, locale string) ([]*entities.LegalDocument, error)

	// Create persiste un nuevo documento (típicamente en status draft).
	Create(ctx context.Context, doc *entities.LegalDocument) error

	// Update actualiza un documento (solo permitido en draft / in_review).
	Update(ctx context.Context, doc *entities.LegalDocument) error

	// PublishAtomic ejecuta en transacción: archiva la versión 'published' actual del mismo
	// (type, locale) y marca la nueva como 'published'. Garantiza la invariante de "una sola versión vigente".
	PublishAtomic(ctx context.Context, newDocID uuid.UUID, approvedBy *uuid.UUID) error

	// Archive marca explícitamente una versión como archived (con archived_at).
	Archive(ctx context.Context, id uuid.UUID) error
}
