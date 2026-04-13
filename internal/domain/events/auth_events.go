// internal/domain/events/auth_events.go
package events

import "time"

// Tipos de eventos de autenticación
const (
	EventUserRegistered   = "USER_REGISTERED"
	EventUserLogin        = "USER_LOGIN"
	EventUserLogout       = "USER_LOGOUT"
	EventUserEmailChanged = "USER_EMAIL_CHANGED"
	EventUserRoleChanged  = "USER_ROLE_CHANGED"
)

// AuthEvent representa un evento de autenticación.
// Los consumidores (ej. user-service) materializan estos datos como vista local.
type AuthEvent struct {
	EventType string            `json:"event_type"`
	UserID    string            `json:"user_id"`
	Email     string            `json:"email,omitempty"`
	FullName  string            `json:"full_name,omitempty"`
	Phone     string            `json:"phone,omitempty"`
	Role      string            `json:"role,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// NewUserRegisteredEvent crea un evento de registro de usuario.
// Lleva los datos que user-service necesita para crear el perfil inicial proyectado.
func NewUserRegisteredEvent(userID, email, fullName, phone, role string) AuthEvent {
	return AuthEvent{
		EventType: EventUserRegistered,
		UserID:    userID,
		Email:     email,
		FullName:  fullName,
		Phone:     phone,
		Role:      role,
		Timestamp: time.Now(),
	}
}

// NewUserEmailChangedEvent crea un evento cuando el email cambia en auth-service.
// (Aún no hay endpoint que lo dispare; el consumer ya está preparado para recibirlo).
func NewUserEmailChangedEvent(userID, newEmail string) AuthEvent {
	return AuthEvent{
		EventType: EventUserEmailChanged,
		UserID:    userID,
		Email:     newEmail,
		Timestamp: time.Now(),
	}
}

// NewUserRoleChangedEvent crea un evento cuando el role cambia en auth-service.
func NewUserRoleChangedEvent(userID, newRole string) AuthEvent {
	return AuthEvent{
		EventType: EventUserRoleChanged,
		UserID:    userID,
		Role:      newRole,
		Timestamp: time.Now(),
	}
}

// NewUserLoginEvent crea un evento de login de usuario
func NewUserLoginEvent(userID string, email string) AuthEvent {
	return AuthEvent{
		EventType: EventUserLogin,
		UserID:    userID,
		Email:     email,
		Timestamp: time.Now(),
	}
}

// NewUserLogoutEvent crea un evento de logout de usuario
func NewUserLogoutEvent(userID string) AuthEvent {
	return AuthEvent{
		EventType: EventUserLogout,
		UserID:    userID,
		Timestamp: time.Now(),
	}
}
