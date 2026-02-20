// internal/application/validators/login_validator.go
package validators

import (
	"context"
	"errors"
	"regexp"

	"github.com/farmanexo/auth-service/internal/application/commands"
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
	"github.com/farmanexo/auth-service/pkg/mediator"
)

// LoginValidator valida el comando LoginCommand
type LoginValidator struct {
	emailRegex *regexp.Regexp
}

// NewLoginValidator crea un nuevo validador
func NewLoginValidator() *LoginValidator {
	return &LoginValidator{
		emailRegex: regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`),
	}
}

// Validate ejecuta las validaciones
func (v *LoginValidator) Validate(ctx context.Context, cmd commands.LoginCommand) error {
	// Validar formato de email
	if !v.emailRegex.MatchString(cmd.Email) {
		return errors.New("formato de email inválido")
	}

	// Validar longitud mínima de password
	if len(cmd.Password) < 8 {
		return errors.New("password debe tener al menos 8 caracteres")
	}

	return nil
}

// Asegurar que implementa la interfaz Validator
var _ mediator.Validator[commands.LoginCommand, responses.LoginResponse] = (*LoginValidator)(nil)
