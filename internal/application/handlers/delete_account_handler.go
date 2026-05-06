// internal/application/handlers/delete_account_handler.go
package handlers

import (
	"context"
	"errors"
	"fmt"

	"github.com/farmanexo/auth-service/internal/application/commands"
	"github.com/farmanexo/auth-service/internal/domain/events"
	"github.com/farmanexo/auth-service/internal/domain/repositories"
	"github.com/farmanexo/auth-service/internal/domain/services"
	"github.com/farmanexo/auth-service/internal/infrastructure/persistence/postgres"
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
	"github.com/farmanexo/auth-service/internal/shared/common"
	"github.com/farmanexo/auth-service/pkg/mediator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// DeleteAccountHandler maneja el comando DeleteAccountCommand (ejercicio ARCO de cancelación).
type DeleteAccountHandler struct {
	userRepo       repositories.UserRepository
	tokenRepo      repositories.TokenRepository
	eventPublisher services.EventPublisher
	logger         *zap.Logger
}

func NewDeleteAccountHandler(
	userRepo repositories.UserRepository,
	tokenRepo repositories.TokenRepository,
	eventPublisher services.EventPublisher,
	logger *zap.Logger,
) *DeleteAccountHandler {
	return &DeleteAccountHandler{
		userRepo:       userRepo,
		tokenRepo:      tokenRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

func (h *DeleteAccountHandler) Handle(
	ctx context.Context,
	command commands.DeleteAccountCommand,
) (*common.ApiResponse[responses.EmptyResponse], error) {

	h.logger.Info("Procesando eliminación de cuenta (ARCO)",
		zap.String("user_id", command.UserID),
	)

	userID, err := uuid.Parse(command.UserID)
	if err != nil {
		return common.BadRequestResponse[responses.EmptyResponse](
			"VAL_001", "ID de usuario inválido",
		), nil
	}

	// 1. Localizar usuario
	user, err := h.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, postgres.ErrUserNotFound) {
			return common.NotFoundResponse[responses.EmptyResponse](
				"Usuario no encontrado",
			), nil
		}
		h.logger.Error("Error buscando usuario para eliminación",
			zap.Error(err), zap.String("user_id", command.UserID),
		)
		return common.InternalServerErrorResponse[responses.EmptyResponse](
			"Error procesando la solicitud",
		), nil
	}

	// 2. Anonimizar PII: reemplaza email/nombre/teléfono con placeholders.
	//    El email queda con formato determinístico que no colisiona con usuarios reales.
	user.Email = fmt.Sprintf("deleted-%s@deleted.local", user.ID.String())
	user.FullName = "USUARIO ELIMINADO"
	user.Phone = ""
	user.IsActive = false

	if err := h.userRepo.Update(ctx, user); err != nil {
		h.logger.Error("Error anonimizando usuario",
			zap.Error(err), zap.String("user_id", command.UserID),
		)
		return common.InternalServerErrorResponse[responses.EmptyResponse](
			"Error anonimizando datos personales",
		), nil
	}

	// 3. Soft-delete del usuario (marca deleted_at).
	//    Los consentimientos en auth.user_consents se preservan para auditoría LPDP;
	//    no son "orphan" porque user_id sigue apuntando a una fila existente (soft-deleted).
	if err := h.userRepo.Delete(ctx, user.ID); err != nil {
		h.logger.Error("Error haciendo soft-delete del usuario",
			zap.Error(err), zap.String("user_id", command.UserID),
		)
		return common.InternalServerErrorResponse[responses.EmptyResponse](
			"Error eliminando la cuenta",
		), nil
	}

	// 4. Revocar todos los refresh tokens vigentes.
	if err := h.tokenRepo.RevokeAllUserTokens(ctx, user.ID); err != nil {
		// No bloqueante: si falla la revocación, el access token expira en 15min
		// y sin refresh token vigente no hay manera de volver a sesionar.
		h.logger.Warn("Error revocando tokens (continúa el flujo)",
			zap.Error(err), zap.String("user_id", command.UserID),
		)
	}

	// 5. Publicar evento USER_DELETED (fire-and-forget).
	//    User-service y otros consumidores deben anonimizar sus proyecciones.
	go h.publishEvent(context.Background(), events.NewUserDeletedEvent(user.ID.String()))

	h.logger.Info("Cuenta eliminada (ARCO) exitosamente",
		zap.String("user_id", user.ID.String()),
	)

	return common.OkResponse(responses.EmptyResponse{}), nil
}

func (h *DeleteAccountHandler) publishEvent(ctx context.Context, event events.AuthEvent) {
	if err := h.eventPublisher.Publish(ctx, event); err != nil {
		h.logger.Warn("Error publicando evento USER_DELETED",
			zap.String("user_id", event.UserID),
			zap.Error(err),
		)
	}
}

var _ mediator.RequestHandler[commands.DeleteAccountCommand, responses.EmptyResponse] = (*DeleteAccountHandler)(nil)
