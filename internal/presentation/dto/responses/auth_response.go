// internal/presentation/dto/responses/auth_response.go
package responses

import (
	"time"

	"github.com/google/uuid"
)

// AuthResponse es la respuesta COMPLETA de autenticación (para login/refresh)
// INCLUYE datos de usuario
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"` // Segundos
	User         UserResponse `json:"user"`
}

// LoginResponse es la respuesta SIMPLIFICADA de login (SIN datos de usuario)
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"` // Segundos
}

// RegisterResponse es la respuesta SIMPLIFICADA de registro (sin tokens)
type RegisterResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// UserResponse representa los datos del usuario en la respuesta
type UserResponse struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	FullName   string    `json:"full_name"`
	Phone      string    `json:"phone,omitempty"`
	Role       string    `json:"role"`
	IsVerified bool      `json:"is_verified"`
	CreatedAt  time.Time `json:"created_at"`
}

// NewAuthResponse crea una respuesta de autenticación COMPLETA (con usuario)
func NewAuthResponse(
	accessToken, refreshToken string,
	expiresIn int64,
	user UserResponse,
) *AuthResponse {
	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		User:         user,
	}
}

// NewLoginResponse crea una respuesta de login SIMPLIFICADA (sin usuario)
func NewLoginResponse(
	accessToken, refreshToken string,
	expiresIn int64,
) *LoginResponse {
	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}
}

// EmptyResponse representa una respuesta sin datos (usado en logout, etc.)
type EmptyResponse struct{}

// NewRegisterResponse crea una respuesta simplificada de registro
func NewRegisterResponse(
	userID uuid.UUID,
	email, fullName string,
	createdAt time.Time,
) *RegisterResponse {
	return &RegisterResponse{
		UserID:    userID,
		Email:     email,
		FullName:  fullName,
		Message:   "Usuario registrado exitosamente. Por favor inicia sesión.",
		CreatedAt: createdAt,
	}
}
