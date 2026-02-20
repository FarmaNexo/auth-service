// internal/infrastructure/cache/token_blacklist_impl.go
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/farmanexo/auth-service/internal/domain/services"
	"go.uber.org/zap"
)

const blacklistKeyPrefix = "blacklist:token:"

// RedisTokenBlacklist implementa TokenBlacklist usando Redis
type RedisTokenBlacklist struct {
	redisClient *RedisClient
	logger      *zap.Logger
}

// NewRedisTokenBlacklist crea una nueva instancia de RedisTokenBlacklist
func NewRedisTokenBlacklist(redisClient *RedisClient, logger *zap.Logger) *RedisTokenBlacklist {
	return &RedisTokenBlacklist{
		redisClient: redisClient,
		logger:      logger,
	}
}

// BlacklistToken agrega un token a la blacklist con TTL automático
func (b *RedisTokenBlacklist) BlacklistToken(ctx context.Context, jti string, expiration time.Duration) error {
	key := fmt.Sprintf("%s%s", blacklistKeyPrefix, jti)

	err := b.redisClient.Client.Set(ctx, key, "1", expiration).Err()
	if err != nil {
		b.logger.Error("Error agregando token a blacklist",
			zap.String("jti", jti),
			zap.Duration("ttl", expiration),
			zap.Error(err),
		)
		return fmt.Errorf("error agregando token a blacklist: %w", err)
	}

	b.logger.Info("Token agregado a blacklist",
		zap.String("jti", jti),
		zap.Duration("ttl", expiration),
	)

	return nil
}

// IsBlacklisted verifica si un token está en la blacklist
func (b *RedisTokenBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	key := fmt.Sprintf("%s%s", blacklistKeyPrefix, jti)

	result, err := b.redisClient.Client.Exists(ctx, key).Result()
	if err != nil {
		b.logger.Error("Error verificando blacklist de token",
			zap.String("jti", jti),
			zap.Error(err),
		)
		return false, fmt.Errorf("error verificando blacklist: %w", err)
	}

	return result > 0, nil
}

// Asegurar que implementa la interfaz TokenBlacklist
var _ services.TokenBlacklist = (*RedisTokenBlacklist)(nil)
