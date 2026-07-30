package commands

// ResetPasswordCommand restablece la contraseña usando un token válido.
type ResetPasswordCommand struct {
	Token       string
	NewPassword string
}

// GetName retorna el nombre del comando.
func (c ResetPasswordCommand) GetName() string {
	return "ResetPasswordCommand"
}
