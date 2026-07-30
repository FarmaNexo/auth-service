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

// ErrPasswordResetTokenNotFound se devuelve cuando no existe el token buscado.
var ErrPasswordResetTokenNotFound = errors.New("password reset token not found")

// PasswordResetTokenRepositoryImpl implementa PasswordResetTokenRepository con GORM.
type PasswordResetTokenRepositoryImpl struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewPasswordResetTokenRepository crea una nueva instancia del repositorio.
func NewPasswordResetTokenRepository(db *gorm.DB, logger *zap.Logger) repositories.PasswordResetTokenRepository {
	return &PasswordResetTokenRepositoryImpl{db: db, logger: logger}
}

func (r *PasswordResetTokenRepositoryImpl) Create(ctx context.Context, token *entities.PasswordResetToken) error {
	if result := dbConn(ctx, r.db).Create(token); result.Error != nil {
		r.logger.Error("Error creando token de reset", zap.Error(result.Error))
		return fmt.Errorf("error creating password reset token: %w", result.Error)
	}
	return nil
}

func (r *PasswordResetTokenRepositoryImpl) FindByTokenHash(ctx context.Context, tokenHash string) (*entities.PasswordResetToken, error) {
	var token entities.PasswordResetToken
	result := dbConn(ctx, r.db).Where("token_hash = ?", tokenHash).First(&token)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrPasswordResetTokenNotFound
		}
		r.logger.Error("Error buscando token de reset", zap.Error(result.Error))
		return nil, fmt.Errorf("error finding password reset token: %w", result.Error)
	}
	return &token, nil
}

func (r *PasswordResetTokenRepositoryImpl) MarkUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	result := dbConn(ctx, r.db).
		Model(&entities.PasswordResetToken{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"used": true, "used_at": now})
	if result.Error != nil {
		r.logger.Error("Error marcando token de reset como usado", zap.Error(result.Error))
		return fmt.Errorf("error marking password reset token used: %w", result.Error)
	}
	return nil
}

func (r *PasswordResetTokenRepositoryImpl) InvalidateUserTokens(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	result := dbConn(ctx, r.db).
		Model(&entities.PasswordResetToken{}).
		Where("user_id = ? AND used = false", userID).
		Updates(map[string]interface{}{"used": true, "used_at": now})
	if result.Error != nil {
		r.logger.Error("Error invalidando tokens de reset del usuario", zap.Error(result.Error))
		return fmt.Errorf("error invalidating user password reset tokens: %w", result.Error)
	}
	return nil
}

var _ repositories.PasswordResetTokenRepository = (*PasswordResetTokenRepositoryImpl)(nil)
