// internal/application/commands/legal_document_queries.go
package commands

import "github.com/farmanexo/auth-service/internal/presentation/dto/responses"

// GetLegalDocumentTypesCommand — lista de tipos activos (público).
type GetLegalDocumentTypesCommand struct{}

func (c GetLegalDocumentTypesCommand) GetName() string { return "GetLegalDocumentTypesCommand" }

type GetLegalDocumentTypesResponse = responses.LegalDocumentTypesResponse

// GetCurrentLegalDocumentCommand — versión vigente (status=published) de un tipo.
type GetCurrentLegalDocumentCommand struct {
	TypeCode string
	Locale   string
}

func (c GetCurrentLegalDocumentCommand) GetName() string { return "GetCurrentLegalDocumentCommand" }

type GetCurrentLegalDocumentResponse = responses.LegalDocumentResponse

// GetLegalDocumentByVersionCommand — versión específica (incluye archivada, para ARCO).
type GetLegalDocumentByVersionCommand struct {
	TypeCode string
	Version  string
	Locale   string
}

func (c GetLegalDocumentByVersionCommand) GetName() string { return "GetLegalDocumentByVersionCommand" }

type GetLegalDocumentByVersionResponse = responses.LegalDocumentResponse

// ListLegalDocumentVersionsCommand — todas las versiones publicadas/archivadas.
type ListLegalDocumentVersionsCommand struct {
	TypeCode string
	Locale   string
}

func (c ListLegalDocumentVersionsCommand) GetName() string { return "ListLegalDocumentVersionsCommand" }

type ListLegalDocumentVersionsResponse = responses.LegalDocumentVersionsResponse
