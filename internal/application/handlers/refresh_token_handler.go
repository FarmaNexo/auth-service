// internal/application/handlers/refresh_token_handler.go
package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
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
	txManager  repositories.TransactionManager
	userRepo   repositories.UserRepository
	tokenRepo  repositories.TokenRepository
	jwtService security.JWTService
	logger     *zap.Logger
}

// NewRefreshTokenHandler crea una nueva instancia del handler
func NewRefreshTokenHandler(
	txManager repositories.TransactionManager,
	userRepo repositories.UserRepository,
	tokenRepo repositories.TokenRepository,
	jwtService security.JWTService,
	logger *zap.Logger,
) *RefreshTokenHandler {
	return &RefreshTokenHandler{
		txManager:  txManager,
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

	// 3. Buscar token en BD INCLUYENDO revocados/expirados, para poder detectar reuso.
	storedToken, err := h.tokenRepo.FindAnyByTokenHash(ctx, tokenHash)
	if err != nil {
		h.logger.Warn("Refresh token no encontrado en BD",
			zap.String("user_id", userID),
			zap.Error(err),
		)
		return h.unauthorizedResponse("Refresh token inválido o expirado"), nil
	}

	// 4. BREACH DETECTION — reuso de un refresh token ya revocado.
	//    Con rotación de tokens, cada refresh exitoso revoca el token usado. Si un
	//    token revocado se vuelve a presentar, significa que (a) un atacante robó un
	//    token y el usuario legítimo ya rotó, o (b) el usuario legítimo presenta uno
	//    robado/reemplazado. En ambos casos se revoca TODA la familia de tokens del
	//    usuario para forzar re-login y cortar el acceso del atacante (patrón estándar
	//    de refresh token rotation reuse detection).
	if storedToken.IsRevoked {
		h.logger.Error("SECURITY_EVENT: reuso de refresh token revocado — revocando todas las sesiones del usuario",
			zap.String("security_event", "REFRESH_TOKEN_REUSE"),
			zap.String("user_id", storedToken.UserID.String()),
			zap.String("token_id", storedToken.ID.String()),
		)
		if err := h.tokenRepo.RevokeAllUserTokens(ctx, storedToken.UserID); err != nil {
			h.logger.Error("Error revocando todas las sesiones tras detección de reuso",
				zap.String("user_id", storedToken.UserID.String()),
				zap.Error(err),
			)
		}
		return h.unauthorizedResponse("Sesión inválida. Por seguridad se cerraron todas las sesiones; vuelve a iniciar sesión"), nil
	}

	// 4b. Verificar expiración (FindAnyByTokenHash no filtra por expiración).
	if storedToken.IsExpired() {
		h.logger.Warn("Refresh token expirado",
			zap.String("user_id", userID),
			zap.String("token_id", storedToken.ID.String()),
		)
		return h.unauthorizedResponse("Refresh token inválido o expirado"), nil
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

	// 7. Generar nuevo access token (cómputo puro, fuera de transacción)
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

	// 8. Generar nuevo refresh token (cómputo puro, fuera de transacción)
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

	// 9. Rotación atómica: revocar el token viejo y persistir el nuevo en una sola
	//    transacción. Si la creación falla, el rollback mantiene válido el token
	//    viejo — así un error transitorio de BD no deja al usuario sin sesión ni
	//    dispara un falso positivo de breach detection en el siguiente intento.
	newTokenHash := h.hashToken(newRefreshToken)
	rotateErr := h.txManager.Do(ctx, func(txCtx context.Context) error {
		if err := h.tokenRepo.RevokeToken(txCtx, storedToken.ID); err != nil {
			return fmt.Errorf("revocando token viejo: %w", err)
		}
		if err := h.tokenRepo.CreateRefreshToken(
			txCtx,
			user.ID,
			newTokenID,
			newTokenHash,
			refreshExpiry,
			"", // IP address - se puede obtener del contexto
			"", // User agent - se puede obtener del contexto
		); err != nil {
			return fmt.Errorf("guardando nuevo refresh token: %w", err)
		}
		return nil
	})
	if rotateErr != nil {
		h.logger.Error("Error rotando refresh token (rollback aplicado)",
			zap.String("user_id", user.ID.String()),
			zap.Error(rotateErr),
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
