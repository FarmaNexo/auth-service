// internal/application/handlers/accept_consents_handler.go
package handlers

import (
	"context"
	"fmt"

	"github.com/farmanexo/auth-service/internal/application/commands"
	"github.com/farmanexo/auth-service/internal/domain/entities"
	"github.com/farmanexo/auth-service/internal/domain/repositories"
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
	"github.com/farmanexo/auth-service/internal/shared/common"
	"github.com/farmanexo/auth-service/pkg/mediator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AcceptConsentsHandler procesa re-aceptación de consentimientos pendientes (HU-011).
// Inserta nuevas filas en auth.user_consents apuntando al legal_document vigente.
// Las filas viejas no se tocan ni revocan: queda histórico completo.
type AcceptConsentsHandler struct {
	consentRepo  repositories.ConsentRepository
	legalDocRepo repositories.LegalDocumentRepository
	logger       *zap.Logger
}

func NewAcceptConsentsHandler(
	consentRepo repositories.ConsentRepository,
	legalDocRepo repositories.LegalDocumentRepository,
	logger *zap.Logger,
) *AcceptConsentsHandler {
	return &AcceptConsentsHandler{
		consentRepo:  consentRepo,
		legalDocRepo: legalDocRepo,
		logger:       logger,
	}
}

func (h *AcceptConsentsHandler) Handle(
	ctx context.Context,
	command commands.AcceptConsentsCommand,
) (*common.ApiResponse[responses.EmptyResponse], error) {

	userID, err := uuid.Parse(command.UserID)
	if err != nil {
		return common.BadRequestResponse[responses.EmptyResponse](
			"VAL_001", "ID de usuario inválido",
		), nil
	}

	if len(command.ConsentTypes) == 0 {
		return common.BadRequestResponse[responses.EmptyResponse](
			"VAL_002", "Debes indicar al menos un tipo de consentimiento a aceptar",
		), nil
	}

	consents := make([]*entities.UserConsent, 0, len(command.ConsentTypes))
	for _, ct := range command.ConsentTypes {
		// Solo terms y privacy son aceptables vía este endpoint (marketing tiene su propio toggle)
		if ct != entities.ConsentTypeTermsOfService && ct != entities.ConsentTypePrivacyPolicy {
			return common.BadRequestResponse[responses.EmptyResponse](
				"VAL_003",
				fmt.Sprintf("Tipo de consentimiento no aceptable vía este endpoint: %s", ct),
			), nil
		}

		legalCode := entities.LegalDocCodeFromConsentType(ct)
		doc, err := h.legalDocRepo.FindCurrentPublished(ctx, legalCode, defaultConsentLocale)
		if err != nil || doc == nil {
			h.logger.Error("Error cargando documento legal vigente",
				zap.Error(err),
				zap.String("legal_code", legalCode),
			)
			return common.InternalServerErrorResponse[responses.EmptyResponse](
				"Error cargando documentos legales vigentes",
			), nil
		}

		consents = append(consents, entities.NewUserConsentWithDocument(
			userID, ct, doc.Version, doc.ID, doc.ContentHash,
			true, command.IPAddress, command.UserAgent,
		))
	}

	if err := h.consentRepo.CreateBatch(ctx, consents); err != nil {
		h.logger.Error("Error persistiendo re-aceptación de consentimientos",
			zap.Error(err), zap.String("user_id", command.UserID),
		)
		return common.InternalServerErrorResponse[responses.EmptyResponse](
			"Error registrando los consentimientos",
		), nil
	}

	h.logger.Info("Consentimientos re-aceptados",
		zap.String("user_id", command.UserID),
		zap.Strings("consent_types", command.ConsentTypes),
	)
	return common.OkResponse(responses.EmptyResponse{}), nil
}

var _ mediator.RequestHandler[commands.AcceptConsentsCommand, responses.EmptyResponse] = (*AcceptConsentsHandler)(nil)
