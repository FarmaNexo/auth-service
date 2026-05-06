// internal/presentation/dto/responses/consents_status_response.go
package responses

// PendingConsent describe un consentimiento cuya versión aceptada quedó
// desactualizada respecto a la versión vigente.
type PendingConsent struct {
	ConsentType            string `json:"consent_type"`
	CurrentAcceptedVersion string `json:"current_accepted_version,omitempty"`
	RequiredVersion        string `json:"required_version"`
}

// ConsentsStatusResponse indica si el usuario tiene sus consentimientos al día.
type ConsentsStatusResponse struct {
	UpToDate bool             `json:"up_to_date"`
	Pending  []PendingConsent `json:"pending"`
}
