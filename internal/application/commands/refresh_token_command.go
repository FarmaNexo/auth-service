// internal/application/commands/refresh_token_command.go
package commands

import "github.com/farmanexo/auth-service/internal/presentation/dto/responses"

// RefreshTokenCommand representa el comando para refrescar tokens
type RefreshTokenCommand struct {
	RefreshToken string
}

// GetName retorna el nombre del comando
func (c RefreshTokenCommand) GetName() string {
	return "RefreshTokenCommand"
}

// RefreshTokenResponse es el alias de tipo para la respuesta (SIN datos de usuario)
type RefreshTokenResponse = responses.LoginResponse
