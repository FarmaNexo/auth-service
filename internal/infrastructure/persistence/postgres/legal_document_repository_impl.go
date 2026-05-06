// internal/infrastructure/persistence/postgres/legal_document_repository_impl.go
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/farmanexo/auth-service/internal/domain/entities"
	"github.com/farmanexo/auth-service/internal/domain/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ============================================================
// LegalDocumentTypeRepository implementation
// ============================================================

type LegalDocumentTypeRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewLegalDocumentTypeRepository(db *gorm.DB, logger *zap.Logger) repositories.LegalDocumentTypeRepository {
	return &LegalDocumentTypeRepositoryImpl{db: db, logger: logger}
}

func (r *LegalDocumentTypeRepositoryImpl) FindAllActive(ctx context.Context) ([]*entities.LegalDocumentType, error) {
	types := []*entities.LegalDocumentType{}
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("display_order ASC, name ASC").
		Find(&types).Error
	if err != nil {
		return nil, fmt.Errorf("error listing legal document types: %w", err)
	}
	return types, nil
}

func (r *LegalDocumentTypeRepositoryImpl) FindByCode(ctx context.Context, code string) (*entities.LegalDocumentType, error) {
	t := entities.LegalDocumentType{}
	err := r.db.WithContext(ctx).Where("code = ?", code).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding legal document type by code: %w", err)
	}
	return &t, nil
}

func (r *LegalDocumentTypeRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entities.LegalDocumentType, error) {
	t := entities.LegalDocumentType{}
	err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding legal document type by id: %w", err)
	}
	return &t, nil
}

var _ repositories.LegalDocumentTypeRepository = (*LegalDocumentTypeRepositoryImpl)(nil)

// ============================================================
// LegalDocumentRepository implementation
// ============================================================

type LegalDocumentRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewLegalDocumentRepository(db *gorm.DB, logger *zap.Logger) repositories.LegalDocumentRepository {
	return &LegalDocumentRepositoryImpl{db: db, logger: logger}
}

func (r *LegalDocumentRepositoryImpl) FindCurrentPublished(ctx context.Context, typeCode, locale string) (*entities.LegalDocument, error) {
	doc := entities.LegalDocument{}
	err := r.db.WithContext(ctx).
		Joins("JOIN auth.legal_document_types t ON t.id = legal_documents.document_type_id").
		Preload("DocumentType").
		Where("t.code = ? AND legal_documents.locale = ? AND legal_documents.status = ? AND legal_documents.deleted_at IS NULL",
			typeCode, locale, entities.LegalDocStatusPublished).
		First(&doc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding current published document: %w", err)
	}
	return &doc, nil
}

func (r *LegalDocumentRepositoryImpl) FindByTypeAndVersion(ctx context.Context, typeCode, version, locale string) (*entities.LegalDocument, error) {
	doc := entities.LegalDocument{}
	err := r.db.WithContext(ctx).
		Joins("JOIN auth.legal_document_types t ON t.id = legal_documents.document_type_id").
		Preload("DocumentType").
		Where("t.code = ? AND legal_documents.version = ? AND legal_documents.locale = ? AND legal_documents.deleted_at IS NULL",
			typeCode, version, locale).
		First(&doc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding document by type+version: %w", err)
	}
	return &doc, nil
}

func (r *LegalDocumentRepositoryImpl) FindByID(ctx context.Context, id uuid.UUID) (*entities.LegalDocument, error) {
	doc := entities.LegalDocument{}
	err := r.db.WithContext(ctx).
		Preload("DocumentType").
		First(&doc, "id = ? AND deleted_at IS NULL", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding document by id: %w", err)
	}
	return &doc, nil
}

func (r *LegalDocumentRepositoryImpl) ListVersionsByType(ctx context.Context, typeCode, locale string) ([]*entities.LegalDocument, error) {
	docs := []*entities.LegalDocument{}
	err := r.db.WithContext(ctx).
		Joins("JOIN auth.legal_document_types t ON t.id = legal_documents.document_type_id").
		Preload("DocumentType").
		Where("t.code = ? AND legal_documents.locale = ? AND legal_documents.deleted_at IS NULL AND legal_documents.status IN (?, ?)",
			typeCode, locale, entities.LegalDocStatusPublished, entities.LegalDocStatusArchived).
		Order("legal_documents.published_at DESC NULLS LAST, legal_documents.created_at DESC").
		Find(&docs).Error
	if err != nil {
		return nil, fmt.Errorf("error listing document versions: %w", err)
	}
	return docs, nil
}

func (r *LegalDocumentRepositoryImpl) Create(ctx context.Context, doc *entities.LegalDocument) error {
	if err := r.db.WithContext(ctx).Create(doc).Error; err != nil {
		r.logger.Error("Error creando legal document", zap.Error(err))
		return fmt.Errorf("error creating legal document: %w", err)
	}
	return nil
}

func (r *LegalDocumentRepositoryImpl) Update(ctx context.Context, doc *entities.LegalDocument) error {
	if err := r.db.WithContext(ctx).Save(doc).Error; err != nil {
		r.logger.Error("Error actualizando legal document", zap.Error(err))
		return fmt.Errorf("error updating legal document: %w", err)
	}
	return nil
}

// PublishAtomic — transacción: archiva la versión publicada actual del mismo
// (document_type_id, locale) y marca la nueva como publicada. Garantiza que nunca
// haya dos versiones 'published' simultáneas (también respaldado por unique index parcial).
func (r *LegalDocumentRepositoryImpl) PublishAtomic(ctx context.Context, newDocID uuid.UUID, approvedBy *uuid.UUID) error {
	now := time.Now()

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		newDoc := entities.LegalDocument{}
		if err := tx.First(&newDoc, "id = ? AND deleted_at IS NULL", newDocID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("document %s not found", newDocID)
			}
			return fmt.Errorf("error loading document to publish: %w", err)
		}

		if !newDoc.CanTransitionTo(entities.LegalDocStatusPublished) {
			return fmt.Errorf("document in status %q cannot transition to published", newDoc.Status)
		}

		// 1) Archivar la versión publicada actual del mismo (type, locale), si existe.
		archiveResult := tx.Model(&entities.LegalDocument{}).
			Where("document_type_id = ? AND locale = ? AND status = ? AND id <> ? AND deleted_at IS NULL",
				newDoc.DocumentTypeID, newDoc.Locale, entities.LegalDocStatusPublished, newDocID).
			Updates(map[string]interface{}{
				"status":      entities.LegalDocStatusArchived,
				"archived_at": now,
			})
		if archiveResult.Error != nil {
			return fmt.Errorf("error archiving previous published version: %w", archiveResult.Error)
		}

		// 2) Publicar la nueva versión.
		updates := map[string]interface{}{
			"status":       entities.LegalDocStatusPublished,
			"published_at": now,
		}
		if approvedBy != nil {
			updates["approved_by"] = *approvedBy
			updates["approved_at"] = now
		}
		if err := tx.Model(&entities.LegalDocument{}).
			Where("id = ?", newDocID).
			Updates(updates).Error; err != nil {
			return fmt.Errorf("error publishing new version: %w", err)
		}

		r.logger.Info("Documento publicado atómicamente",
			zap.String("new_doc_id", newDocID.String()),
			zap.Int64("archived_count", archiveResult.RowsAffected),
		)

		return nil
	})
}

func (r *LegalDocumentRepositoryImpl) Archive(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	res := r.db.WithContext(ctx).Model(&entities.LegalDocument{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Updates(map[string]interface{}{
			"status":      entities.LegalDocStatusArchived,
			"archived_at": now,
		})
	if res.Error != nil {
		return fmt.Errorf("error archiving document: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("document %s not found", id)
	}
	return nil
}

var _ repositories.LegalDocumentRepository = (*LegalDocumentRepositoryImpl)(nil)
