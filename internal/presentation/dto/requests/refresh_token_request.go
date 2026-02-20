// internal/presentation/dto/requests/refresh_token_request.go
package requests

// RefreshTokenRequest representa el request para refrescar token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
