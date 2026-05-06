// internal/application/handlers/get_my_consents_handler.go
package handlers

import (
	"context"

	"github.com/farmanexo/auth-service/internal/application/commands"
	"github.com/farmanexo/auth-service/internal/domain/repositories"
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
	"github.com/farmanexo/auth-service/internal/shared/common"
	"github.com/farmanexo/auth-service/pkg/mediator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetMyConsentsHandler devuelve el historial de consentimientos del usuario.
type GetMyConsentsHandler struct {
	consentRepo repositories.ConsentRepository
	logger      *zap.Logger
}

func NewGetMyConsentsHandler(
	consentRepo repositories.ConsentRepository,
	logger *zap.Logger,
) *GetMyConsentsHandler {
	return &GetMyConsentsHandler{
		consentRepo: consentRepo,
		logger:      logger,
	}
}

func (h *GetMyConsentsHandler) Handle(
	ctx context.Context,
	command commands.GetMyConsentsCommand,
) (*common.ApiResponse[responses.ConsentsListResponse], error) {

	userID, err := uuid.Parse(command.UserID)
	if err != nil {
		return common.BadRequestResponse[responses.ConsentsListResponse](
			"VAL_001", "ID de usuario inválido",
		), nil
	}

	consents, err := h.consentRepo.FindAllByUserID(ctx, userID)
	if err != nil {
		h.logger.Error("Error listando consentimientos",
			zap.Error(err), zap.String("user_id", command.UserID),
		)
		return common.InternalServerErrorResponse[responses.ConsentsListResponse](
			"Error consultando consentimientos",
		), nil
	}

	resp := responses.NewConsentsListResponse(consents)
	return common.OkResponse(*resp), nil
}

var _ mediator.RequestHandler[commands.GetMyConsentsCommand, responses.ConsentsListResponse] = (*GetMyConsentsHandler)(nil)
