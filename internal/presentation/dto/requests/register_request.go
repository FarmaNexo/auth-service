// internal/presentation/dto/requests/register_request.go
package requests

// RegisterRequest representa el request de registro de usuario
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required,min=3"`
	Phone    string `json:"phone,omitempty" validate:"omitempty,min=7"`
}
