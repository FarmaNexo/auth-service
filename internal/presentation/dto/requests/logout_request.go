// internal/presentation/dto/requests/logout_request.go
package requests

// LogoutRequest representa el request para cerrar sesión
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
