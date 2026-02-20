// internal/application/handlers/refresh_token_handler.go
package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/farmanexo/auth-service/internal/application/commands"
	"github.com/farmanexo/auth-service/internal/domain/repositories"
	"github.com/farmanexo/auth-service/internal/infrastructure/security"
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
	"github.com/farmanexo/auth-service/internal/shared/common"
	"github.com/farmanexo/auth-service/internal/shared/constants"
	"github.com/farmanexo/auth-service/pkg/mediator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RefreshTokenHandler maneja el comando RefreshTokenCommand
type RefreshTokenHandler struct {
	userRepo   repositories.UserRepository
	tokenRepo  repositories.TokenRepository
	jwtService security.JWTService
	logger     *zap.Logger
}

// NewRefreshTokenHandler crea una nueva instancia del handler
func NewRefreshTokenHandler(
	userRepo repositories.UserRepository,
	tokenRepo repositories.TokenRepository,
	jwtService security.JWTService,
	logger *zap.Logger,
) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtService: jwtService,
		logger:     logger,
	}
}

// Handle procesa el comando de refresh token
func (h *RefreshTokenHandler) Handle(
	ctx context.Context,
	command commands.RefreshTokenCommand,
) (*common.ApiResponse[responses.LoginResponse], error) {

	h.logger.Info("Procesando refresh token",
		zap.String("request_name", command.GetName()),
	)

	// 1. Validar JWT del refresh token
	userID, _, err := h.jwtService.ValidateRefreshToken(command.RefreshToken)
	if err != nil {
		h.logger.Warn("Refresh token JWT inválido", zap.Error(err))
		return h.unauthorizedResponse("Refresh token inválido o expirado"), nil
	}

	// 2. Hash del token para buscar en BD
	tokenHash := h.hashToken(command.RefreshToken)

	// 3. Buscar token en BD (verifica que no esté revocado ni expirado)
	storedToken, err := h.tokenRepo.FindByToken(ctx, tokenHash)
	if err != nil {
		h.logger.Warn("Refresh token no encontrado en BD",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return h.unauthorizedResponse("Refresh token inválido o expirado"), nil
	}

	// 4. Verificar que no esté revocado (doble verificación)
	if storedToken.IsRevoked {
		h.logger.Warn("Intento de usar refresh token revocado",
			zap.String("user_id", userID),
			zap.String("token_id", storedToken.ID.String()),
		)
		return h.unauthorizedResponse("Refresh token revocado"), nil
	}

	// 5. Buscar usuario
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.logger.Error("Error parseando user ID del token",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return common.InternalServerErrorResponse[responses.LoginResponse](
			"Error procesando solicitud",
		), nil
	}

	user, err := h.userRepo.FindByID(ctx, userUUID)
	if err != nil {
		h.logger.Error("Usuario no encontrado para refresh token",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return h.unauthorizedResponse("Usuario no encontrado"), nil
	}

	// 6. Verificar que el usuario esté activo
	if !user.IsActive {
		h.logger.Warn("Intento de refresh con usuario inactivo",
			zap.String("user_id", user.ID.String()),
			zap.String("email", user.Email),
		)
		return h.unauthorizedResponse("Usuario inactivo"), nil
	}

	// 7. Revocar el refresh token viejo
	if err := h.tokenRepo.RevokeToken(ctx, storedToken.ID); err != nil {
		h.logger.Error("Error revocando token viejo",
			zap.String("token_id", storedToken.ID.String()),
			zap.Error(err),
		)
		// Continuar de todas formas - el token nuevo reemplazará al viejo
	}

	// 8. Generar nuevo access token
	accessToken, accessExpiry, err := h.jwtService.GenerateAccessToken(
		user.ID.String(),
		user.Email,
		user.FullName,
		user.Role,
	)
	if err != nil {
		h.logger.Error("Error generando nuevo access token",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		return common.InternalServerErrorResponse[responses.LoginResponse](
			"Error generando token de acceso",
		), nil
	}

	// 9. Generar nuevo refresh token
	newRefreshToken, refreshExpiry, newTokenID, err := h.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		h.logger.Error("Error generando nuevo refresh token",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		return common.InternalServerErrorResponse[responses.LoginResponse](
			"Error generando refresh token",
		), nil
	}

	// 10. Guardar nuevo refresh token en BD
	newTokenHash := h.hashToken(newRefreshToken)
	if err := h.tokenRepo.CreateRefreshToken(
		ctx,
		user.ID,
		newTokenID,
		newTokenHash,
		refreshExpiry,
		"", // IP address - se puede obtener del contexto
		"", // User agent - se puede obtener del contexto
	); err != nil {
		h.logger.Error("Error guardando nuevo refresh token",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		return common.InternalServerErrorResponse[responses.LoginResponse](
			"Error procesando renovación de token",
		), nil
	}

	h.logger.Info("Tokens renovados exitosamente",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
	)

	// 11. Construir respuesta (SIN datos de usuario)
	expiresIn := int64(time.Until(accessExpiry).Seconds())

	loginResponse := responses.NewLoginResponse(
		accessToken,
		newRefreshToken,
		expiresIn,
	)

	return common.OkResponse(*loginResponse), nil
}

// hashToken genera un SHA-256 hash del token
func (h *RefreshTokenHandler) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// unauthorizedResponse crea una respuesta 401 Unauthorized
func (h *RefreshTokenHandler) unauthorizedResponse(message string) *common.ApiResponse[responses.LoginResponse] {
	response := common.NewApiResponse[responses.LoginResponse]()
	response.SetHttpStatus(constants.StatusUnauthorized.Int())
	response.AddError(constants.CodeUnauthorized, message)
	return response
}

// Asegurar que implementa la interfaz RequestHandler
var _ mediator.RequestHandler[commands.RefreshTokenCommand, responses.LoginResponse] = (*RefreshTokenHandler)(nil)
