// internal/application/handlers/get_consents_status_handler.go
package handlers

import (
	"context"

	"github.com/farmanexo/auth-service/internal/application/commands"
	"github.com/farmanexo/auth-service/internal/domain/entities"
	"github.com/farmanexo/auth-service/internal/domain/repositories"
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
	"github.com/farmanexo/auth-service/internal/shared/common"
	"github.com/farmanexo/auth-service/pkg/mediator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetConsentsStatusHandler evalúa si el usuario necesita re-aceptar consentimientos (HU-011).
// Lee la versión vigente directamente desde auth.legal_documents (única fuente de verdad).
type GetConsentsStatusHandler struct {
	consentRepo  repositories.ConsentRepository
	legalDocRepo repositories.LegalDocumentRepository
	logger       *zap.Logger
}

func NewGetConsentsStatusHandler(
	consentRepo repositories.ConsentRepository,
	legalDocRepo repositories.LegalDocumentRepository,
	logger *zap.Logger,
) *GetConsentsStatusHandler {
	return &GetConsentsStatusHandler{
		consentRepo:  consentRepo,
		legalDocRepo: legalDocRepo,
		logger:       logger,
	}
}

func (h *GetConsentsStatusHandler) Handle(
	ctx context.Context,
	command commands.GetConsentsStatusCommand,
) (*common.ApiResponse[responses.ConsentsStatusResponse], error) {

	userID, err := uuid.Parse(command.UserID)
	if err != nil {
		return common.BadRequestResponse[responses.ConsentsStatusResponse](
			"VAL_001", "ID de usuario inválido",
		), nil
	}

	pending := []responses.PendingConsent{}

	// Únicos consentimientos versionados que bloquean: terms + privacy.
	// (marketing es opt-in/out libre; bumpear su versión no fuerza re-aceptación.)
	checks := []struct {
		consentType string
		legalCode   string
	}{
		{entities.ConsentTypeTermsOfService, entities.LegalDocCodeTerms},
		{entities.ConsentTypePrivacyPolicy, entities.LegalDocCodePrivacy},
	}

	for _, check := range checks {
		// Versión vigente desde DB
		current, err := h.legalDocRepo.FindCurrentPublished(ctx, check.legalCode, defaultConsentLocale)
		if err != nil || current == nil {
			h.logger.Error("Error cargando documento vigente",
				zap.Error(err),
				zap.String("legal_code", check.legalCode),
			)
			return common.InternalServerErrorResponse[responses.ConsentsStatusResponse](
				"Error consultando documentos legales vigentes",
			), nil
		}

		latest, err := h.consentRepo.FindLatestActive(ctx, userID, check.consentType)
		if err != nil {
			h.logger.Error("Error buscando último consent activo",
				zap.Error(err),
				zap.String("user_id", command.UserID),
				zap.String("consent_type", check.consentType),
			)
			return common.InternalServerErrorResponse[responses.ConsentsStatusResponse](
				"Error consultando el estado de consentimientos",
			), nil
		}

		// Si no hay activo o la versión no coincide con la vigente, queda pendiente.
		if latest == nil || !latest.Accepted || latest.DocumentVersion != current.Version {
			currentAccepted := ""
			if latest != nil {
				currentAccepted = latest.DocumentVersion
			}
			pending = append(pending, responses.PendingConsent{
				ConsentType:            check.consentType,
				CurrentAcceptedVersion: currentAccepted,
				RequiredVersion:        current.Version,
			})
		}
	}

	resp := responses.ConsentsStatusResponse{
		UpToDate: len(pending) == 0,
		Pending:  pending,
	}
	return common.OkResponse(resp), nil
}

var _ mediator.RequestHandler[commands.GetConsentsStatusCommand, responses.ConsentsStatusResponse] = (*GetConsentsStatusHandler)(nil)
