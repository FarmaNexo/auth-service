package email

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/farmanexo/auth-service/internal/domain/services"
	"go.uber.org/zap"
)

// SMTPEmailService envía correos transaccionales vía SMTP. En local apunta a Mailpit
// (localhost:1025, sin autenticación) y los correos quedan en su bandeja web; en
// producción se configura con el host y credenciales de SES o del proveedor de correo.
type SMTPEmailService struct {
	host        string
	port        string
	from        string
	username    string
	password    string
	frontendURL string
	logger      *zap.Logger
}

// NewSMTPEmailService crea el servicio de correo por SMTP.
func NewSMTPEmailService(host, port, from, username, password, frontendURL string, logger *zap.Logger) *SMTPEmailService {
	return &SMTPEmailService{
		host:        host,
		port:        port,
		from:        from,
		username:    username,
		password:    password,
		frontendURL: frontendURL,
		logger:      logger,
	}
}

func (s *SMTPEmailService) SendPasswordReset(ctx context.Context, toEmail string, resetToken string) error {
	subject, htmlBody := passwordResetContent(s.frontendURL, resetToken)
	msg := s.buildMessage(toEmail, subject, htmlBody)
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	// La autenticación solo se usa si hay credenciales (Mailpit local no las requiere).
	var auth smtp.Auth
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	if err := smtp.SendMail(addr, auth, s.from, []string{toEmail}, msg); err != nil {
		s.logger.Error("Error enviando correo de restablecimiento vía SMTP",
			zap.String("to", toEmail),
			zap.String("smtp", addr),
			zap.Error(err),
		)
		return err
	}

	s.logger.Info("Correo de restablecimiento enviado",
		zap.String("to", toEmail),
		zap.String("smtp", addr),
	)
	return nil
}

// buildMessage arma el mensaje RFC 822 con cuerpo HTML.
func (s *SMTPEmailService) buildMessage(to, subject, htmlBody string) []byte {
	headers := fmt.Sprintf("From: %s\r\n", s.from)
	headers += fmt.Sprintf("To: %s\r\n", to)
	headers += fmt.Sprintf("Subject: %s\r\n", subject)
	headers += "MIME-Version: 1.0\r\n"
	headers += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	return []byte(headers + "\r\n" + htmlBody)
}

var _ services.EmailService = (*SMTPEmailService)(nil)
