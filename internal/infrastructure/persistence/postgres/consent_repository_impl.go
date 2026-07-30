// internal/infrastructure/persistence/postgres/consent_repository_impl.go
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/farmanexo/auth-service/internal/domain/entities"
	"github.com/farmanexo/auth-service/internal/domain/repositories"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ConsentRepositoryImpl implementa ConsentRepository usando PostgreSQL.
type ConsentRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewConsentRepository crea una nueva instancia del repositorio de consentimientos.
func NewConsentRepository(db *gorm.DB, logger *zap.Logger) repositories.ConsentRepository {
	return &ConsentRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

// Create persiste un solo consentimiento.
func (r *ConsentRepositoryImpl) Create(ctx context.Context, consent *entities.UserConsent) error {
	r.logger.Debug("Creando consentimiento",
		zap.String("user_id", consent.UserID.String()),
		zap.String("consent_type", consent.ConsentType),
		zap.Bool("accepted", consent.Accepted),
	)

	result := dbConn(ctx, r.db).Create(consent)
	if result.Error != nil {
		r.logger.Error("Error creando consentimiento",
			zap.Error(result.Error),
			zap.String("user_id", consent.UserID.String()),
			zap.String("consent_type", consent.ConsentType),
		)
		return fmt.Errorf("error creating consent: %w", result.Error)
	}

	return nil
}

// CreateBatch persiste múltiples consentimientos en una transacción.
func (r *ConsentRepositoryImpl) CreateBatch(ctx context.Context, consents []*entities.UserConsent) error {
	if len(consents) == 0 {
		return nil
	}

	r.logger.Debug("Creando consentimientos en batch",
		zap.Int("count", len(consents)),
	)

	result := dbConn(ctx, r.db).Create(&consents)
	if result.Error != nil {
		r.logger.Error("Error creando consentimientos en batch",
			zap.Error(result.Error),
			zap.Int("count", len(consents)),
		)
		return fmt.Errorf("error creating consents batch: %w", result.Error)
	}

	return nil
}

// FindAllByUserID devuelve todos los consentimientos de un usuario ordenados por fecha descendente.
func (r *ConsentRepositoryImpl) FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.UserConsent, error) {
	var consents []*entities.UserConsent

	result := dbConn(ctx, r.db).
		Where("user_id = ?", userID).
		Order("accepted_at DESC").
		Find(&consents)

	if result.Error != nil {
		r.logger.Error("Error listando consentimientos",
			zap.Error(result.Error),
			zap.String("user_id", userID.String()),
		)
		return nil, fmt.Errorf("error finding consents: %w", result.Error)
	}

	return consents, nil
}

// FindLatestActive busca el consentimiento vigente (no revocado) más reciente de un tipo.
func (r *ConsentRepositoryImpl) FindLatestActive(ctx context.Context, userID uuid.UUID, consentType string) (*entities.UserConsent, error) {
	var consent entities.UserConsent

	result := dbConn(ctx, r.db).
		Where("user_id = ? AND consent_type = ? AND withdrawn_at IS NULL", userID, consentType).
		Order("accepted_at DESC").
		First(&consent)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		r.logger.Error("Error buscando consentimiento activo",
			zap.Error(result.Error),
			zap.String("user_id", userID.String()),
			zap.String("consent_type", consentType),
		)
		return nil, fmt.Errorf("error finding active consent: %w", result.Error)
	}

	return &consent, nil
}

// WithdrawActive marca como revocado el consentimiento vigente de un tipo.
func (r *ConsentRepositoryImpl) WithdrawActive(ctx context.Context, userID uuid.UUID, consentType string) error {
	r.logger.Debug("Revocando consentimiento activo",
		zap.String("user_id", userID.String()),
		zap.String("consent_type", consentType),
	)

	result := dbConn(ctx, r.db).
		Model(&entities.UserConsent{}).
		Where("user_id = ? AND consent_type = ? AND withdrawn_at IS NULL", userID, consentType).
		Update("withdrawn_at", gorm.Expr("NOW()"))

	if result.Error != nil {
		r.logger.Error("Error revocando consentimiento",
			zap.Error(result.Error),
			zap.String("user_id", userID.String()),
			zap.String("consent_type", consentType),
		)
		return fmt.Errorf("error withdrawing consent: %w", result.Error)
	}

	r.logger.Info("Consentimiento revocado",
		zap.String("user_id", userID.String()),
		zap.String("consent_type", consentType),
		zap.Int64("rows_affected", result.RowsAffected),
	)

	return nil
}

// ========================================
// COMPILE-TIME INTERFACE CHECK
// ========================================

var _ repositories.ConsentRepository = (*ConsentRepositoryImpl)(nil)
