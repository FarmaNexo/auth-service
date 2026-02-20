// internal/presentation/http/controllers/auth_controller.go
package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/farmanexo/auth-service/internal/application/commands"
	"github.com/farmanexo/auth-service/internal/application/queries"
	"github.com/farmanexo/auth-service/internal/presentation/dto/requests"
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
	"github.com/farmanexo/auth-service/internal/presentation/http/middlewares"
	"github.com/farmanexo/auth-service/internal/shared/common"
	"github.com/farmanexo/auth-service/internal/shared/constants"
	"github.com/farmanexo/auth-service/pkg/mediator"
	"go.uber.org/zap"
)

type AuthController struct {
	mediator *mediator.Mediator
	logger   *zap.Logger
}

func NewAuthController(mediator *mediator.Mediator, logger *zap.Logger) *AuthController {
	return &AuthController{
		mediator: mediator,
		logger:   logger,
	}
}

// ========================================
// HTTP HANDLERS
// ========================================

// Register godoc
// @Summary      Registrar nuevo usuario
// @Description  Registra un nuevo usuario en el sistema. Después de registrarse, debe iniciar sesión.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      requests.RegisterRequest  true  "Datos de registro"
// @Success      201      {object}  common.ApiResponse[responses.RegisterResponse]  "Usuario registrado exitosamente"
// @Failure      400      {object}  common.ApiResponse[responses.RegisterResponse]  "Error de validación"
// @Failure      409      {object}  common.ApiResponse[responses.RegisterResponse]  "Email ya registrado"
// @Failure      500      {object}  common.ApiResponse[responses.RegisterResponse]  "Error interno del servidor"
// @Router       /api/v1/auth/register [post]
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("POST /api/v1/auth/register - Iniciando registro de usuario")

	var req requests.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.logger.Warn("Error decodificando request body", zap.Error(err))
		c.respondJSON(w, common.BadRequestResponse[responses.RegisterResponse](
			constants.CodeValidationError,
			"Invalid request body",
		))
		return
	}

	command := commands.RegisterUserCommand{
		Email:    req.Email,
		Password: req.Password,
		FullName: req.FullName,
		Phone:    req.Phone,
	}

	response, err := mediator.Send[commands.RegisterUserCommand, responses.RegisterResponse](
		r.Context(),
		c.mediator,
		command,
	)

	if err != nil {
		c.logger.Error("Error ejecutando RegisterUserCommand",
			zap.Error(err),
			zap.String("email", req.Email),
		)
		c.respondJSON(w, common.InternalServerErrorResponse[responses.RegisterResponse](
			"Error procesando el registro",
		))
		return
	}

	c.respondJSON(w, response)
}

// Login godoc
// @Summary      Iniciar sesión
// @Description  Autentica un usuario y retorna tokens de acceso SIN datos de usuario
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      requests.LoginRequest  true  "Credenciales de acceso"
// @Success      200      {object}  common.ApiResponse[responses.LoginResponse]  "Login exitoso"
// @Failure      400      {object}  common.ApiResponse[responses.LoginResponse]  "Error de validación"
// @Failure      401      {object}  common.ApiResponse[responses.LoginResponse]  "Credenciales inválidas"
// @Failure      500      {object}  common.ApiResponse[responses.LoginResponse]  "Error interno del servidor"
// @Router       /api/v1/auth/login [post]
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("POST /api/v1/auth/login - Iniciando login")

	var req requests.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.logger.Warn("Error decodificando request body", zap.Error(err))
		c.respondJSON(w, common.BadRequestResponse[responses.LoginResponse](
			constants.CodeValidationError,
			"Invalid request body",
		))
		return
	}

	command := commands.LoginCommand{
		Email:    req.Email,
		Password: req.Password,
	}

	response, err := mediator.Send[commands.LoginCommand, responses.LoginResponse](
		r.Context(),
		c.mediator,
		command,
	)

	if err != nil {
		c.logger.Error("Error ejecutando LoginCommand",
			zap.Error(err),
			zap.String("email", req.Email),
		)
		c.respondJSON(w, common.InternalServerErrorResponse[responses.LoginResponse](
			"Error procesando el login",
		))
		return
	}

	c.respondJSON(w, response)
}

// RefreshToken godoc
// @Summary      Refrescar token de acceso
// @Description  Genera nuevos access y refresh tokens usando un refresh token válido. El refresh token anterior es revocado (rotación de tokens).
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      requests.RefreshTokenRequest  true  "Refresh token"
// @Success      200      {object}  common.ApiResponse[responses.LoginResponse]  "Tokens renovados exitosamente"
// @Failure      400      {object}  common.ApiResponse[responses.LoginResponse]  "Error de validación"
// @Failure      401      {object}  common.ApiResponse[responses.LoginResponse]  "Token inválido, expirado o revocado"
// @Failure      500      {object}  common.ApiResponse[responses.LoginResponse]  "Error interno del servidor"
// @Router       /api/v1/auth/refresh [post]
func (c *AuthController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("POST /api/v1/auth/refresh - Refrescando token")

	var req requests.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.logger.Warn("Error decodificando request body", zap.Error(err))
		c.respondJSON(w, common.BadRequestResponse[responses.LoginResponse](
			constants.CodeValidationError,
			"Invalid request body",
		))
		return
	}

	if req.RefreshToken == "" {
		c.respondJSON(w, common.BadRequestResponse[responses.LoginResponse](
			constants.CodeRequiredField,
			"El campo refresh_token es requerido",
		))
		return
	}

	command := commands.RefreshTokenCommand{
		RefreshToken: req.RefreshToken,
	}

	response, err := mediator.Send[commands.RefreshTokenCommand, responses.LoginResponse](
		r.Context(),
		c.mediator,
		command,
	)

	if err != nil {
		c.logger.Error("Error ejecutando RefreshTokenCommand", zap.Error(err))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.LoginResponse](
			"Error procesando la renovación de token",
		))
		return
	}

	c.respondJSON(w, response)
}

// GetProfile godoc
// @Summary      Obtener perfil de usuario
// @Description  Retorna la información del usuario autenticado. Requiere token de acceso válido.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  common.ApiResponse[responses.UserResponse]  "Perfil obtenido exitosamente"
// @Failure      401  {object}  common.ApiResponse[responses.UserResponse]  "No autorizado"
// @Failure      404  {object}  common.ApiResponse[responses.UserResponse]  "Usuario no encontrado"
// @Failure      500  {object}  common.ApiResponse[responses.UserResponse]  "Error interno del servidor"
// @Router       /api/v1/auth/me [get]
func (c *AuthController) GetProfile(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("GET /api/v1/auth/me - Obteniendo perfil de usuario")

	// Extraer user_id del contexto (inyectado por AuthMiddleware)
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.logger.Warn("User ID no encontrado en contexto")
		c.respondJSON(w, common.UnauthorizedResponse[responses.UserResponse]("Usuario no autenticado"))
		return
	}

	query := queries.GetProfileQuery{
		UserID: userID,
	}

	response, err := mediator.Send[queries.GetProfileQuery, responses.UserResponse](
		r.Context(),
		c.mediator,
		query,
	)

	if err != nil {
		c.logger.Error("Error ejecutando GetProfileQuery",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		c.respondJSON(w, common.InternalServerErrorResponse[responses.UserResponse](
			"Error obteniendo perfil de usuario",
		))
		return
	}

	c.respondJSON(w, response)
}

// Logout godoc
// @Summary      Cerrar sesión
// @Description  Revoca el refresh token del usuario autenticado, cerrando la sesión actual. Requiere token de acceso válido.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      requests.LogoutRequest  true  "Refresh token a revocar"
// @Success      200      {object}  common.ApiResponse[responses.EmptyResponse]  "Sesión cerrada exitosamente"
// @Failure      400      {object}  common.ApiResponse[responses.EmptyResponse]  "Error de validación"
// @Failure      401      {object}  common.ApiResponse[responses.EmptyResponse]  "No autorizado"
// @Failure      403      {object}  common.ApiResponse[responses.EmptyResponse]  "Token no pertenece al usuario"
// @Failure      500      {object}  common.ApiResponse[responses.EmptyResponse]  "Error interno del servidor"
// @Router       /api/v1/auth/logout [post]
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	c.logger.Info("POST /api/v1/auth/logout - Cerrando sesión")

	// 1. Extraer user_id del contexto (inyectado por AuthMiddleware)
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.logger.Warn("User ID no encontrado en contexto para logout")
		c.respondJSON(w, common.UnauthorizedResponse[responses.EmptyResponse]("Usuario no autenticado"))
		return
	}

	// 2. Decodificar request body
	var req requests.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.logger.Warn("Error decodificando request body", zap.Error(err))
		c.respondJSON(w, common.BadRequestResponse[responses.EmptyResponse](
			constants.CodeValidationError,
			"Invalid request body",
		))
		return
	}

	if req.RefreshToken == "" {
		c.respondJSON(w, common.BadRequestResponse[responses.EmptyResponse](
			constants.CodeRequiredField,
			"El campo refresh_token es requerido",
		))
		return
	}

	// 3. Crear comando
	command := commands.LogoutCommand{
		UserID:       userID,
		RefreshToken: req.RefreshToken,
	}

	// 4. Enviar al mediator
	response, err := mediator.Send[commands.LogoutCommand, responses.EmptyResponse](
		r.Context(),
		c.mediator,
		command,
	)

	if err != nil {
		c.logger.Error("Error ejecutando LogoutCommand",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		c.respondJSON(w, common.InternalServerErrorResponse[responses.EmptyResponse](
			"Error cerrando sesión",
		))
		return
	}

	c.respondJSON(w, response)
}

// ========================================
// RESPONSE HELPERS (MEJORADOS)
// ========================================

func (c *AuthController) respondJSON(w http.ResponseWriter, response interface{}) {
	statusCode := http.StatusOK // Default

	// Usar interface para extraer GetHttpStatus() de cualquier ApiResponse
	if resp, ok := response.(interface{ GetHttpStatus() *int }); ok {
		if httpStatus := resp.GetHttpStatus(); httpStatus != nil {
			statusCode = *httpStatus
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		c.logger.Error("Error codificando respuesta JSON", zap.Error(err))
	}
}

// ========================================
// HEALTH CHECK
// ========================================

// HealthCheck godoc
// @Summary      Health check
// @Description  Verifica el estado del servicio
// @Tags         Health
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "Servicio saludable"
// @Router       /health [get]
func (c *AuthController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	type HealthResponse struct {
		Status  string `json:"status" example:"healthy"`
		Service string `json:"service" example:"auth-service"`
		Version string `json:"version" example:"1.0.0"`
	}

	health := HealthResponse{
		Status:  "healthy",
		Service: "auth-service",
		Version: "1.0.0",
	}

	response := common.OkResponse(health)
	c.respondJSON(w, response)
}
