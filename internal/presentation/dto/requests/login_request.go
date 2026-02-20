// internal/presentation/dto/requests/login_request.go
package requests

// LoginRequest representa el request de login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}
