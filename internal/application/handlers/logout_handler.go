// internal/application/handlers/logout_handler.go
package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/farmanexo/auth-service/internal/application/commands"
	"github.com/farmanexo/auth-service/internal/domain/events"
	"github.com/farmanexo/auth-service/internal/domain/repositories"
	"github.com/farmanexo/auth-service/internal/domain/services"
	"github.com/farmanexo/auth-service/internal/infrastructure/security"
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
	"github.com/farmanexo/auth-service/internal/shared/common"
	"github.com/farmanexo/auth-service/internal/shared/constants"
	"github.com/farmanexo/auth-service/pkg/mediator"
	"go.uber.org/zap"
)

// LogoutHandler maneja el comando LogoutCommand
type LogoutHandler struct {
	tokenRepo      repositories.TokenRepository
	tokenBlacklist services.TokenBlacklist
	jwtService     security.JWTService
	eventPublisher services.EventPublisher
	logger         *zap.Logger
}

// NewLogoutHandler crea una nueva instancia del handler
func NewLogoutHandler(
	tokenRepo repositories.TokenRepository,
	tokenBlacklist services.TokenBlacklist,
	jwtService security.JWTService,
	eventPublisher services.EventPublisher,
	logger *zap.Logger,
) *LogoutHandler {
	return &LogoutHandler{
		tokenRepo:      tokenRepo,
		tokenBlacklist: tokenBlacklist,
		jwtService:     jwtService,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

// Handle procesa el comando de logout
func (h *LogoutHandler) Handle(
	ctx context.Context,
	command commands.LogoutCommand,
) (*common.ApiResponse[responses.EmptyResponse], error) {

	h.logger.Info("Procesando logout",
		zap.String("user_id", command.UserID),
		zap.String("request_name", command.GetName()),
	)

	// 1. Hash del refresh token para buscar en BD
	tokenHash := h.hashToken(command.RefreshToken)

	// 2. Buscar token en BD
	storedToken, err := h.tokenRepo.FindByToken(ctx, tokenHash)
	if err != nil {
		h.logger.Warn("Refresh token no encontrado en BD para logout",
			zap.String("user_id", command.UserID),
			zap.Error(err),
		)
		return h.unauthorizedResponse("Refresh token inválido o expirado"), nil
	}

	// 3. Verificar que el token pertenezca al usuario autenticado
	if storedToken.UserID.String() != command.UserID {
		h.logger.Warn("Intento de revocar token de otro usuario",
			zap.String("authenticated_user_id", command.UserID),
			zap.String("token_owner_id", storedToken.UserID.String()),
		)
		response := common.NewApiResponse[responses.EmptyResponse]()
		response.SetHttpStatus(constants.StatusForbidden.Int())
		response.AddError(constants.CodeUnauthorized, "El token no pertenece al usuario autenticado")
		return response, nil
	}

	// 4. Verificar que no esté ya revocado
	if storedToken.IsRevoked {
		h.logger.Info("Refresh token ya estaba revocado",
			zap.String("user_id", command.UserID),
			zap.String("token_id", storedToken.ID.String()),
		)
		// Retornar éxito de todas formas (idempotente)
		return h.successResponse(), nil
	}

	// 5. Revocar el refresh token
	if err := h.tokenRepo.RevokeToken(ctx, storedToken.ID); err != nil {
		h.logger.Error("Error revocando refresh token en logout",
			zap.String("user_id", command.UserID),
			zap.String("token_id", storedToken.ID.String()),
			zap.Error(err),
		)
		return common.InternalServerErrorResponse[responses.EmptyResponse](
			"Error cerrando sesión",
		), nil
	}

	// 6. Blacklistear el access token en Redis (best-effort)
	h.blacklistAccessToken(ctx, command.AccessToken, command.UserID)

	h.logger.Info("Logout exitoso - Refresh token revocado y access token blacklisteado",
		zap.String("user_id", command.UserID),
		zap.String("token_id", storedToken.ID.String()),
	)

	// Publicar evento de logout (fire-and-forget)
	go h.publishEvent(context.Background(), events.NewUserLogoutEvent(command.UserID))

	return h.successResponse(), nil
}

// publishEvent publica un evento de autenticación (best-effort)
func (h *LogoutHandler) publishEvent(ctx context.Context, event events.AuthEvent) {
	if err := h.eventPublisher.Publish(ctx, event); err != nil {
		h.logger.Warn("Error publicando evento de autenticación",
			zap.String("event_type", event.EventType),
			zap.String("user_id", event.UserID),
			zap.Error(err),
		)
	}
}

// blacklistAccessToken agrega el access token a la blacklist de Redis (best-effort)
func (h *LogoutHandler) blacklistAccessToken(ctx context.Context, accessToken string, userID string) {
	if accessToken == "" {
		h.logger.Warn("Access token vacío, no se puede blacklistear",
			zap.String("user_id", userID),
		)
		return
	}

	// Obtener JTI del access token
	_, _, jti, err := h.jwtService.ValidateAccessToken(accessToken)
	if err != nil {
		h.logger.Warn("No se pudo validar access token para blacklist",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return
	}

	// Obtener expiración para calcular TTL
	expiresAt, err := h.jwtService.GetAccessTokenExpiration(accessToken)
	if err != nil {
		h.logger.Warn("No se pudo obtener expiración del access token",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return
	}

	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		h.logger.Debug("Access token ya expirado, no es necesario blacklistear",
			zap.String("user_id", userID),
			zap.String("jti", jti),
		)
		return
	}

	// Agregar a blacklist (best-effort: si falla, solo logueamos warning)
	if err := h.tokenBlacklist.BlacklistToken(ctx, jti, ttl); err != nil {
		h.logger.Warn("Error agregando access token a blacklist (best-effort)",
			zap.String("user_id", userID),
			zap.String("jti", jti),
			zap.Error(err),
		)
		return
	}

	h.logger.Info("Access token agregado a blacklist",
		zap.String("user_id", userID),
		zap.String("jti", jti),
		zap.Duration("ttl", ttl),
	)
}

// hashToken genera un SHA-256 hash del token
func (h *LogoutHandler) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// successResponse crea una respuesta exitosa sin datos
func (h *LogoutHandler) successResponse() *common.ApiResponse[responses.EmptyResponse] {
	response := common.NewApiResponse[responses.EmptyResponse]()
	response.SetHttpStatus(constants.StatusOK.Int())
	response.AddMessage(constants.CodeLogoutSuccess, "Sesión cerrada exitosamente")
	return response
}

// unauthorizedResponse crea una respuesta 401 Unauthorized
func (h *LogoutHandler) unauthorizedResponse(message string) *common.ApiResponse[responses.EmptyResponse] {
	response := common.NewApiResponse[responses.EmptyResponse]()
	response.SetHttpStatus(constants.StatusUnauthorized.Int())
	response.AddError(constants.CodeUnauthorized, message)
	return response
}

// Asegurar que implementa la interfaz RequestHandler
var _ mediator.RequestHandler[commands.LogoutCommand, responses.EmptyResponse] = (*LogoutHandler)(nil)
