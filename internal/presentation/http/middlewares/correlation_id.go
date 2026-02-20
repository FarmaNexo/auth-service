// internal/presentation/http/middlewares/correlation_id.go
package middlewares

import (
	"context"
	"net/http"

	"github.com/farmanexo/auth-service/pkg/mediator"
	"github.com/google/uuid"
)

// CorrelationID es un middleware que agrega correlation ID a cada request
// para tracking distribuido
func CorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Generar o extraer correlation ID
		correlationID := r.Header.Get("X-Correlation-ID")
		if correlationID == "" {
			correlationID = uuid.New().String()
		}

		// Agregar al header de respuesta
		w.Header().Set("X-Correlation-ID", correlationID)

		// Agregar al contexto usando el mediator context
		ctx := mediator.WithValue(r.Context(), mediator.CorrelationKey, correlationID)

		// Continuar con el siguiente handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetCorrelationID obtiene el correlation ID del contexto
func GetCorrelationID(ctx context.Context) string {
	if val := ctx.Value(mediator.CorrelationKey); val != nil {
		if corrID, ok := val.(string); ok {
			return corrID
		}
	}
	return ""
}
