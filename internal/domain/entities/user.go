// internal/domain/entities/user.go
package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User representa la entidad de dominio Usuario
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Email        string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	PasswordHash string    `gorm:"not null;size:255" json:"-"` // No exponer en JSON
	FullName     string    `gorm:"size:255" json:"full_name"`
	Phone        string    `gorm:"size:50" json:"phone,omitempty"`

	// Status flags
	IsActive        bool       `gorm:"default:true" json:"is_active"`
	IsVerified      bool       `gorm:"default:false" json:"is_verified"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`

	// Role and permissions
	Role string `gorm:"size:50;default:user" json:"role"` // user, admin, pharmacy_owner

	// Login tracking
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	LoginCount  int        `gorm:"default:0" json:"login_count"`

	// Audit fields
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName especifica el nombre de la tabla
func (User) TableName() string {
	return "auth.users"
}

// ========================================
// DOMAIN METHODS
// ========================================

// IsAccountActive verifica si la cuenta está activa y no eliminada
func (u *User) IsAccountActive() bool {
	return u.IsActive && u.DeletedAt.Time.IsZero()
}

// CanLogin verifica si el usuario puede iniciar sesión
func (u *User) CanLogin() bool {
	return u.IsActive && !u.DeletedAt.Valid
}

// Activate activa el usuario
func (u *User) Activate() {
	u.IsActive = true
}

// Deactivate desactiva el usuario
func (u *User) Deactivate() {
	u.IsActive = false
}

// VerifyEmail marca el email como verificado
func (u *User) VerifyEmail() {
	now := time.Now()
	u.IsVerified = true
	u.EmailVerifiedAt = &now
}

// RecordLogin registra un login exitoso
func (u *User) RecordLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.LoginCount++
}

// IsAdmin verifica si el usuario es administrador
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// IsPharmacyOwner verifica si es dueño de farmacia
func (u *User) IsPharmacyOwner() bool {
	return u.Role == "pharmacy_owner"
}

// ========================================
// FACTORY METHODS
// ========================================

// NewUser crea un nuevo usuario con valores por defecto
func NewUser(email, passwordHash, fullName string) *User {
	return &User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: passwordHash,
		FullName:     fullName,
		IsActive:     true,
		IsVerified:   false,
		Role:         "user",
		LoginCount:   0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// NewUserWithPhone crea usuario con teléfono
func NewUserWithPhone(email, passwordHash, fullName, phone string) *User {
	user := NewUser(email, passwordHash, fullName)
	user.Phone = phone
	return user
}
