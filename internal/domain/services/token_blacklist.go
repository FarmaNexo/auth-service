// internal/domain/services/token_blacklist.go
package services

import (
	"context"
	"time"
)

// TokenBlacklist define la interfaz para gestionar la blacklist de access tokens
type TokenBlacklist interface {
	// BlacklistToken agrega un token a la blacklist con un tiempo de expiración
	BlacklistToken(ctx context.Context, jti string, expiration time.Duration) error

	// IsBlacklisted verifica si un token está en la blacklist
	IsBlacklisted(ctx context.Context, jti string) (bool, error)
}
