// internal/application/commands/get_my_consents_command.go
package commands

import "github.com/farmanexo/auth-service/internal/presentation/dto/responses"

// GetMyConsentsCommand lista todos los consentimientos del usuario autenticado.
// Aunque es una lectura, se modela como command para reutilizar el pipeline del mediator.
type GetMyConsentsCommand struct {
	UserID string
}

func (c GetMyConsentsCommand) GetName() string {
	return "GetMyConsentsCommand"
}

type GetMyConsentsResponse = responses.ConsentsListResponse
