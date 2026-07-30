package requests

import "strings"

// ForgotPasswordRequest representa el request de solicitud de restablecimiento.
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// Sanitize normaliza el request en el boundary del controller.
func (r *ForgotPasswordRequest) Sanitize() {
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
}
