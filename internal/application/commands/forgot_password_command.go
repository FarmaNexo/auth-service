package commands

// ForgotPasswordCommand solicita el envío del enlace de restablecimiento.
type ForgotPasswordCommand struct {
	Email string
}

// GetName retorna el nombre del comando.
func (c ForgotPasswordCommand) GetName() string {
	return "ForgotPasswordCommand"
}
