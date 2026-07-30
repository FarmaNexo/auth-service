package email

import (
	"context"
	"fmt"

	"github.com/farmanexo/auth-service/internal/domain/services"
	"go.uber.org/zap"
)

// LogEmailService es la implementación de EmailService para desarrollo local:
// en lugar de enviar un correo real, registra el enlace de restablecimiento en el
// log para poder probar el flujo completo. En producción se reemplaza por una
// implementación sobre SES o un proveedor externo (fase 2).
type LogEmailService struct {
	frontendURL string
	logger      *zap.Logger
}

// NewLogEmailService crea el servicio de email de desarrollo.
func NewLogEmailService(frontendURL string, logger *zap.Logger) *LogEmailService {
	return &LogEmailService{frontendURL: frontendURL, logger: logger}
}

func (s *LogEmailService) SendPasswordReset(ctx context.Context, toEmail string, resetToken string) error {
	link := fmt.Sprintf("%s/restablecer-contrasena?token=%s", s.frontendURL, resetToken)
	s.logger.Info("EMAIL (dev) — restablecimiento de contraseña",
		zap.String("to", toEmail),
		zap.String("reset_link", link),
	)
	return nil
}

var _ services.EmailService = (*LogEmailService)(nil)
