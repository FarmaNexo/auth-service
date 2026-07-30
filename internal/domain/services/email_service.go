package services

import "context"

// EmailService abstrae el envío de correos transaccionales. En local se usa una
// implementación que registra el contenido en el log; en producción se conecta a
// SES o a un proveedor externo (pendiente, fase 2).
type EmailService interface {
	// SendPasswordReset envía al usuario el enlace para restablecer su contraseña.
	SendPasswordReset(ctx context.Context, toEmail string, resetToken string) error
}
