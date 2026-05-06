// internal/application/commands/delete_account_command.go
package commands

import "github.com/farmanexo/auth-service/internal/presentation/dto/responses"

// DeleteAccountCommand representa el ejercicio del derecho ARCO de cancelación
// por parte del usuario (LPDP Ley 29733).
// Efectos:
//   - Soft-delete del usuario (auth.users.deleted_at)
//   - Anonimización de PII (email, full_name, phone)
//   - Revocación de todos los refresh tokens vigentes
//   - Publicación del evento USER_DELETED para que los demás servicios limpien proyecciones
type DeleteAccountCommand struct {
	UserID string // Del contexto (AuthMiddleware)
}

func (c DeleteAccountCommand) GetName() string {
	return "DeleteAccountCommand"
}

// DeleteAccountResponse alias sin datos sensibles
type DeleteAccountResponse = responses.EmptyResponse
