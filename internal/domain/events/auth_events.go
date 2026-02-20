// internal/domain/events/auth_events.go
package events

import "time"

// Tipos de eventos de autenticación
const (
	EventUserRegistered = "USER_REGISTERED"
	EventUserLogin      = "USER_LOGIN"
	EventUserLogout     = "USER_LOGOUT"
)

// AuthEvent representa un evento de autenticación
type AuthEvent struct {
	EventType string            `json:"event_type"`
	UserID    string            `json:"user_id"`
	Email     string            `json:"email,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// NewUserRegisteredEvent crea un evento de registro de usuario
func NewUserRegisteredEvent(userID string, email string) AuthEvent {
	return AuthEvent{
		EventType: EventUserRegistered,
		UserID:    userID,
		Email:     email,
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
