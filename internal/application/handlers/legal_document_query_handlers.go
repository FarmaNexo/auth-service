// internal/application/handlers/legal_document_query_handlers.go
package handlers

import (
	"context"

	"github.com/farmanexo/auth-service/internal/application/commands"
	"github.com/farmanexo/auth-service/internal/domain/repositories"
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
	"github.com/farmanexo/auth-service/internal/shared/common"
	"github.com/farmanexo/auth-service/pkg/mediator"
	"go.uber.org/zap"
)

const defaultLegalLocale = "es-PE"

// ============================================================
// GetLegalDocumentTypesHandler
// ============================================================

type GetLegalDocumentTypesHandler struct {
	typeRepo repositories.LegalDocumentTypeRepository
	logger   *zap.Logger
}

func NewGetLegalDocumentTypesHandler(typeRepo repositories.LegalDocumentTypeRepository, logger *zap.Logger) *GetLegalDocumentTypesHandler {
	return &GetLegalDocumentTypesHandler{typeRepo: typeRepo, logger: logger}
}

func (h *GetLegalDocumentTypesHandler) Handle(
	ctx context.Context,
	cmd commands.GetLegalDocumentTypesCommand,
) (*common.ApiResponse[responses.LegalDocumentTypesResponse], error) {
	types, err := h.typeRepo.FindAllActive(ctx)
	if err != nil {
		h.logger.Error("Error listando tipos de documento legal", zap.Error(err))
		return common.InternalServerErrorResponse[responses.LegalDocumentTypesResponse](
			"Error consultando tipos de documento",
		), nil
	}
	resp := responses.NewLegalDocumentTypesResponse(types)
	return common.OkResponse(resp), nil
}

var _ mediator.RequestHandler[commands.GetLegalDocumentTypesCommand, responses.LegalDocumentTypesResponse] = (*GetLegalDocumentTypesHandler)(nil)

// ============================================================
// GetCurrentLegalDocumentHandler
// ============================================================

type GetCurrentLegalDocumentHandler struct {
	docRepo repositories.LegalDocumentRepository
	logger  *zap.Logger
}

func NewGetCurrentLegalDocumentHandler(docRepo repositories.LegalDocumentRepository, logger *zap.Logger) *GetCurrentLegalDocumentHandler {
	return &GetCurrentLegalDocumentHandler{docRepo: docRepo, logger: logger}
}

func (h *GetCurrentLegalDocumentHandler) Handle(
	ctx context.Context,
	cmd commands.GetCurrentLegalDocumentCommand,
) (*common.ApiResponse[responses.LegalDocumentResponse], error) {
	locale := cmd.Locale
	if locale == "" {
		locale = defaultLegalLocale
	}

	doc, err := h.docRepo.FindCurrentPublished(ctx, cmd.TypeCode, locale)
	if err != nil {
		h.logger.Error("Error buscando documento vigente",
			zap.Error(err), zap.String("type", cmd.TypeCode), zap.String("locale", locale))
		return common.InternalServerErrorResponse[responses.LegalDocumentResponse](
			"Error consultando el documento legal",
		), nil
	}
	if doc == nil {
		return common.NotFoundResponse[responses.LegalDocumentResponse](
			"No existe versión vigente de este documento",
		), nil
	}

	resp := responses.NewLegalDocumentResponse(doc)
	return common.OkResponse(resp), nil
}

var _ mediator.RequestHandler[commands.GetCurrentLegalDocumentCommand, responses.LegalDocumentResponse] = (*GetCurrentLegalDocumentHandler)(nil)

// ============================================================
// GetLegalDocumentByVersionHandler
// ============================================================

type GetLegalDocumentByVersionHandler struct {
	docRepo repositories.LegalDocumentRepository
	logger  *zap.Logger
}

func NewGetLegalDocumentByVersionHandler(docRepo repositories.LegalDocumentRepository, logger *zap.Logger) *GetLegalDocumentByVersionHandler {
	return &GetLegalDocumentByVersionHandler{docRepo: docRepo, logger: logger}
}

func (h *GetLegalDocumentByVersionHandler) Handle(
	ctx context.Context,
	cmd commands.GetLegalDocumentByVersionCommand,
) (*common.ApiResponse[responses.LegalDocumentResponse], error) {
	locale := cmd.Locale
	if locale == "" {
		locale = defaultLegalLocale
	}

	doc, err := h.docRepo.FindByTypeAndVersion(ctx, cmd.TypeCode, cmd.Version, locale)
	if err != nil {
		h.logger.Error("Error buscando documento por versión",
			zap.Error(err), zap.String("type", cmd.TypeCode), zap.String("version", cmd.Version))
		return common.InternalServerErrorResponse[responses.LegalDocumentResponse](
			"Error consultando el documento legal",
		), nil
	}
	if doc == nil {
		return common.NotFoundResponse[responses.LegalDocumentResponse](
			"Versión no encontrada",
		), nil
	}

	resp := responses.NewLegalDocumentResponse(doc)
	return common.OkResponse(resp), nil
}

var _ mediator.RequestHandler[commands.GetLegalDocumentByVersionCommand, responses.LegalDocumentResponse] = (*GetLegalDocumentByVersionHandler)(nil)

// ============================================================
// ListLegalDocumentVersionsHandler
// ============================================================

type ListLegalDocumentVersionsHandler struct {
	docRepo repositories.LegalDocumentRepository
	logger  *zap.Logger
}

func NewListLegalDocumentVersionsHandler(docRepo repositories.LegalDocumentRepository, logger *zap.Logger) *ListLegalDocumentVersionsHandler {
	return &ListLegalDocumentVersionsHandler{docRepo: docRepo, logger: logger}
}

func (h *ListLegalDocumentVersionsHandler) Handle(
	ctx context.Context,
	cmd commands.ListLegalDocumentVersionsCommand,
) (*common.ApiResponse[responses.LegalDocumentVersionsResponse], error) {
	locale := cmd.Locale
	if locale == "" {
		locale = defaultLegalLocale
	}

	docs, err := h.docRepo.ListVersionsByType(ctx, cmd.TypeCode, locale)
	if err != nil {
		h.logger.Error("Error listando versiones",
			zap.Error(err), zap.String("type", cmd.TypeCode))
		return common.InternalServerErrorResponse[responses.LegalDocumentVersionsResponse](
			"Error consultando versiones",
		), nil
	}

	resp := responses.NewLegalDocumentVersionsResponse(cmd.TypeCode, docs)
	return common.OkResponse(resp), nil
}

var _ mediator.RequestHandler[commands.ListLegalDocumentVersionsCommand, responses.LegalDocumentVersionsResponse] = (*ListLegalDocumentVersionsHandler)(nil)
