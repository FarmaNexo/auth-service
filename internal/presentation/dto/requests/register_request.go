// internal/presentation/dto/requests/register_request.go
package requests

import "strings"

// RegisterRequest representa el request de registro de usuario
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required,min=3"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,min=7"`
}

// Sanitize normaliza los campos del request: trim, lowercase del email y title case del nombre.
// Se ejecuta en el controller (boundary) antes de construir el command, así el resto del
// pipeline trabaja con datos ya limpios.
func (r *RegisterRequest) Sanitize() {
	r.Email = strings.ToLower(strings.TrimSpace(r.Email))
	r.FullName = titleCase(strings.TrimSpace(r.FullName))
	r.Phone = strings.TrimSpace(r.Phone)
	r.Password = strings.TrimSpace(r.Password)
}

func titleCase(s string) string {
	words := strings.Fields(s)
	for i, w := range words {
		if w == "" {
			continue
		}
		words[i] = strings.ToUpper(string(w[0])) + strings.ToLower(w[1:])
	}
	return strings.Join(words, " ")
}
