// internal/infrastructure/security/jwt_service_interface.go
package security

import (
	"time"
)

// JWTService define la interfaz para el servicio de JWT
type JWTService interface {
	// GenerateAccessToken genera un token de acceso (solo usa userID y role internamente)
	GenerateAccessToken(userID, email, fullName, role string) (token string, expiresAt time.Time, err error)

	// GenerateRefreshToken genera un token de refresco
	GenerateRefreshToken(userID string) (token string, expiresAt time.Time, tokenID string, err error)

	// ValidateAccessToken valida un token de acceso y retorna userID y role
	ValidateAccessToken(token string) (userID, role string, err error)

	// ValidateRefreshToken valida un token de refresco
	ValidateRefreshToken(token string) (userID, tokenID string, err error)
}
