// internal/presentation/dto/responses/auth_response.go
package responses

import (
	"time"

	"github.com/google/uuid"
)

// LoginResponse es la respuesta de login y refresh token (solo tokens)
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"` // Segundos
}

// RegisterResponse es la respuesta de registro (información mínima)
type RegisterResponse struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// NewLoginResponse crea una respuesta de login/refresh
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

// NewRegisterResponse crea una respuesta de registro
func NewRegisterResponse(
	userID uuid.UUID,
	email string,
	createdAt time.Time,
) *RegisterResponse {
	return &RegisterResponse{
		UserID:    userID,
		Email:     email,
		Message:   "Usuario registrado exitosamente. Por favor inicia sesión.",
		CreatedAt: createdAt,
	}
}
