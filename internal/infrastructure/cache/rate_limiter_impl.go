// internal/infrastructure/cache/rate_limiter_impl.go
package cache

import (
	"context"
	"time"

	"github.com/farmanexo/auth-service/internal/domain/services"
	"go.uber.org/zap"
)

// RedisRateLimiter implementa RateLimiter usando Redis con fixed-window counter
type RedisRateLimiter struct {
	redisClient *RedisClient
	logger      *zap.Logger
}

// NewRedisRateLimiter crea una nueva instancia de RedisRateLimiter
func NewRedisRateLimiter(redisClient *RedisClient, logger *zap.Logger) *RedisRateLimiter {
	return &RedisRateLimiter{
		redisClient: redisClient,
		logger:      logger,
	}
}

// Check verifica si una key ha excedido el límite en la ventana de tiempo
// Algoritmo fixed-window: INCR + EXPIRE atómico
func (r *RedisRateLimiter) Check(ctx context.Context, key string, limit int, window time.Duration) (*services.RateLimitResult, error) {
	// 1. Incrementar contador
	count, err := r.redisClient.Client.Incr(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	// 2. Si es la primera solicitud, establecer TTL
	if count == 1 {
		r.redisClient.Client.Expire(ctx, key, window)
	}

	// 3. Obtener TTL para calcular ResetAt
	ttl, err := r.redisClient.Client.TTL(ctx, key).Result()
	if err != nil {
		// Si falla obtener TTL, usamos la ventana completa como fallback
		ttl = window
	}

	// Si TTL retorna -1 (sin expiración), establecer expiración de seguridad
	if ttl < 0 {
		r.redisClient.Client.Expire(ctx, key, window)
		ttl = window
	}

	resetAt := time.Now().Add(ttl)
	remaining := limit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	result := &services.RateLimitResult{
		Allowed:   count <= int64(limit),
		Limit:     limit,
		Remaining: remaining,
		ResetAt:   resetAt,
	}

	if !result.Allowed {
		r.logger.Warn("Rate limit excedido",
			zap.String("key", key),
			zap.Int64("count", count),
			zap.Int("limit", limit),
			zap.Duration("reset_in", ttl),
		)
	}

	return result, nil
}

// Asegurar que implementa la interfaz RateLimiter
var _ services.RateLimiter = (*RedisRateLimiter)(nil)
