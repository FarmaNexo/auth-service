// internal/presentation/dto/requests/accept_consents_request.go
package requests

// AcceptConsentsRequest representa el body para re-aceptar consentimientos pendientes.
type AcceptConsentsRequest struct {
	ConsentTypes []string `json:"consent_types" validate:"required,min=1"`
}
