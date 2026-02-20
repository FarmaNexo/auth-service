// internal/shared/constants/message_codes.go
package constants

// MessageCode contiene todos los códigos de respuesta del sistema
type MessageCode string

const (
	// Success codes
	CodeSuccess        MessageCode = "SUCCESS_001"
	CodeCreatedSuccess MessageCode = "SUCCESS_002"
	CodeUpdatedSuccess MessageCode = "SUCCESS_003"
	CodeDeletedSuccess MessageCode = "SUCCESS_004"

	// Authentication codes
	CodeAuthSuccess    MessageCode = "AUTH_001"
	CodeLoginSuccess   MessageCode = "AUTH_002"
	CodeLogoutSuccess  MessageCode = "AUTH_003"
	CodeTokenRefreshed MessageCode = "AUTH_004"

	// Validation errors
	CodeValidationError  MessageCode = "VAL_001"
	CodeInvalidEmail     MessageCode = "VAL_002"
	CodeInvalidPassword  MessageCode = "VAL_003"
	CodePasswordTooShort MessageCode = "VAL_004"
	CodeRequiredField    MessageCode = "VAL_005"

	// Authentication errors
	CodeUnauthorized       MessageCode = "AUTH_ERR_001"
	CodeInvalidToken       MessageCode = "AUTH_ERR_002"
	CodeTokenExpired       MessageCode = "AUTH_ERR_003"
	CodeInvalidCredentials MessageCode = "AUTH_ERR_004"
	CodeUserNotFound       MessageCode = "AUTH_ERR_005"
	CodeUserInactive       MessageCode = "AUTH_ERR_006"

	// Business errors
	CodeUserAlreadyExists MessageCode = "BUS_001"
	CodeEmailAlreadyTaken MessageCode = "BUS_002"
	CodeResourceNotFound  MessageCode = "BUS_003"

	// System errors
	CodeInternalError      MessageCode = "SYS_001"
	CodeDatabaseError      MessageCode = "SYS_002"
	CodeServiceUnavailable MessageCode = "SYS_003"
)

// MessageDescription contiene las descripciones predefinidas
var MessageDescription = map[MessageCode]string{
	// Success
	CodeSuccess:        "Operación exitosa",
	CodeCreatedSuccess: "Recurso creado exitosamente",
	CodeUpdatedSuccess: "Recurso actualizado exitosamente",
	CodeDeletedSuccess: "Recurso eliminado exitosamente",

	// Auth success
	CodeAuthSuccess:    "Autenticación exitosa",
	CodeLoginSuccess:   "Inicio de sesión exitoso",
	CodeLogoutSuccess:  "Cierre de sesión exitoso",
	CodeTokenRefreshed: "Token actualizado exitosamente",

	// Validation
	CodeValidationError:  "Error de validación",
	CodeInvalidEmail:     "El formato del email es inválido",
	CodeInvalidPassword:  "La contraseña no cumple los requisitos",
	CodePasswordTooShort: "La contraseña debe tener al menos 8 caracteres",
	CodeRequiredField:    "Campo requerido",

	// Auth errors
	CodeUnauthorized:       "No autorizado",
	CodeInvalidToken:       "Token inválido",
	CodeTokenExpired:       "Token expirado",
	CodeInvalidCredentials: "Credenciales inválidas",
	CodeUserNotFound:       "Usuario no encontrado",
	CodeUserInactive:       "Usuario inactivo",

	// Business
	CodeUserAlreadyExists: "El usuario ya existe",
	CodeEmailAlreadyTaken: "El email ya está registrado",
	CodeResourceNotFound:  "Recurso no encontrado",

	// System
	CodeInternalError:      "Error interno del servidor",
	CodeDatabaseError:      "Error de base de datos",
	CodeServiceUnavailable: "Servicio no disponible",
}

// GetDescription retorna la descripción del código
func GetDescription(code MessageCode) string {
	if desc, ok := MessageDescription[code]; ok {
		return desc
	}
	return "Descripción no disponible"
}
