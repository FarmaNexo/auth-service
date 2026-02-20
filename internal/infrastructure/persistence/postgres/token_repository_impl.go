// internal/infrastructure/persistence/postgres/token_repository_impl.go
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

// TokenRepositoryImpl implementa TokenRepository usando PostgreSQL
type TokenRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewTokenRepository crea una nueva instancia del repositorio
func NewTokenRepository(db *gorm.DB, logger *zap.Logger) repositories.TokenRepository {
	return &TokenRepositoryImpl{
		db:     db,
		logger: logger,
	}
}

// ========================================
// REPOSITORY METHODS
// ========================================

// Create crea un nuevo refresh token
func (r *TokenRepositoryImpl) Create(ctx context.Context, token *entities.RefreshToken) error {
	r.logger.Debug("Creando refresh token",
		zap.String("user_id", token.UserID.String()),
	)

	result := r.db.WithContext(ctx).Create(token)
	if result.Error != nil {
		r.logger.Error("Error creando refresh token",
			zap.Error(result.Error),
			zap.String("user_id", token.UserID.String()),
		)
		return fmt.Errorf("error creating refresh token: %w", result.Error)
	}

	r.logger.Info("Refresh token creado",
		zap.String("token_id", token.ID.String()),
		zap.String("user_id", token.UserID.String()),
	)

	return nil
}

// CreateRefreshToken crea un refresh token con parámetros individuales
func (r *TokenRepositoryImpl) CreateRefreshToken(
	ctx context.Context,
	userID uuid.UUID,
	tokenID string,
	tokenHash string,
	expiresAt time.Time,
	ipAddress string,
	userAgent string,
) error {
	r.logger.Debug("Creando refresh token",
		zap.String("user_id", userID.String()),
	)

	token := &entities.RefreshToken{
		ID:        uuid.MustParse(tokenID),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		IsRevoked: false,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	}

	result := r.db.WithContext(ctx).Create(token)
	if result.Error != nil {
		r.logger.Error("Error creando refresh token",
			zap.Error(result.Error),
			zap.String("user_id", userID.String()),
		)
		return result.Error
	}

	r.logger.Info("Refresh token creado",
		zap.String("token_id", token.ID.String()),
		zap.String("user_id", userID.String()),
	)

	return nil
}

// FindByToken busca token por su hash
func (r *TokenRepositoryImpl) FindByToken(ctx context.Context, tokenHash string) (*entities.RefreshToken, error) {
	r.logger.Debug("Buscando refresh token por hash")

	var token entities.RefreshToken
	result := r.db.WithContext(ctx).
		Where("token_hash = ? AND is_revoked = false", tokenHash).
		First(&token)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.Debug("Refresh token no encontrado")
			return nil, ErrTokenNotFound
		}
		r.logger.Error("Error buscando refresh token",
			zap.Error(result.Error),
		)
		return nil, fmt.Errorf("error finding refresh token: %w", result.Error)
	}

	// Verificar si expiró
	if time.Now().After(token.ExpiresAt) {
		r.logger.Warn("Refresh token expirado",
			zap.String("token_id", token.ID.String()),
			zap.Time("expires_at", token.ExpiresAt),
		)
		return nil, ErrTokenExpired
	}

	return &token, nil
}

// RevokeToken revoca un token específico
func (r *TokenRepositoryImpl) RevokeToken(ctx context.Context, tokenID uuid.UUID) error {
	r.logger.Debug("Revocando refresh token",
		zap.String("token_id", tokenID.String()),
	)

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&entities.RefreshToken{}).
		Where("id = ?", tokenID).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": now,
		})

	if result.Error != nil {
		r.logger.Error("Error revocando token",
			zap.Error(result.Error),
		)
		return fmt.Errorf("error revoking token: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrTokenNotFound
	}

	r.logger.Info("Token revocado exitosamente",
		zap.Int64("rows_affected", result.RowsAffected),
	)

	return nil
}

// RevokeAllUserTokens revoca todos los tokens de un usuario
func (r *TokenRepositoryImpl) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	r.logger.Debug("Revocando todos los tokens del usuario",
		zap.String("user_id", userID.String()),
	)

	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&entities.RefreshToken{}).
		Where("user_id = ? AND is_revoked = false", userID).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": now,
		})

	if result.Error != nil {
		r.logger.Error("Error revocando tokens del usuario",
			zap.Error(result.Error),
			zap.String("user_id", userID.String()),
		)
		return fmt.Errorf("error revoking user tokens: %w", result.Error)
	}

	r.logger.Info("Tokens del usuario revocados",
		zap.String("user_id", userID.String()),
		zap.Int64("tokens_revoked", result.RowsAffected),
	)

	return nil
}

// DeleteExpiredTokens elimina tokens expirados (cleanup)
func (r *TokenRepositoryImpl) DeleteExpiredTokens(ctx context.Context) (int64, error) {
	r.logger.Debug("Eliminando tokens expirados")

	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&entities.RefreshToken{})

	if result.Error != nil {
		r.logger.Error("Error eliminando tokens expirados",
			zap.Error(result.Error),
		)
		return 0, fmt.Errorf("error deleting expired tokens: %w", result.Error)
	}

	r.logger.Info("Tokens expirados eliminados",
		zap.Int64("tokens_deleted", result.RowsAffected),
	)

	return result.RowsAffected, nil
}

// ========================================
// DOMAIN ERRORS
// ========================================

var (
	ErrTokenNotFound = errors.New("refresh token not found")
	ErrTokenExpired  = errors.New("refresh token expired")
	ErrTokenRevoked  = errors.New("refresh token revoked")
)

// ========================================
// COMPILE-TIME INTERFACE CHECK
// ========================================

var _ repositories.TokenRepository = (*TokenRepositoryImpl)(nil)
