// internal/application/commands/logout_command.go
package commands

import "github.com/farmanexo/auth-service/internal/presentation/dto/responses"

// LogoutCommand representa el comando para cerrar sesión
type LogoutCommand struct {
	UserID       string // Del contexto (AuthMiddleware)
	RefreshToken string // Del body
	AccessToken  string // Del header Authorization (para blacklist)
}

// GetName retorna el nombre del comando
func (c LogoutCommand) GetName() string {
	return "LogoutCommand"
}

// LogoutResponse es el alias de tipo para la respuesta (sin datos)
type LogoutResponse = responses.EmptyResponse
