// internal/presentation/http/middlewares/rate_limit_middleware.go
package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/farmanexo/auth-service/internal/domain/services"
	"github.com/farmanexo/auth-service/internal/infrastructure/security"
	"github.com/farmanexo/auth-service/internal/shared/common"
	"github.com/farmanexo/auth-service/internal/shared/constants"
	"go.uber.org/zap"
)

// RateLimitMiddleware maneja el rate limiting en endpoints específicos
type RateLimitMiddleware struct {
	rateLimiter services.RateLimiter
	jwtService  security.JWTService
	logger      *zap.Logger
}

// NewRateLimitMiddleware crea una nueva instancia del middleware de rate limiting
func NewRateLimitMiddleware(rateLimiter services.RateLimiter, jwtService security.JWTService, logger *zap.Logger) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		rateLimiter: rateLimiter,
		jwtService:  jwtService,
		logger:      logger,
	}
}

// refreshRequestBody estructura parcial para extraer el refresh_token del body
type refreshRequestBody struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshRateLimit aplica rate limiting al endpoint de refresh token por user_id
func (m *RateLimitMiddleware) RefreshRateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 1. Leer body completo
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			m.logger.Warn("Error leyendo body para rate limit, continuando sin rate limit",
				zap.Error(err),
			)
			next.ServeHTTP(w, r)
			return
		}

		// 2. Restaurar body para que el handler pueda leerlo
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		// 3. Parsear JSON para extraer refresh_token
		requestBody := refreshRequestBody{}
		if err := json.Unmarshal(bodyBytes, &requestBody); err != nil || requestBody.RefreshToken == "" {
			m.logger.Debug("No se pudo extraer refresh_token del body, continuando sin rate limit",
				zap.Error(err),
			)
			next.ServeHTTP(w, r)
			return
		}

		// 4. Validar refresh token para obtener userID
		userID, _, err := m.jwtService.ValidateRefreshToken(requestBody.RefreshToken)
		if err != nil {
			// Si el token es inválido, dejamos que el handler maneje el error
			next.ServeHTTP(w, r)
			return
		}

		// 5. Verificar rate limit por userID
		key := "ratelimit:refresh:" + userID
		result, err := m.rateLimiter.Check(r.Context(), key, 10, 1*time.Hour)
		if err != nil {
			m.logger.Warn("Error verificando rate limit en refresh, permitiendo acceso (fail-open)",
				zap.String("user_id", userID),
				zap.Error(err),
			)
			next.ServeHTTP(w, r)
			return
		}

		// 6. Si excede el límite, retornar 429
		if !result.Allowed {
			m.respondTooManyRequests(w, "Demasiadas solicitudes de refresh token. Intente nuevamente más tarde", result)
			return
		}

		// 7. Setear headers de rate limit en la respuesta
		m.setRateLimitHeaders(w, result)

		// 8. Continuar con el siguiente handler
		next.ServeHTTP(w, r)
	})
}

// respondTooManyRequests escribe una respuesta 429 estandarizada
func (m *RateLimitMiddleware) respondTooManyRequests(w http.ResponseWriter, message string, result *services.RateLimitResult) {
	response := common.NewApiResponse[struct{}]()
	response.SetHttpStatus(constants.StatusTooManyRequests.Int())
	response.AddError(constants.CodeRateLimitExceeded, message)

	m.setRateLimitHeaders(w, result)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	json.NewEncoder(w).Encode(response)
}

// setRateLimitHeaders establece los headers estándar de rate limiting
func (m *RateLimitMiddleware) setRateLimitHeaders(w http.ResponseWriter, result *services.RateLimitResult) {
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(result.Limit))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
	w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", result.ResetAt.Unix()))
}
