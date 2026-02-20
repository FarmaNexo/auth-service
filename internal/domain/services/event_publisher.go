// internal/domain/services/event_publisher.go
package services

import (
	"context"

	"github.com/farmanexo/auth-service/internal/domain/events"
)

// EventPublisher define la interfaz para publicar eventos de autenticación
type EventPublisher interface {
	// Publish publica un evento de autenticación
	Publish(ctx context.Context, event events.AuthEvent) error
}
