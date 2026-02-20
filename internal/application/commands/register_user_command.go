// internal/application/commands/register_user_command.go
package commands

import (
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
)

// RegisterUserCommand representa el comando para registrar un usuario
type RegisterUserCommand struct {
	Email    string
	Password string
	FullName string
	Phone    string
}

func (c RegisterUserCommand) GetName() string {
	return "RegisterUserCommand"
}

func (c RegisterUserCommand) Validate() error {
	if c.Email == "" {
		return ErrEmailRequired
	}
	if c.Password == "" {
		return ErrPasswordRequired
	}
	if len(c.Password) < 8 {
		return ErrPasswordTooShort
	}
	if c.FullName == "" {
		return ErrFullNameRequired
	}
	return nil
}

// ========================================
// DOMAIN ERRORS
// ========================================

var (
	ErrEmailRequired    = NewValidationError("email", "Email es requerido")
	ErrPasswordRequired = NewValidationError("password", "Password es requerido")
	ErrPasswordTooShort = NewValidationError("password", "Password debe tener al menos 8 caracteres")
	ErrFullNameRequired = NewValidationError("full_name", "Nombre completo es requerido")
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func NewValidationError(field, message string) ValidationError {
	return ValidationError{Field: field, Message: message}
}

// ========================================
// TYPE ALIAS - CAMBIAR A RegisterResponse
// ========================================

// RegisterUserResponse ahora devuelve RegisterResponse (sin tokens)
type RegisterUserResponse = responses.RegisterResponse
