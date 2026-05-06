// internal/domain/repositories/consent_repository.go
package repositories

import (
	"context"

	"github.com/farmanexo/auth-service/internal/domain/entities"
	"github.com/google/uuid"
)

// ConsentRepository define el contrato para operaciones de persistencia de consentimientos.
// Parte del cumplimiento LPDP (Ley 29733): el usuario debe poder consultar y revocar
// los consentimientos que otorgó, y auditar la plataforma debe poder trazar cuándo se aceptó qué versión.
type ConsentRepository interface {
	// Create persiste un evento de consentimiento (aceptado o rechazado).
	// Nunca actualiza filas previas: cada aceptación es inmutable salvo Withdraw.
	Create(ctx context.Context, consent *entities.UserConsent) error

	// CreateBatch persiste múltiples consentimientos en una sola transacción.
	// Usado en el flujo de registro donde se guardan terms + privacy + marketing juntos.
	CreateBatch(ctx context.Context, consents []*entities.UserConsent) error

	// FindAllByUserID devuelve todos los consentimientos de un usuario (ordenados por fecha descendente).
	// Incluye revocados para que el usuario pueda ver su historial completo (derecho ARCO de acceso).
	FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]*entities.UserConsent, error)

	// FindLatestActive busca el consentimiento vigente más reciente de un tipo específico.
	// Retorna nil si el usuario nunca aceptó o ya revocó.
	FindLatestActive(ctx context.Context, userID uuid.UUID, consentType string) (*entities.UserConsent, error)

	// WithdrawActive revoca el consentimiento vigente de un tipo (marca withdrawn_at).
	// No-op si no hay consentimiento activo.
	WithdrawActive(ctx context.Context, userID uuid.UUID, consentType string) error
}
