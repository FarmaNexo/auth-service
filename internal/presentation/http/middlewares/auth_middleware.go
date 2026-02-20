// internal/presentation/http/middlewares/auth_middleware.go
package middlewares

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/farmanexo/auth-service/internal/domain/services"
	"github.com/farmanexo/auth-service/internal/infrastructure/security"
	"github.com/farmanexo/auth-service/internal/shared/common"
	"github.com/farmanexo/auth-service/internal/shared/constants"
	"github.com/farmanexo/auth-service/pkg/mediator"
	"go.uber.org/zap"
)

// contextKey tipo dedicado para claves de contexto del middleware
type contextKey string

const (
	// UserIDCtxKey clave para el user ID en el contexto
	UserIDCtxKey contextKey = "user_id"
	// UserRoleCtxKey clave para el role en el contexto
	UserRoleCtxKey contextKey = "user_role"
	// AccessTokenCtxKey clave para el access token raw en el contexto
	AccessTokenCtxKey contextKey = "access_token"
)

// AuthMiddleware maneja la autenticación JWT en rutas protegidas
type AuthMiddleware struct {
	jwtService     security.JWTService
	tokenBlacklist services.TokenBlacklist
	logger         *zap.Logger
}

// NewAuthMiddleware crea una nueva instancia del middleware de autenticación
func NewAuthMiddleware(jwtService security.JWTService, tokenBlacklist services.TokenBlacklist, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:     jwtService,
		tokenBlacklist: tokenBlacklist,
		logger:         logger,
	}
}

// RequireAuth middleware que valida el access token JWT
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Extraer Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Warn("Request sin header Authorization",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method),
			)
			m.respondUnauthorized(w, "Header Authorization es requerido")
			return
		}

		// 2. Validar formato "Bearer {token}"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			m.logger.Warn("Formato de Authorization inválido",
				zap.String("path", r.URL.Path),
			)
			m.respondUnauthorized(w, "Formato de token inválido. Use: Bearer {token}")
			return
		}

		tokenString := parts[1]
		if tokenString == "" {
			m.respondUnauthorized(w, "Token vacío")
			return
		}

		// 3. Validar JWT
		userID, role, jti, err := m.jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			m.logger.Warn("Access token inválido",
				zap.String("path", r.URL.Path),
				zap.Error(err),
			)
			m.respondUnauthorized(w, "Token inválido o expirado")
			return
		}

		// 4. Verificar blacklist (fail-open: si Redis falla, permitimos el paso)
		blacklisted, err := m.tokenBlacklist.IsBlacklisted(r.Context(), jti)
		if err != nil {
			m.logger.Warn("Error verificando blacklist, permitiendo acceso (fail-open)",
				zap.String("jti", jti),
				zap.String("user_id", userID),
				zap.Error(err),
			)
		} else if blacklisted {
			m.logger.Warn("Access token revocado (blacklisted)",
				zap.String("jti", jti),
				zap.String("user_id", userID),
				zap.String("path", r.URL.Path),
			)
			m.respondUnauthorized(w, "Token revocado")
			return
		}

		m.logger.Debug("Token validado exitosamente",
			zap.String("user_id", userID),
			zap.String("role", role),
			zap.String("path", r.URL.Path),
		)

		// 5. Guardar user_id, role y access_token en el contexto
		ctx := context.WithValue(r.Context(), UserIDCtxKey, userID)
		ctx = context.WithValue(ctx, UserRoleCtxKey, role)
		ctx = context.WithValue(ctx, AccessTokenCtxKey, tokenString)

		// También guardar en el mediator context para uso en handlers
		ctx = mediator.WithValue(ctx, mediator.UserIDKey, userID)

		// 6. Continuar con el siguiente handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// respondUnauthorized escribe una respuesta 401 estandarizada
func (m *AuthMiddleware) respondUnauthorized(w http.ResponseWriter, message string) {
	response := common.NewApiResponse[struct{}]()
	response.SetHttpStatus(constants.StatusUnauthorized.Int())
	response.AddError(constants.CodeUnauthorized, message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(response)
}

// GetUserIDFromContext obtiene el user ID del contexto
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDCtxKey).(string)
	return userID, ok
}

// GetUserRoleFromContext obtiene el role del contexto
func GetUserRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleCtxKey).(string)
	return role, ok
}

// GetAccessTokenFromContext obtiene el access token raw del contexto
func GetAccessTokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(AccessTokenCtxKey).(string)
	return token, ok
}
