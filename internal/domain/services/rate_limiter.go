// internal/domain/services/rate_limiter.go
package services

import (
	"context"
	"time"
)

// RateLimitResult contiene el resultado de una verificación de rate limit
type RateLimitResult struct {
	Allowed   bool
	Limit     int
	Remaining int
	ResetAt   time.Time
}

// RateLimiter define la interfaz para el servicio de rate limiting
type RateLimiter interface {
	// Check verifica si una key ha excedido el límite en la ventana de tiempo
	Check(ctx context.Context, key string, limit int, window time.Duration) (*RateLimitResult, error)
}
