// internal/application/postprocessors/log_audit_postprocessor.go
package postprocessors

import (
	"context"
	"time"

	"github.com/farmanexo/auth-service/internal/application/commands"
	"github.com/farmanexo/auth-service/pkg/mediator"
	"go.uber.org/zap"
)

// LogAuditPostProcessor registra eventos de auditoría
type LogAuditPostProcessor struct {
	logger *zap.Logger
}

// NewLogAuditPostProcessor crea un nuevo postprocessor
func NewLogAuditPostProcessor(logger *zap.Logger) *LogAuditPostProcessor {
	return &LogAuditPostProcessor{
		logger: logger,
	}
}

// Process ejecuta el logging de auditoría
func (p *LogAuditPostProcessor) Process(
	ctx context.Context,
	request interface{},
	response interface{},
) error {
	userID := p.getUserIDFromContext(ctx)
	correlationID := mediator.GetCorrelationID(ctx)

	// Determinar si la respuesta fue exitosa
	isSuccess := p.checkSuccess(response)

	switch cmd := request.(type) {
	case commands.RegisterUserCommand:
		p.logAudit("USER_REGISTERED", cmd.Email, userID, correlationID, isSuccess)
	case *commands.RegisterUserCommand:
		p.logAudit("USER_REGISTERED", cmd.Email, userID, correlationID, isSuccess)

	case commands.LoginCommand:
		p.logAudit("USER_LOGIN", cmd.Email, userID, correlationID, isSuccess)
	case *commands.LoginCommand:
		p.logAudit("USER_LOGIN", cmd.Email, userID, correlationID, isSuccess)

	case commands.RefreshTokenCommand:
		p.logAudit("TOKEN_REFRESHED", "", userID, correlationID, isSuccess)
	case *commands.RefreshTokenCommand:
		p.logAudit("TOKEN_REFRESHED", "", userID, correlationID, isSuccess)

	case commands.LogoutCommand:
		p.logAudit("USER_LOGOUT", "", cmd.UserID, correlationID, isSuccess)
	case *commands.LogoutCommand:
		p.logAudit("USER_LOGOUT", "", cmd.UserID, correlationID, isSuccess)

	default:
		p.logger.Debug("Post-processor: comando sin auditoría configurada")
	}

	return nil
}

// logAudit escribe el log de auditoría estructurado
func (p *LogAuditPostProcessor) logAudit(
	eventType, email, userID, correlationID string,
	success bool,
) {
	fields := []zap.Field{
		zap.String("event_type", eventType),
		zap.Bool("success", success),
		zap.String("correlation_id", correlationID),
		zap.String("user_id", userID),
		zap.Time("timestamp", time.Now()),
	}

	if email != "" {
		fields = append(fields, zap.String("email", email))
	}

	p.logger.Info("AUDIT", fields...)
}

// checkSuccess verifica si la respuesta fue exitosa usando la interfaz IsValid
func (p *LogAuditPostProcessor) checkSuccess(response interface{}) bool {
	if resp, ok := response.(interface{ IsValid() bool }); ok {
		return resp.IsValid()
	}
	return false
}

// getUserIDFromContext extrae el user ID del contexto
func (p *LogAuditPostProcessor) getUserIDFromContext(ctx context.Context) string {
	userID, _ := mediator.GetUserID(ctx)
	if userID == "" {
		return "ANONYMOUS"
	}
	return userID
}
