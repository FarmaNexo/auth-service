// internal/application/handlers/logout_handler.go
package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"

	"github.com/farmanexo/auth-service/internal/application/commands"
	"github.com/farmanexo/auth-service/internal/domain/repositories"
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
	"github.com/farmanexo/auth-service/internal/shared/common"
	"github.com/farmanexo/auth-service/internal/shared/constants"
	"github.com/farmanexo/auth-service/pkg/mediator"
	"go.uber.org/zap"
)

// LogoutHandler maneja el comando LogoutCommand
type LogoutHandler struct {
	tokenRepo repositories.TokenRepository
	logger    *zap.Logger
}

// NewLogoutHandler crea una nueva instancia del handler
func NewLogoutHandler(
	tokenRepo repositories.TokenRepository,
	logger *zap.Logger,
) *LogoutHandler {
	return &LogoutHandler{
		tokenRepo: tokenRepo,
		logger:    logger,
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

	h.logger.Info("Logout exitoso - Refresh token revocado",
		zap.String("user_id", command.UserID),
		zap.String("token_id", storedToken.ID.String()),
	)

	return h.successResponse(), nil
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
