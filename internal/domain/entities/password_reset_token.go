package entities

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetToken representa un token de restablecimiento de contraseña.
// Solo se persiste el hash del token, nunca el valor en claro.
type PasswordResetToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	TokenHash string     `gorm:"not null;size:255;uniqueIndex" json:"-"`
	ExpiresAt time.Time  `gorm:"not null;index" json:"expires_at"`
	Used      bool       `gorm:"not null;default:false" json:"used"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// TableName especifica el nombre de la tabla con su schema.
func (PasswordResetToken) TableName() string {
	return "auth.password_reset_tokens"
}

// IsExpired indica si el token ya venció.
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsUsable indica si el token sigue siendo válido para restablecer la contraseña.
func (t *PasswordResetToken) IsUsable() bool {
	return !t.Used && !t.IsExpired()
}
