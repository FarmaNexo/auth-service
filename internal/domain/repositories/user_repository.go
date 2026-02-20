// internal/domain/repositories/user_repository.go
package repositories

import (
	"context"

	"github.com/farmanexo/auth-service/internal/domain/entities"
	"github.com/google/uuid"
)

// UserRepository define el contrato para operaciones de persistencia de usuarios
// SOLID: Dependency Inversion - Depende de abstracción, no de implementación
type UserRepository interface {
	// Create crea un nuevo usuario
	Create(ctx context.Context, user *entities.User) error

	// FindByEmail busca usuario por email
	FindByEmail(ctx context.Context, email string) (*entities.User, error)

	// FindByID busca usuario por ID
	FindByID(ctx context.Context, id uuid.UUID) (*entities.User, error)

	// Update actualiza un usuario existente
	Update(ctx context.Context, user *entities.User) error

	// UpdateLoginInfo actualiza información de login
	UpdateLoginInfo(ctx context.Context, userID uuid.UUID) error

	// Exists verifica si existe un usuario con el email
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// Delete realiza soft delete del usuario
	Delete(ctx context.Context, userID uuid.UUID) error
}
