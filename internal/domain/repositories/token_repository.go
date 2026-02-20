// internal/domain/repositories/token_repository.go
package repositories

import (
	"context"
	"time"

	"github.com/farmanexo/auth-service/internal/domain/entities"
	"github.com/google/uuid"
)

// TokenRepository define las operaciones de persistencia para tokens
type TokenRepository interface {
	// Create crea un nuevo refresh token (método legacy)
	Create(ctx context.Context, token *entities.RefreshToken) error

	// CreateRefreshToken crea un refresh token con parámetros individuales
	CreateRefreshToken(
		ctx context.Context,
		userID uuid.UUID,
		tokenID string,
		tokenHash string,
		expiresAt time.Time,
		ipAddress string,
		userAgent string,
	) error

	// FindByToken busca un token por su hash
	FindByToken(ctx context.Context, tokenHash string) (*entities.RefreshToken, error)

	// RevokeToken revoca un token
	RevokeToken(ctx context.Context, tokenID uuid.UUID) error

	// RevokeAllUserTokens revoca todos los tokens de un usuario
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error

	// DeleteExpiredTokens elimina tokens expirados
	DeleteExpiredTokens(ctx context.Context) (int64, error)
}
