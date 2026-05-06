// internal/presentation/dto/responses/legal_document_response.go
package responses

import (
	"time"

	"github.com/farmanexo/auth-service/internal/domain/entities"
)

// LegalDocumentTypeResponse — vista pública de un tipo de documento.
type LegalDocumentTypeResponse struct {
	Code         string `json:"code" example:"terms"`
	Name         string `json:"name" example:"Términos y Condiciones"`
	Description  string `json:"description,omitempty"`
	IsRequired   bool   `json:"is_required" example:"true"`
	DisplayOrder int    `json:"display_order" example:"1"`
}

func NewLegalDocumentTypeResponse(t *entities.LegalDocumentType) LegalDocumentTypeResponse {
	return LegalDocumentTypeResponse{
		Code:         t.Code,
		Name:         t.Name,
		Description:  t.Description,
		IsRequired:   t.IsRequired,
		DisplayOrder: t.DisplayOrder,
	}
}

// LegalDocumentTypesResponse — lista de tipos.
type LegalDocumentTypesResponse struct {
	Types []LegalDocumentTypeResponse `json:"types"`
}

func NewLegalDocumentTypesResponse(types []*entities.LegalDocumentType) LegalDocumentTypesResponse {
	out := make([]LegalDocumentTypeResponse, 0, len(types))
	for _, t := range types {
		out = append(out, NewLegalDocumentTypeResponse(t))
	}
	return LegalDocumentTypesResponse{Types: out}
}

// LegalDocumentResponse — vista pública de un documento legal.
// Incluye contenido completo (markdown) y metadatos relevantes para el cliente.
type LegalDocumentResponse struct {
	ID              string     `json:"id" example:"5e1c2..."`
	TypeCode        string     `json:"type_code" example:"terms"`
	TypeName        string     `json:"type_name" example:"Términos y Condiciones"`
	Version         string     `json:"version" example:"1.0.0"`
	Locale          string     `json:"locale" example:"es-PE"`
	Jurisdiction    string     `json:"jurisdiction" example:"PE"`
	Title           string     `json:"title"`
	Summary         string     `json:"summary,omitempty"`
	ContentMarkdown string     `json:"content_markdown"`
	ContentHash     string     `json:"content_hash"`
	Changelog       string     `json:"changelog,omitempty"`
	ChangeSeverity  string     `json:"change_severity,omitempty"`
	Status          string     `json:"status"`
	EffectiveDate   *time.Time `json:"effective_date,omitempty"`
	PublishedAt     *time.Time `json:"published_at,omitempty"`
}

func NewLegalDocumentResponse(d *entities.LegalDocument) LegalDocumentResponse {
	r := LegalDocumentResponse{
		ID:              d.ID.String(),
		Version:         d.Version,
		Locale:          d.Locale,
		Jurisdiction:    d.Jurisdiction,
		Title:           d.Title,
		Summary:         d.Summary,
		ContentMarkdown: d.ContentMarkdown,
		ContentHash:     d.ContentHash,
		Changelog:       d.Changelog,
		ChangeSeverity:  d.ChangeSeverity,
		Status:          string(d.Status),
		EffectiveDate:   d.EffectiveDate,
		PublishedAt:     d.PublishedAt,
	}
	if d.DocumentType != nil {
		r.TypeCode = d.DocumentType.Code
		r.TypeName = d.DocumentType.Name
	}
	return r
}

// LegalDocumentVersionResponse — versión histórica resumida (sin contenido completo).
type LegalDocumentVersionResponse struct {
	Version        string     `json:"version"`
	Locale         string     `json:"locale"`
	Status         string     `json:"status"`
	ChangeSeverity string     `json:"change_severity,omitempty"`
	Summary        string     `json:"summary,omitempty"`
	EffectiveDate  *time.Time `json:"effective_date,omitempty"`
	PublishedAt    *time.Time `json:"published_at,omitempty"`
	ArchivedAt     *time.Time `json:"archived_at,omitempty"`
}

type LegalDocumentVersionsResponse struct {
	TypeCode string                         `json:"type_code"`
	Versions []LegalDocumentVersionResponse `json:"versions"`
}

func NewLegalDocumentVersionsResponse(typeCode string, docs []*entities.LegalDocument) LegalDocumentVersionsResponse {
	versions := make([]LegalDocumentVersionResponse, 0, len(docs))
	for _, d := range docs {
		versions = append(versions, LegalDocumentVersionResponse{
			Version:        d.Version,
			Locale:         d.Locale,
			Status:         string(d.Status),
			ChangeSeverity: d.ChangeSeverity,
			Summary:        d.Summary,
			EffectiveDate:  d.EffectiveDate,
			PublishedAt:    d.PublishedAt,
			ArchivedAt:     d.ArchivedAt,
		})
	}
	return LegalDocumentVersionsResponse{
		TypeCode: typeCode,
		Versions: versions,
	}
}
