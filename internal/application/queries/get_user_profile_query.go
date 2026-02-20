// internal/application/queries/get_user_profile_query.go
package queries

import "github.com/farmanexo/auth-service/internal/presentation/dto/responses"

// GetProfileQuery representa la query para obtener el perfil del usuario autenticado
type GetProfileQuery struct {
	UserID string
}

// GetName retorna el nombre de la query
func (q GetProfileQuery) GetName() string {
	return "GetProfileQuery"
}

// GetProfileResponse es el alias de tipo para la respuesta
type GetProfileResponse = responses.UserResponse
