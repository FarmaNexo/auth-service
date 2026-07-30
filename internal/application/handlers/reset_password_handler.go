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
	"golang.org/x/crypto/bcrypt"
)

// minPasswordLength es la longitud mínima aceptada para la nueva contraseña.
const minPasswordLength = 8

// ResetPasswordHandler maneja el restablecimiento de contraseña con un token válido.
type ResetPasswordHandler struct {
	txManager repositories.TransactionManager
	userRepo  repositories.UserRepository
	resetRepo repositories.PasswordResetTokenRepository
	tokenRepo repositories.TokenRepository
	logger    *zap.Logger
}

// NewResetPasswordHandler crea una nueva instancia del handler.
func NewResetPasswordHandler(
	txManager repositories.TransactionManager,
	userRepo repositories.UserRepository,
	resetRepo repositories.PasswordResetTokenRepository,
	tokenRepo repositories.TokenRepository,
	logger *zap.Logger,
) *ResetPasswordHandler {
	return &ResetPasswordHandler{
		txManager: txManager,
		userRepo:  userRepo,
		resetRepo: resetRepo,
		tokenRepo: tokenRepo,
		logger:    logger,
	}
}

// Handle procesa el comando de reset-password.
func (h *ResetPasswordHandler) Handle(
	ctx context.Context,
	command commands.ResetPasswordCommand,
) (*common.ApiResponse[responses.EmptyResponse], error) {

	h.logger.Info("Procesando restablecimiento de contraseña")

	if len(command.NewPassword) < minPasswordLength {
		return h.badRequest("La contraseña debe tener al menos 8 caracteres"), nil
	}

	tokenHash := h.hashToken(command.Token)
	stored, err := h.resetRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		h.logger.Warn("Token de reset no encontrado", zap.Error(err))
		return h.badRequest("El enlace de restablecimiento es inválido o expiró"), nil
	}
	if !stored.IsUsable() {
		h.logger.Warn("Token de reset ya usado o expirado",
			zap.String("token_id", stored.ID.String()),
		)
		return h.badRequest("El enlace de restablecimiento es inválido o expiró"), nil
	}

	user, err := h.userRepo.FindByID(ctx, stored.UserID)
	if err != nil {
		h.logger.Error("Usuario del token de reset no encontrado",
			zap.String("user_id", stored.UserID.String()),
			zap.Error(err),
		)
		return h.badRequest("El enlace de restablecimiento es inválido o expiró"), nil
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(command.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		h.logger.Error("Error hasheando nueva contraseña", zap.Error(err))
		return common.InternalServerErrorResponse[responses.EmptyResponse](
			"Error procesando la solicitud",
		), nil
	}

	// Todo o nada: actualiza la contraseña, marca el token como usado y revoca todas
	// las sesiones del usuario para forzar un nuevo inicio de sesión.
	txErr := h.txManager.Do(ctx, func(txCtx context.Context) error {
		user.PasswordHash = string(newHash)
		if err := h.userRepo.Update(txCtx, user); err != nil {
			return err
		}
		if err := h.resetRepo.MarkUsed(txCtx, stored.ID); err != nil {
			return err
		}
		if err := h.tokenRepo.RevokeAllUserTokens(txCtx, user.ID); err != nil {
			return err
		}
		return nil
	})
	if txErr != nil {
		h.logger.Error("Error restableciendo la contraseña (rollback aplicado)",
			zap.String("user_id", user.ID.String()),
			zap.Error(txErr),
		)
		return common.InternalServerErrorResponse[responses.EmptyResponse](
			"Error procesando la solicitud",
		), nil
	}

	h.logger.Info("Contraseña restablecida exitosamente",
		zap.String("user_id", user.ID.String()),
	)

	resp := common.OkResponse(responses.EmptyResponse{})
	resp.AddMessageWithType(
		constants.CodeUpdatedSuccess,
		"Contraseña actualizada. Inicia sesión con tu nueva contraseña.",
		constants.MessageTypeSuccess,
	)
	return resp, nil
}

func (h *ResetPasswordHandler) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (h *ResetPasswordHandler) badRequest(message string) *common.ApiResponse[responses.EmptyResponse] {
	return common.BadRequestResponse[responses.EmptyResponse](constants.CodeValidationError, message)
}

var _ mediator.RequestHandler[commands.ResetPasswordCommand, responses.EmptyResponse] = (*ResetPasswordHandler)(nil)
