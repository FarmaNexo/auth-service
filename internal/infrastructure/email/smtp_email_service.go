package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"

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
	subject, htmlBody, textBody := passwordResetContent(s.frontendURL, resetToken)
	msg := s.buildMessage(toEmail, subject, textBody, htmlBody)
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

// buildMessage arma un mensaje RFC 822 multipart/alternative con parte de texto plano
// y parte HTML (la parte de texto mejora la entregabilidad).
func (s *SMTPEmailService) buildMessage(to, subject, textBody, htmlBody string) []byte {
	const boundary = "farmanexo-alt-boundary"
	var b strings.Builder
	b.WriteString(fmt.Sprintf("From: %s\r\n", s.from))
	b.WriteString(fmt.Sprintf("To: %s\r\n", to))
	b.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n\r\n", boundary))

	b.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	b.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n\r\n")
	b.WriteString(textBody + "\r\n\r\n")

	b.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	b.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
	b.WriteString(htmlBody + "\r\n\r\n")

	b.WriteString(fmt.Sprintf("--%s--\r\n", boundary))
	return []byte(b.String())
}

var _ services.EmailService = (*SMTPEmailService)(nil)
