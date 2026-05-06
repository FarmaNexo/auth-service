// internal/presentation/http/controllers/auth_controller.go
package controllers

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"

	"github.com/farmanexo/auth-service/internal/application/commands"
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

	req.Sanitize()

	command := commands.RegisterUserCommand{
		Email:           req.Email,
		Password:        req.Password,
		FullName:        req.FullName,
		Phone:           req.Phone,
		AcceptedTerms:   req.AcceptedTerms,
		AcceptedPrivacy: req.AcceptedPrivacy,
		MarketingOptIn:  req.MarketingOptIn,
		IPAddress:       extractClientIP(r),
		UserAgent:       r.UserAgent(),
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
// @Failure      429      {object}  common.ApiResponse[responses.LoginResponse]  "Demasiados intentos"
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

	req.Sanitize()

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
// @Failure      429      {object}  common.ApiResponse[responses.LoginResponse]  "Demasiadas solicitudes"
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

	// 3. Extraer access token del contexto (inyectado por AuthMiddleware)
	accessToken, _ := middlewares.GetAccessTokenFromContext(r.Context())

	// 4. Crear comando
	command := commands.LogoutCommand{
		UserID:       userID,
		RefreshToken: req.RefreshToken,
		AccessToken:  accessToken,
	}

	// 5. Enviar al mediator
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

// GetMyConsents godoc
// @Summary      Listar mis consentimientos
// @Description  Devuelve el historial de consentimientos (terms, privacy, marketing) del usuario autenticado. Parte del derecho ARCO de acceso (LPDP Ley 29733).
// @Tags         Privacy
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  common.ApiResponse[responses.ConsentsListResponse]  "Historial de consentimientos"
// @Failure      401  {object}  common.ApiResponse[responses.ConsentsListResponse]  "No autorizado"
// @Failure      500  {object}  common.ApiResponse[responses.ConsentsListResponse]  "Error interno"
// @Router       /api/v1/auth/me/consents [get]
func (c *AuthController) GetMyConsents(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.respondJSON(w, common.UnauthorizedResponse[responses.ConsentsListResponse]("Usuario no autenticado"))
		return
	}

	command := commands.GetMyConsentsCommand{UserID: userID}
	response, err := mediator.Send[commands.GetMyConsentsCommand, responses.ConsentsListResponse](
		r.Context(), c.mediator, command,
	)
	if err != nil {
		c.logger.Error("Error listando consentimientos", zap.Error(err), zap.String("user_id", userID))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.ConsentsListResponse]("Error consultando consentimientos"))
		return
	}
	c.respondJSON(w, response)
}

// DeleteMyAccount godoc
// @Summary      Eliminar mi cuenta (ARCO)
// @Description  Ejercicio del derecho ARCO de cancelación (LPDP Ley 29733). Anonimiza PII y soft-deletea la cuenta. Publica USER_DELETED para que otros servicios limpien sus proyecciones. Irreversible.
// @Tags         Privacy
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  common.ApiResponse[responses.EmptyResponse]  "Cuenta eliminada"
// @Failure      401  {object}  common.ApiResponse[responses.EmptyResponse]  "No autorizado"
// @Failure      404  {object}  common.ApiResponse[responses.EmptyResponse]  "Usuario no encontrado"
// @Failure      500  {object}  common.ApiResponse[responses.EmptyResponse]  "Error interno"
// @Router       /api/v1/auth/me [delete]
func (c *AuthController) DeleteMyAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.respondJSON(w, common.UnauthorizedResponse[responses.EmptyResponse]("Usuario no autenticado"))
		return
	}

	command := commands.DeleteAccountCommand{UserID: userID}
	response, err := mediator.Send[commands.DeleteAccountCommand, responses.EmptyResponse](
		r.Context(), c.mediator, command,
	)
	if err != nil {
		c.logger.Error("Error eliminando cuenta (ARCO)", zap.Error(err), zap.String("user_id", userID))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.EmptyResponse]("Error eliminando la cuenta"))
		return
	}
	c.respondJSON(w, response)
}

// GetConsentsStatus godoc
// @Summary      Estado de consentimientos (versionado)
// @Description  Indica si el usuario tiene al día los consentimientos vigentes. Si `up_to_date=false`, devuelve en `pending` los tipos que requieren re-aceptación (HU-011).
// @Tags         Privacy
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  common.ApiResponse[responses.ConsentsStatusResponse]  "Estado"
// @Failure      401  {object}  common.ApiResponse[responses.ConsentsStatusResponse]  "No autorizado"
// @Router       /api/v1/auth/me/consents/status [get]
func (c *AuthController) GetConsentsStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.respondJSON(w, common.UnauthorizedResponse[responses.ConsentsStatusResponse]("Usuario no autenticado"))
		return
	}

	command := commands.GetConsentsStatusCommand{UserID: userID}
	response, err := mediator.Send[commands.GetConsentsStatusCommand, responses.ConsentsStatusResponse](
		r.Context(), c.mediator, command,
	)
	if err != nil {
		c.logger.Error("Error consultando estado de consentimientos", zap.Error(err), zap.String("user_id", userID))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.ConsentsStatusResponse]("Error consultando estado"))
		return
	}
	c.respondJSON(w, response)
}

// AcceptConsents godoc
// @Summary      Re-aceptar consentimientos pendientes
// @Description  Registra aceptación explícita de los consentimientos indicados con la versión vigente. Usado tras bump de TyC o Política de Privacidad.
// @Tags         Privacy
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      requests.AcceptConsentsRequest  true  "Tipos a aceptar"
// @Success      200      {object}  common.ApiResponse[responses.EmptyResponse]  "Aceptado"
// @Failure      400      {object}  common.ApiResponse[responses.EmptyResponse]  "Validación"
// @Failure      401      {object}  common.ApiResponse[responses.EmptyResponse]  "No autorizado"
// @Router       /api/v1/auth/me/consents/accept [post]
func (c *AuthController) AcceptConsents(w http.ResponseWriter, r *http.Request) {
	userID, ok := middlewares.GetUserIDFromContext(r.Context())
	if !ok {
		c.respondJSON(w, common.UnauthorizedResponse[responses.EmptyResponse]("Usuario no autenticado"))
		return
	}

	var req requests.AcceptConsentsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		c.respondJSON(w, common.BadRequestResponse[responses.EmptyResponse](
			constants.CodeValidationError, "Invalid request body",
		))
		return
	}

	command := commands.AcceptConsentsCommand{
		UserID:       userID,
		ConsentTypes: req.ConsentTypes,
		IPAddress:    extractClientIP(r),
		UserAgent:    r.UserAgent(),
	}
	response, err := mediator.Send[commands.AcceptConsentsCommand, responses.EmptyResponse](
		r.Context(), c.mediator, command,
	)
	if err != nil {
		c.logger.Error("Error aceptando consentimientos", zap.Error(err), zap.String("user_id", userID))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.EmptyResponse]("Error registrando consentimientos"))
		return
	}
	c.respondJSON(w, response)
}

// extractClientIP obtiene la IP real del cliente respetando X-Forwarded-For (via ALB/API Gateway).
// Retorna string vacío si no se puede determinar; la BD acepta NULL en ese caso.
func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return strings.TrimSpace(realIP)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ========================================
// RESPONSE HELPERS
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
