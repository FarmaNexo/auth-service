// internal/application/handlers/get_profile_handler.go
package handlers

import (
	"context"

	"github.com/farmanexo/auth-service/internal/application/queries"
	"github.com/farmanexo/auth-service/internal/domain/repositories"
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
	"github.com/farmanexo/auth-service/internal/shared/common"
	"github.com/farmanexo/auth-service/internal/shared/constants"
	"github.com/farmanexo/auth-service/pkg/mediator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// GetProfileHandler maneja la query GetProfileQuery
type GetProfileHandler struct {
	userRepo repositories.UserRepository
	logger   *zap.Logger
}

// NewGetProfileHandler crea una nueva instancia del handler
func NewGetProfileHandler(
	userRepo repositories.UserRepository,
	logger *zap.Logger,
) *GetProfileHandler {
	return &GetProfileHandler{
		userRepo: userRepo,
		logger:   logger,
	}
}

// Handle procesa la query de obtener perfil
func (h *GetProfileHandler) Handle(
	ctx context.Context,
	query queries.GetProfileQuery,
) (*common.ApiResponse[responses.UserResponse], error) {

	h.logger.Info("Procesando obtener perfil de usuario",
		zap.String("user_id", query.UserID),
		zap.String("request_name", query.GetName()),
	)

	// 1. Parsear user ID
	userUUID, err := uuid.Parse(query.UserID)
	if err != nil {
		h.logger.Warn("User ID inválido",
			zap.String("user_id", query.UserID),
			zap.Error(err),
		)
		response := common.NewApiResponse[responses.UserResponse]()
		response.SetHttpStatus(constants.StatusBadRequest.Int())
		response.AddError(constants.CodeValidationError, "ID de usuario inválido")
		return response, nil
	}

	// 2. Buscar usuario en BD
	user, err := h.userRepo.FindByID(ctx, userUUID)
	if err != nil {
		h.logger.Warn("Usuario no encontrado",
			zap.String("user_id", query.UserID),
			zap.Error(err),
		)
		return common.NotFoundResponse[responses.UserResponse]("Usuario no encontrado"), nil
	}

	// 3. Verificar que el usuario esté activo
	if !user.IsActive {
		h.logger.Warn("Perfil solicitado de usuario inactivo",
			zap.String("user_id", user.ID.String()),
		)
		response := common.NewApiResponse[responses.UserResponse]()
		response.SetHttpStatus(constants.StatusForbidden.Int())
		response.AddError(constants.CodeUserInactive, "Usuario inactivo")
		return response, nil
	}

	h.logger.Info("Perfil obtenido exitosamente",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
	)

	// 4. Construir UserResponse
	userResponse := responses.UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		FullName:   user.FullName,
		Phone:      user.Phone,
		Role:       user.Role,
		IsVerified: user.IsVerified,
		CreatedAt:  user.CreatedAt,
	}

	return common.OkResponse(userResponse), nil
}

// Asegurar que implementa la interfaz RequestHandler
var _ mediator.RequestHandler[queries.GetProfileQuery, responses.UserResponse] = (*GetProfileHandler)(nil)
