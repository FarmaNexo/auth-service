package requests

import "strings"

// ResetPasswordRequest representa el request de restablecimiento de contraseña.
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// Sanitize normaliza el request en el boundary del controller.
func (r *ResetPasswordRequest) Sanitize() {
	r.Token = strings.TrimSpace(r.Token)
	r.NewPassword = strings.TrimSpace(r.NewPassword)
}
