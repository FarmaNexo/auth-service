// internal/application/commands/login_command.go
package commands

import (
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
)

// LoginCommand representa el comando para iniciar sesión
type LoginCommand struct {
	Email    string
	Password string
}

// GetName retorna el nombre del comando
func (c LoginCommand) GetName() string {
	return "LoginCommand"
}

// Validate valida el comando
func (c LoginCommand) Validate() error {
	if c.Email == "" {
		return ErrEmailRequired
	}
	if c.Password == "" {
		return ErrPasswordRequired
	}
	return nil
}

// LoginResponse es el alias de tipo para la respuesta
type LoginResponse = responses.LoginResponse // ← CAMBIO AQUÍ
