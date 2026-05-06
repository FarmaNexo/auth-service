// internal/application/handlers/register_user_handler.go
package handlers

import (
	"context"
	"fmt"

	"github.com/farmanexo/auth-service/internal/application/commands"
	"github.com/farmanexo/auth-service/internal/domain/entities"
	"github.com/farmanexo/auth-service/internal/domain/events"
	"github.com/farmanexo/auth-service/internal/domain/repositories"
	"github.com/farmanexo/auth-service/internal/domain/services"
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
	"github.com/farmanexo/auth-service/internal/shared/common"
	"github.com/farmanexo/auth-service/internal/shared/constants"
	"github.com/farmanexo/auth-service/pkg/mediator"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const defaultConsentLocale = "es-PE"

// RegisterUserHandler maneja el comando RegisterUserCommand
type RegisterUserHandler struct {
	userRepo       repositories.UserRepository
	consentRepo    repositories.ConsentRepository
	legalDocRepo   repositories.LegalDocumentRepository
	eventPublisher services.EventPublisher
	logger         *zap.Logger
}

func NewRegisterUserHandler(
	userRepo repositories.UserRepository,
	consentRepo repositories.ConsentRepository,
	legalDocRepo repositories.LegalDocumentRepository,
	eventPublisher services.EventPublisher,
	logger *zap.Logger,
) *RegisterUserHandler {
	return &RegisterUserHandler{
		userRepo:       userRepo,
		consentRepo:    consentRepo,
		legalDocRepo:   legalDocRepo,
		eventPublisher: eventPublisher,
		logger:         logger,
	}
}

func (h *RegisterUserHandler) Handle(
	ctx context.Context,
	command commands.RegisterUserCommand,
) (*common.ApiResponse[responses.RegisterResponse], error) {

	h.logger.Info("Procesando registro de usuario",
		zap.String("email", command.Email),
		zap.String("request_name", command.GetName()),
	)

	// 1. Verificar si el email ya existe
	exists, err := h.userRepo.ExistsByEmail(ctx, command.Email)
	if err != nil {
		h.logger.Error("Error verificando email existente",
			zap.Error(err),
			zap.String("email", command.Email),
		)
		return common.InternalServerErrorResponse[responses.RegisterResponse](
			"Error verificando disponibilidad del email",
		), nil
	}

	if exists {
		h.logger.Warn("Intento de registro con email existente",
			zap.String("email", command.Email),
		)
		return common.ConflictResponse[responses.RegisterResponse](
			constants.CodeEmailAlreadyTaken,
			"El email ya está registrado",
		), nil
	}

	// 2. Cargar versiones vigentes de los documentos legales DESDE LA BASE DE DATOS
	//    (DB es la única fuente de verdad — el YAML ya no se usa para esto).
	termsDoc, err := h.legalDocRepo.FindCurrentPublished(ctx, entities.LegalDocCodeTerms, defaultConsentLocale)
	if err != nil || termsDoc == nil {
		h.logger.Error("Error cargando Términos vigentes",
			zap.Error(err),
		)
		return common.InternalServerErrorResponse[responses.RegisterResponse](
			"Error cargando documentos legales vigentes",
		), nil
	}
	privacyDoc, err := h.legalDocRepo.FindCurrentPublished(ctx, entities.LegalDocCodePrivacy, defaultConsentLocale)
	if err != nil || privacyDoc == nil {
		h.logger.Error("Error cargando Política de Privacidad vigente",
			zap.Error(err),
		)
		return common.InternalServerErrorResponse[responses.RegisterResponse](
			"Error cargando documentos legales vigentes",
		), nil
	}
	marketingDoc, err := h.legalDocRepo.FindCurrentPublished(ctx, entities.LegalDocCodeMarketing, defaultConsentLocale)
	if err != nil || marketingDoc == nil {
		h.logger.Error("Error cargando documento de Marketing vigente",
			zap.Error(err),
		)
		return common.InternalServerErrorResponse[responses.RegisterResponse](
			"Error cargando documentos legales vigentes",
		), nil
	}

	// 3. Hash del password
	passwordHash, err := h.hashPassword(command.Password)
	if err != nil {
		h.logger.Error("Error hasheando password", zap.Error(err))
		return common.InternalServerErrorResponse[responses.RegisterResponse](
			"Error procesando la contraseña",
		), nil
	}

	// 4. Crear entidad User
	user := entities.NewUserWithPhone(
		command.Email,
		passwordHash,
		command.FullName,
		command.Phone,
	)

	// 5. Guardar usuario
	if err := h.userRepo.Create(ctx, user); err != nil {
		h.logger.Error("Error creando usuario en BD",
			zap.Error(err),
			zap.String("email", command.Email),
		)
		return common.InternalServerErrorResponse[responses.RegisterResponse](
			"Error creando el usuario",
		), nil
	}

	h.logger.Info("Usuario creado exitosamente",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
	)

	// 6. Persistir consentimientos LPDP ligados al document_id exacto + hash de integridad.
	//    La validación ya garantizó que AcceptedTerms == true y AcceptedPrivacy == true.
	consents := []*entities.UserConsent{
		entities.NewUserConsentWithDocument(
			user.ID, entities.ConsentTypeTermsOfService, termsDoc.Version,
			termsDoc.ID, termsDoc.ContentHash, true,
			command.IPAddress, command.UserAgent,
		),
		entities.NewUserConsentWithDocument(
			user.ID, entities.ConsentTypePrivacyPolicy, privacyDoc.Version,
			privacyDoc.ID, privacyDoc.ContentHash, true,
			command.IPAddress, command.UserAgent,
		),
		entities.NewUserConsentWithDocument(
			user.ID, entities.ConsentTypeMarketingCommunications, marketingDoc.Version,
			marketingDoc.ID, marketingDoc.ContentHash, command.MarketingOptIn,
			command.IPAddress, command.UserAgent,
		),
	}
	if err := h.consentRepo.CreateBatch(ctx, consents); err != nil {
		h.logger.Error("Error persistiendo consentimientos (usuario quedó creado)",
			zap.Error(err),
			zap.String("user_id", user.ID.String()),
		)
		return common.InternalServerErrorResponse[responses.RegisterResponse](
			"Error registrando los consentimientos legales",
		), nil
	}

	// Publicar evento de registro (fire-and-forget)
	go h.publishEvent(context.Background(), events.NewUserRegisteredEvent(
		user.ID.String(),
		user.Email,
		user.FullName,
		user.Phone,
		user.Role,
	))

	registerResponse := responses.NewRegisterResponse(
		user.ID,
		user.Email,
		user.CreatedAt,
	)

	return common.CreatedResponse(*registerResponse), nil
}

// publishEvent publica un evento de autenticación (best-effort)
func (h *RegisterUserHandler) publishEvent(ctx context.Context, event events.AuthEvent) {
	if err := h.eventPublisher.Publish(ctx, event); err != nil {
		h.logger.Warn("Error publicando evento de autenticación",
			zap.String("event_type", event.EventType),
			zap.String("user_id", event.UserID),
			zap.Error(err),
		)
	}
}

func (h *RegisterUserHandler) hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("error hashing password: %w", err)
	}
	return string(hash), nil
}

var _ mediator.RequestHandler[commands.RegisterUserCommand, responses.RegisterResponse] = (*RegisterUserHandler)(nil)
