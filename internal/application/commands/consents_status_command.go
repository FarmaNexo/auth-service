// internal/application/commands/consents_status_command.go
package commands

import "github.com/farmanexo/auth-service/internal/presentation/dto/responses"

// GetConsentsStatusCommand evalúa si los consentimientos vigentes del usuario
// están al día respecto a las versiones actuales publicadas. Usado para forzar
// re-aceptación cuando se actualiza un documento legal (HU-011).
type GetConsentsStatusCommand struct {
	UserID string
}

func (c GetConsentsStatusCommand) GetName() string {
	return "GetConsentsStatusCommand"
}

type GetConsentsStatusResponse = responses.ConsentsStatusResponse

// AcceptConsentsCommand registra aceptación explícita de consentimientos pendientes
// (ej. tras bump de versión). Se insertan nuevos registros — los viejos quedan como
// auditoría.
type AcceptConsentsCommand struct {
	UserID       string
	ConsentTypes []string // Sólo terms_of_service y privacy_policy son válidos aquí
	IPAddress    string
	UserAgent    string
}

func (c AcceptConsentsCommand) GetName() string {
	return "AcceptConsentsCommand"
}

type AcceptConsentsResponse = responses.EmptyResponse
