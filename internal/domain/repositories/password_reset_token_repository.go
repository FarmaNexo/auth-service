package repositories

import (
	"context"

	"github.com/farmanexo/auth-service/internal/domain/entities"
	"github.com/google/uuid"
)

// PasswordResetTokenRepository gestiona los tokens de restablecimiento de contraseña.
type PasswordResetTokenRepository interface {
	// Create persiste un nuevo token de restablecimiento.
	Create(ctx context.Context, token *entities.PasswordResetToken) error
	// FindByTokenHash busca un token por su hash (sin filtrar por uso/expiración).
	FindByTokenHash(ctx context.Context, tokenHash string) (*entities.PasswordResetToken, error)
	// MarkUsed marca un token como consumido.
	MarkUsed(ctx context.Context, id uuid.UUID) error
	// InvalidateUserTokens marca como usados todos los tokens vigentes del usuario,
	// de modo que solicitar un nuevo restablecimiento invalida los enlaces anteriores.
	InvalidateUserTokens(ctx context.Context, userID uuid.UUID) error
}
