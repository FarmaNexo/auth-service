// internal/domain/entities/refresh_token.go
package entities

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken representa un token de refresco en el dominio
type RefreshToken struct {
	ID     uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	User   *User     `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"user,omitempty"`

	TokenHash string    `gorm:"not null;uniqueIndex;size:255" json:"-"`
	ExpiresAt time.Time `gorm:"not null;index" json:"expires_at"`

	IsRevoked bool       `gorm:"default:false;index" json:"is_revoked"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// Metadata
	IPAddress string `gorm:"size:50" json:"ip_address,omitempty"`
	UserAgent string `gorm:"type:text" json:"user_agent,omitempty"`

	CreatedAt time.Time `json:"created_at"`
}

// TableName especifica el nombre de la tabla
func (RefreshToken) TableName() string {
	return "auth.refresh_tokens"
}

// ========================================
// DOMAIN METHODS
// ========================================

// IsExpired verifica si el token ha expirado
func (rt *RefreshToken) IsExpired() bool {
	return time.Now().After(rt.ExpiresAt)
}

// IsValid verifica si el token es válido (no revocado y no expirado)
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsRevoked && !rt.IsExpired()
}

// Revoke revoca el token
func (rt *RefreshToken) Revoke() {
	now := time.Now()
	rt.IsRevoked = true
	rt.RevokedAt = &now
}

// DaysUntilExpiration retorna los días hasta que expire
func (rt *RefreshToken) DaysUntilExpiration() int {
	duration := time.Until(rt.ExpiresAt)
	return int(duration.Hours() / 24)
}

// ========================================
// FACTORY METHODS
// ========================================

// NewRefreshToken crea un nuevo refresh token
func NewRefreshToken(userID uuid.UUID, tokenHash string, expiresAt time.Time) *RefreshToken {
	return &RefreshToken{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		IsRevoked: false,
		CreatedAt: time.Now(),
	}
}

// NewRefreshTokenWithMetadata crea token con metadata
func NewRefreshTokenWithMetadata(
	userID uuid.UUID,
	tokenHash string,
	expiresAt time.Time,
	ipAddress string,
	userAgent string,
) *RefreshToken {
	token := NewRefreshToken(userID, tokenHash, expiresAt)
	token.IPAddress = ipAddress
	token.UserAgent = userAgent
	return token
}
