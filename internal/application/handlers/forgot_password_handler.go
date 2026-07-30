package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/farmanexo/auth-service/internal/application/commands"
	"github.com/farmanexo/auth-service/internal/domain/entities"
	"github.com/farmanexo/auth-service/internal/domain/repositories"
	"github.com/farmanexo/auth-service/internal/domain/services"
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
	"github.com/farmanexo/auth-service/internal/shared/common"
	"github.com/farmanexo/auth-service/internal/shared/constants"
	"github.com/farmanexo/auth-service/pkg/mediator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// passwordResetTokenTTL es la vigencia del enlace de restablecimiento.
const passwordResetTokenTTL = time.Hour

// ForgotPasswordHandler maneja la solicitud de restablecimiento de contraseña.
type ForgotPasswordHandler struct {
	userRepo     repositories.UserRepository
	resetRepo    repositories.PasswordResetTokenRepository
	emailService services.EmailService
	logger       *zap.Logger
}

// NewForgotPasswordHandler crea una nueva instancia del handler.
func NewForgotPasswordHandler(
	userRepo repositories.UserRepository,
	resetRepo repositories.PasswordResetTokenRepository,
	emailService services.EmailService,
	logger *zap.Logger,
) *ForgotPasswordHandler {
	return &ForgotPasswordHandler{
		userRepo:     userRepo,
		resetRepo:    resetRepo,
		emailService: emailService,
		logger:       logger,
	}
}

// Handle procesa el comando de forgot-password.
func (h *ForgotPasswordHandler) Handle(
	ctx context.Context,
	command commands.ForgotPasswordCommand,
) (*common.ApiResponse[responses.EmptyResponse], error) {

	h.logger.Info("Procesando solicitud de restablecimiento de contraseña",
		zap.String("email", command.Email),
	)

	// La respuesta es idéntica exista o no el usuario, para no filtrar qué correos
	// están registrados (anti-enumeración).
	generic := h.genericResponse()

	user, err := h.userRepo.FindByEmail(ctx, command.Email)
	if err != nil {
		h.logger.Info("Solicitud de reset para email no registrado (respuesta genérica)",
			zap.String("email", command.Email),
		)
		return generic, nil
	}

	if !user.IsActive {
		h.logger.Warn("Solicitud de reset para usuario inactivo",
			zap.String("user_id", user.ID.String()),
		)
		return generic, nil
	}

	// Invalida enlaces previos del usuario antes de emitir uno nuevo.
	if err := h.resetRepo.InvalidateUserTokens(ctx, user.ID); err != nil {
		h.logger.Warn("No se pudieron invalidar tokens de reset previos",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
	}

	rawToken, err := h.generateToken()
	if err != nil {
		h.logger.Error("Error generando token de reset", zap.Error(err))
		return common.InternalServerErrorResponse[responses.EmptyResponse](
			"Error procesando la solicitud",
		), nil
	}

	resetToken := &entities.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		TokenHash: h.hashToken(rawToken),
		ExpiresAt: time.Now().Add(passwordResetTokenTTL),
		Used:      false,
		CreatedAt: time.Now(),
	}
	if err := h.resetRepo.Create(ctx, resetToken); err != nil {
		h.logger.Error("Error guardando token de reset",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
		return common.InternalServerErrorResponse[responses.EmptyResponse](
			"Error procesando la solicitud",
		), nil
	}

	// El envío del correo no debe hacer fallar la respuesta: el token ya quedó creado.
	if err := h.emailService.SendPasswordReset(ctx, user.Email, rawToken); err != nil {
		h.logger.Error("Error enviando correo de restablecimiento",
			zap.String("user_id", user.ID.String()),
			zap.Error(err),
		)
	}

	h.logger.Info("Token de restablecimiento generado",
		zap.String("user_id", user.ID.String()),
	)
	return generic, nil
}

// generateToken genera un token aleatorio criptográficamente seguro (256 bits).
func (h *ForgotPasswordHandler) generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// hashToken genera el SHA-256 del token (lo que se persiste).
func (h *ForgotPasswordHandler) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// genericResponse construye la respuesta neutra que se devuelve siempre.
func (h *ForgotPasswordHandler) genericResponse() *common.ApiResponse[responses.EmptyResponse] {
	resp := common.OkResponse(responses.EmptyResponse{})
	resp.AddMessageWithType(
		constants.CodeAuthSuccess,
		"Si el correo está registrado, enviaremos instrucciones para restablecer la contraseña.",
		constants.MessageTypeSuccess,
	)
	return resp
}

var _ mediator.RequestHandler[commands.ForgotPasswordCommand, responses.EmptyResponse] = (*ForgotPasswordHandler)(nil)
