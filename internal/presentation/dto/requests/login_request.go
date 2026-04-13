// internal/presentation/dto/requests/login_request.go
package requests

import "strings"

// LoginRequest representa el request de login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Sanitize normaliza el request en el boundary del controller.
func (r *LoginRequest) Sanitize() {
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
	r.Password = strings.TrimSpace(r.Password)
}
