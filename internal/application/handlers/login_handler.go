// internal/application/handlers/login_handler.go
package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/farmanexo/auth-service/internal/application/commands"
	"github.com/farmanexo/auth-service/internal/domain/repositories"
	"github.com/farmanexo/auth-service/internal/infrastructure/security"
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
	"github.com/farmanexo/auth-service/internal/shared/common"
	"github.com/farmanexo/auth-service/internal/shared/constants"
	"github.com/farmanexo/auth-service/pkg/mediator"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// LoginHandler maneja el comando LoginCommand
type LoginHandler struct {
	userRepo   repositories.UserRepository
	tokenRepo  repositories.TokenRepository
	jwtService security.JWTService
	logger     *zap.Logger
}

// NewLoginHandler crea una nueva instancia del handler
func NewLoginHandler(
	userRepo repositories.UserRepository,
	tokenRepo repositories.TokenRepository,
	jwtService security.JWTService,
	logger *zap.Logger,
) *LoginHandler {
	return &LoginHandler{
		userRepo:   userRepo,
		tokenRepo:  tokenRepo,
		jwtService: jwtService,
		logger:     logger,
	}
}

// Handle procesa el comando de login
func (h *LoginHandler) Handle(
	ctx context.Context,
	command commands.LoginCommand,
) (*common.ApiResponse[responses.LoginResponse], error) { // ← CAMBIO AQUÍ

	h.logger.Info("Procesando login de usuario",
		zap.String("email", command.Email),
		zap.String("request_name", command.GetName()),
	)

	// 1. Buscar usuario por email
	user, err := h.userRepo.FindByEmail(ctx, command.Email)
	if err != nil {
		h.logger.Warn("Usuario no encontrado",
			zap.String("email", command.Email),
			zap.Error(err),
		)
		return h.unauthorizedResponse("Credenciales inválidas"), nil
	}

	// 2. Verificar que el usuario esté activo
	if !user.IsActive {
		h.logger.Warn("Intento de login con usuario inactivo",
			zap.String("user_id", user.ID.String()),
			zap.String("email", user.Email),
		)
		return h.unauthorizedResponse("Usuario inactivo"), nil
	}

	// 3. Verificar password
	if err := h.verifyPassword(user.PasswordHash, command.Password); err != nil {
		h.logger.Warn("Password incorrecto",
			zap.String("email", command.Email),
		)
		return h.unauthorizedResponse("Credenciales inválidas"), nil
	}

	// 4. Generar tokens JWT
	accessToken, accessExpiry, err := h.jwtService.GenerateAccessToken(
		user.ID.String(),
		user.Email,
		user.FullName,
		user.Role,
	)
	if err != nil {
		h.logger.Error("Error generando access token",
			zap.Error(err),
			zap.String("user_id", user.ID.String()),
		)
		return common.InternalServerErrorResponse[responses.LoginResponse](
			"Error generando token de acceso",
		), nil
	}

	refreshToken, refreshExpiry, tokenID, err := h.jwtService.GenerateRefreshToken(user.ID.String())
	if err != nil {
		h.logger.Error("Error generando refresh token",
			zap.Error(err),
			zap.String("user_id", user.ID.String()),
		)
		return common.InternalServerErrorResponse[responses.LoginResponse](
			"Error generando refresh token",
		), nil
	}

	// 5. Hash del refresh token para almacenar
	tokenHash := h.hashToken(refreshToken)

	// 6. Guardar refresh token en base de datos
	if err := h.tokenRepo.CreateRefreshToken(
		ctx,
		user.ID,
		tokenID,
		tokenHash,
		refreshExpiry,
		"", // IP address - se puede obtener del contexto
		"", // User agent - se puede obtener del contexto
	); err != nil {
		h.logger.Error("Error guardando refresh token",
			zap.Error(err),
			zap.String("user_id", user.ID.String()),
		)
		return common.InternalServerErrorResponse[responses.LoginResponse](
			"Error procesando autenticación",
		), nil
	}

	// 7. Actualizar último login y contador
	now := time.Now()
	user.LastLoginAt = &now
	user.LoginCount++
	if err := h.userRepo.Update(ctx, user); err != nil {
		h.logger.Warn("Error actualizando último login",
			zap.Error(err),
			zap.String("user_id", user.ID.String()),
		)
		// No retornamos error, el login fue exitoso
	}

	h.logger.Info("Login exitoso",
		zap.String("user_id", user.ID.String()),
		zap.String("email", user.Email),
		zap.Int("login_count", user.LoginCount),
	)

	// 8. Construir respuesta SIMPLIFICADA (SIN datos de usuario)
	expiresIn := int64(time.Until(accessExpiry).Seconds())

	loginResponse := responses.NewLoginResponse(
		accessToken,
		refreshToken,
		expiresIn,
	)

	return common.OkResponse(*loginResponse), nil
}

// verifyPassword verifica que el password coincida con el hash
func (h *LoginHandler) verifyPassword(hashedPassword string, plainPassword string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(plainPassword),
	)
}

// hashToken genera un SHA-256 hash del token
func (h *LoginHandler) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// unauthorizedResponse crea una respuesta 401 Unauthorized
func (h *LoginHandler) unauthorizedResponse(message string) *common.ApiResponse[responses.LoginResponse] {
	response := common.NewApiResponse[responses.LoginResponse]()
	response.SetHttpStatus(constants.StatusUnauthorized.Int())
	response.AddError(constants.CodeUnauthorized, message)
	return response
}

// Asegurar que implementa la interfaz RequestHandler
var _ mediator.RequestHandler[commands.LoginCommand, responses.LoginResponse] = (*LoginHandler)(nil)
