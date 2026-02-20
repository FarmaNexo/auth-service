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

// RegisterUserHandler maneja el comando RegisterUserCommand
type RegisterUserHandler struct {
	userRepo       repositories.UserRepository
	eventPublisher services.EventPublisher
	logger         *zap.Logger
}

func NewRegisterUserHandler(
	userRepo repositories.UserRepository,
	eventPublisher services.EventPublisher,
	logger *zap.Logger,
) *RegisterUserHandler {
	return &RegisterUserHandler{
		userRepo:       userRepo,
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

	// 2. Hash del password
	passwordHash, err := h.hashPassword(command.Password)
	if err != nil {
		h.logger.Error("Error hasheando password", zap.Error(err))
		return common.InternalServerErrorResponse[responses.RegisterResponse](
			"Error procesando la contraseña",
		), nil
	}

	// 3. Crear entidad User
	user := entities.NewUserWithPhone(
		command.Email,
		passwordHash,
		command.FullName,
		command.Phone,
	)

	// 4. Guardar en base de datos
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

	// Publicar evento de registro (fire-and-forget)
	go h.publishEvent(context.Background(), events.NewUserRegisteredEvent(user.ID.String(), user.Email))

	// 5. Construir respuesta SIMPLIFICADA (sin tokens)
	registerResponse := responses.NewRegisterResponse(
		user.ID,
		user.Email,
		user.CreatedAt,
	)

	// 6. Retornar respuesta exitosa
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
