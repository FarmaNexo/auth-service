// internal/presentation/dto/responses/consent_response.go
package responses

import (
	"time"

	"github.com/farmanexo/auth-service/internal/domain/entities"
)

// ConsentItem representa un único evento de consentimiento del usuario.
type ConsentItem struct {
	ID              string     `json:"id"`
	ConsentType     string     `json:"consent_type"`
	DocumentVersion string     `json:"document_version"`
	Accepted        bool       `json:"accepted"`
	AcceptedAt      time.Time  `json:"accepted_at"`
	WithdrawnAt     *time.Time `json:"withdrawn_at,omitempty"`
	IsActive        bool       `json:"is_active"`
}

// ConsentsListResponse devuelve el historial de consentimientos de un usuario.
type ConsentsListResponse struct {
	Consents []ConsentItem `json:"consents"`
	Total    int           `json:"total"`
}

// NewConsentsListResponse transforma entidades de dominio a DTO.
func NewConsentsListResponse(consents []*entities.UserConsent) *ConsentsListResponse {
	items := make([]ConsentItem, 0, len(consents))
	for _, c := range consents {
		items = append(items, ConsentItem{
			ID:              c.ID.String(),
			ConsentType:     c.ConsentType,
			DocumentVersion: c.DocumentVersion,
			Accepted:        c.Accepted,
			AcceptedAt:      c.AcceptedAt,
			WithdrawnAt:     c.WithdrawnAt,
			IsActive:        c.IsActive(),
		})
	}
	return &ConsentsListResponse{
		Consents: items,
		Total:    len(items),
	}
}
