// internal/domain/entities/user_consent.go
package entities

import (
	"time"

	"github.com/google/uuid"
)

// ========================================
// CONSENT TYPES (LPDP Ley 29733)
// ========================================

const (
	ConsentTypeTermsOfService         = "terms_of_service"
	ConsentTypePrivacyPolicy          = "privacy_policy"
	ConsentTypeMarketingCommunications = "marketing_communications"
)

// RequiredConsentTypes devuelve los tipos de consentimiento obligatorios al registro.
// El usuario no puede crear cuenta sin aceptar estos dos.
func RequiredConsentTypes() []string {
	return []string{ConsentTypeTermsOfService, ConsentTypePrivacyPolicy}
}

// LegalDocCodeFromConsentType traduce el consent_type interno (terms_of_service, privacy_policy, ...)
// al code de legal_document_types (terms, privacy, marketing). Retorna "" si no hay mapeo.
//
// Esta función es la frontera entre los nombres legacy de user_consents.consent_type
// y el catálogo nuevo de auth.legal_document_types. No los unificamos para no romper
// compatibilidad de los consents históricos ya almacenados.
func LegalDocCodeFromConsentType(consentType string) string {
	switch consentType {
	case ConsentTypeTermsOfService:
		return LegalDocCodeTerms
	case ConsentTypePrivacyPolicy:
		return LegalDocCodePrivacy
	case ConsentTypeMarketingCommunications:
		return LegalDocCodeMarketing
	default:
		return ""
	}
}

// UserConsent representa un evento de consentimiento otorgado o rechazado por un usuario.
// Cada fila es inmutable (salvo withdrawn_at al revocar). Nunca se hace UPDATE sobre otros campos.
type UserConsent struct {
	ID                       uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID                   uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	ConsentType              string     `gorm:"size:64;not null" json:"consent_type"`
	DocumentVersion          string     `gorm:"size:32;not null" json:"document_version"`
	DocumentID               *uuid.UUID `gorm:"type:uuid" json:"document_id,omitempty"`
	ContentHashAtAcceptance  string     `gorm:"size:64;column:content_hash_at_acceptance" json:"content_hash_at_acceptance,omitempty"`
	Accepted                 bool       `gorm:"not null" json:"accepted"`
	AcceptedAt               time.Time  `json:"accepted_at"`
	IPAddress                string     `gorm:"type:inet" json:"ip_address,omitempty"`
	UserAgent                string     `json:"user_agent,omitempty"`
	WithdrawnAt              *time.Time `json:"withdrawn_at,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
}

// TableName especifica el nombre de la tabla
func (UserConsent) TableName() string {
	return "auth.user_consents"
}

// ========================================
// DOMAIN METHODS
// ========================================

// IsActive indica si el consentimiento está vigente (no revocado).
func (c *UserConsent) IsActive() bool {
	return c.WithdrawnAt == nil
}

// Withdraw marca el consentimiento como revocado.
func (c *UserConsent) Withdraw() {
	now := time.Now()
	c.WithdrawnAt = &now
}

// ========================================
// FACTORY METHODS
// ========================================

// NewUserConsent crea un registro de consentimiento otorgado o rechazado.
func NewUserConsent(userID uuid.UUID, consentType, documentVersion string, accepted bool, ipAddress, userAgent string) *UserConsent {
	return &UserConsent{
		ID:              uuid.New(),
		UserID:          userID,
		ConsentType:     consentType,
		DocumentVersion: documentVersion,
		Accepted:        accepted,
		AcceptedAt:      time.Now(),
		IPAddress:       ipAddress,
		UserAgent:       userAgent,
		CreatedAt:       time.Now(),
	}
}

// NewUserConsentWithDocument crea un consentimiento ligado a un legal_document específico,
// guardando también el hash del contenido al momento de aceptación (defensa en profundidad).
func NewUserConsentWithDocument(
	userID uuid.UUID,
	consentType, documentVersion string,
	documentID uuid.UUID,
	contentHash string,
	accepted bool,
	ipAddress, userAgent string,
) *UserConsent {
	docID := documentID
	return &UserConsent{
		ID:                      uuid.New(),
		UserID:                  userID,
		ConsentType:             consentType,
		DocumentVersion:         documentVersion,
		DocumentID:              &docID,
		ContentHashAtAcceptance: contentHash,
		Accepted:                accepted,
		AcceptedAt:              time.Now(),
		IPAddress:               ipAddress,
		UserAgent:               userAgent,
		CreatedAt:               time.Now(),
	}
}
