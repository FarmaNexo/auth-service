package email

import "fmt"

// passwordResetContent arma el asunto y el cuerpo (HTML + texto plano) del correo de
// restablecimiento. Lo comparten las implementaciones de EmailService (SMTP / SES).
// Se incluye una parte text/plain además del HTML porque algunos filtros penalizan los
// correos HTML-only y baja la entregabilidad.
func passwordResetContent(frontendURL, resetToken string) (subject, htmlBody, textBody string) {
	link := fmt.Sprintf("%s/restablecer-contrasena?token=%s", frontendURL, resetToken)
	subject = "Restablece tu contraseña - FarmaNexo"
	htmlBody = fmt.Sprintf(`<!doctype html><html><body style="font-family:Arial,Helvetica,sans-serif;color:#222;line-height:1.5">
<h2>Restablece tu contraseña</h2>
<p>Recibimos una solicitud para restablecer la contraseña de tu cuenta en FarmaNexo.</p>
<p><a href="%s" style="display:inline-block;background:#db1a85;color:#fff;padding:12px 20px;border-radius:8px;text-decoration:none">Restablecer contraseña</a></p>
<p>Si el botón no funciona, copia este enlace en tu navegador:<br><a href="%s">%s</a></p>
<p style="color:#666;font-size:13px">El enlace vence en 1 hora. Si no solicitaste esto, ignora este correo.</p>
</body></html>`, link, link, link)
	textBody = fmt.Sprintf(`Restablece tu contraseña - FarmaNexo

Recibimos una solicitud para restablecer la contraseña de tu cuenta en FarmaNexo.

Abre este enlace para continuar:
%s

El enlace vence en 1 hora. Si no solicitaste esto, ignora este correo.`, link)
	return subject, htmlBody, textBody
}
