// internal/presentation/http/controllers/legal_controller.go
package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/farmanexo/auth-service/internal/application/commands"
	"github.com/farmanexo/auth-service/internal/presentation/dto/responses"
	"github.com/farmanexo/auth-service/internal/shared/common"
	"github.com/farmanexo/auth-service/pkg/mediator"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// LegalController expone los documentos legales versionados (Términos, Privacidad, etc.).
// Todas las queries son públicas (no requieren JWT) — el contenido legal vigente debe ser
// accesible incluso para usuarios no autenticados (p.ej. página /terminos antes de registrarse).
type LegalController struct {
	mediator *mediator.Mediator
	logger   *zap.Logger
}

func NewLegalController(med *mediator.Mediator, logger *zap.Logger) *LegalController {
	return &LegalController{mediator: med, logger: logger}
}

// ListTypes godoc
// @Summary      Listar tipos de documento legal
// @Description  Devuelve los tipos de documento legal activos (terms, privacy, marketing, ...).
// @Tags         Legal
// @Produce      json
// @Success      200  {object}  common.ApiResponse[responses.LegalDocumentTypesResponse]
// @Failure      500  {object}  common.ApiResponse[responses.LegalDocumentTypesResponse]
// @Router       /api/v1/auth/legal/types [get]
func (c *LegalController) ListTypes(w http.ResponseWriter, r *http.Request) {
	cmd := commands.GetLegalDocumentTypesCommand{}
	resp, err := mediator.Send[commands.GetLegalDocumentTypesCommand, responses.LegalDocumentTypesResponse](
		r.Context(), c.mediator, cmd,
	)
	if err != nil {
		c.logger.Error("Error listando tipos legal", zap.Error(err))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.LegalDocumentTypesResponse](
			"Error consultando tipos de documento",
		))
		return
	}
	c.respondJSON(w, resp)
}

// GetCurrent godoc
// @Summary      Obtener versión vigente
// @Description  Devuelve la versión actualmente publicada del tipo de documento solicitado.
// @Tags         Legal
// @Produce      json
// @Param        type_code  path  string  true  "Código del tipo (terms, privacy, marketing)"
// @Param        locale     query string  false "Locale (default: es-PE)"
// @Success      200  {object}  common.ApiResponse[responses.LegalDocumentResponse]
// @Failure      404  {object}  common.ApiResponse[responses.LegalDocumentResponse]
// @Router       /api/v1/auth/legal/{type_code} [get]
func (c *LegalController) GetCurrent(w http.ResponseWriter, r *http.Request) {
	typeCode := chi.URLParam(r, "type_code")
	locale := r.URL.Query().Get("locale")

	cmd := commands.GetCurrentLegalDocumentCommand{TypeCode: typeCode, Locale: locale}
	resp, err := mediator.Send[commands.GetCurrentLegalDocumentCommand, responses.LegalDocumentResponse](
		r.Context(), c.mediator, cmd,
	)
	if err != nil {
		c.logger.Error("Error obteniendo documento vigente",
			zap.Error(err), zap.String("type", typeCode))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.LegalDocumentResponse](
			"Error consultando el documento legal",
		))
		return
	}
	c.respondJSON(w, resp)
}

// GetByVersion godoc
// @Summary      Obtener versión específica
// @Description  Devuelve una versión específica (incluyendo archivada). Necesario para ARCO: el usuario debe poder ver el texto que aceptó en su momento.
// @Tags         Legal
// @Produce      json
// @Param        type_code  path  string  true  "Código del tipo (terms, privacy, marketing)"
// @Param        version    path  string  true  "Versión semver (ej. 1.0.0)"
// @Param        locale     query string  false "Locale (default: es-PE)"
// @Success      200  {object}  common.ApiResponse[responses.LegalDocumentResponse]
// @Failure      404  {object}  common.ApiResponse[responses.LegalDocumentResponse]
// @Router       /api/v1/auth/legal/{type_code}/versions/{version} [get]
func (c *LegalController) GetByVersion(w http.ResponseWriter, r *http.Request) {
	typeCode := chi.URLParam(r, "type_code")
	version := chi.URLParam(r, "version")
	locale := r.URL.Query().Get("locale")

	cmd := commands.GetLegalDocumentByVersionCommand{
		TypeCode: typeCode,
		Version:  version,
		Locale:   locale,
	}
	resp, err := mediator.Send[commands.GetLegalDocumentByVersionCommand, responses.LegalDocumentResponse](
		r.Context(), c.mediator, cmd,
	)
	if err != nil {
		c.logger.Error("Error obteniendo documento por versión",
			zap.Error(err), zap.String("type", typeCode), zap.String("version", version))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.LegalDocumentResponse](
			"Error consultando el documento legal",
		))
		return
	}
	c.respondJSON(w, resp)
}

// ListVersions godoc
// @Summary      Listar versiones publicadas/archivadas
// @Description  Lista todas las versiones publicadas y archivadas de un tipo, ordenadas por fecha de publicación descendente. Excluye drafts.
// @Tags         Legal
// @Produce      json
// @Param        type_code  path  string  true  "Código del tipo"
// @Param        locale     query string  false "Locale (default: es-PE)"
// @Success      200  {object}  common.ApiResponse[responses.LegalDocumentVersionsResponse]
// @Router       /api/v1/auth/legal/{type_code}/versions [get]
func (c *LegalController) ListVersions(w http.ResponseWriter, r *http.Request) {
	typeCode := chi.URLParam(r, "type_code")
	locale := r.URL.Query().Get("locale")

	cmd := commands.ListLegalDocumentVersionsCommand{TypeCode: typeCode, Locale: locale}
	resp, err := mediator.Send[commands.ListLegalDocumentVersionsCommand, responses.LegalDocumentVersionsResponse](
		r.Context(), c.mediator, cmd,
	)
	if err != nil {
		c.logger.Error("Error listando versiones",
			zap.Error(err), zap.String("type", typeCode))
		c.respondJSON(w, common.InternalServerErrorResponse[responses.LegalDocumentVersionsResponse](
			"Error consultando versiones",
		))
		return
	}
	c.respondJSON(w, resp)
}

func (c *LegalController) respondJSON(w http.ResponseWriter, response interface{}) {
	statusCode := http.StatusOK
	if resp, ok := response.(interface{ GetHttpStatus() *int }); ok {
		if hs := resp.GetHttpStatus(); hs != nil {
			statusCode = *hs
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		c.logger.Error("Error codificando respuesta JSON", zap.Error(err))
	}
}
