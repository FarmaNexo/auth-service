// internal/shared/common/response_extensions.go
package common

import (
	"github.com/farmanexo/auth-service/internal/shared/constants"
)

// ResponseBuilder proporciona métodos fluent para construir responses
// Equivalente a los extension methods de C#
type ResponseBuilder[T any] struct {
	response *ApiResponse[T]
}

// NewResponseBuilder crea un nuevo builder
func NewResponseBuilder[T any]() *ResponseBuilder[T] {
	return &ResponseBuilder[T]{
		response: NewApiResponse[T](),
	}
}

// ========================================
// BUILDER METHODS (Fluent API)
// ========================================

// WithError establece un error y retorna el builder
func (b *ResponseBuilder[T]) WithError(
	code constants.MessageCode,
	message string,
	statusCode constants.HTTPStatusCode,
) *ResponseBuilder[T] {
	b.response.SetHttpStatus(statusCode.Int())
	b.response.AddError(code, message)
	return b
}

// WithErrorSimple establece un error simple
func (b *ResponseBuilder[T]) WithErrorSimple(
	message string,
	statusCode constants.HTTPStatusCode,
) *ResponseBuilder[T] {
	b.response.SetHttpStatus(statusCode.Int())
	b.response.AddErrorSimple(message)
	return b
}

// WithSuccess establece datos exitosos
func (b *ResponseBuilder[T]) WithSuccess(
	data T,
	statusCode constants.HTTPStatusCode,
) *ResponseBuilder[T] {
	b.response.SetData(data)
	b.response.AddSuccessMessage()
	b.response.SetHttpStatus(statusCode.Int())
	return b
}

// WithData solo establece los datos sin mensaje
func (b *ResponseBuilder[T]) WithData(data T) *ResponseBuilder[T] {
	b.response.SetData(data)
	return b
}

// WithMessage agrega un mensaje personalizado
func (b *ResponseBuilder[T]) WithMessage(
	code constants.MessageCode,
	message string,
	messageType constants.MessageType,
) *ResponseBuilder[T] {
	b.response.AddMessageWithType(code, message, messageType)
	return b
}

// WithHttpStatus establece el código HTTP
func (b *ResponseBuilder[T]) WithHttpStatus(statusCode constants.HTTPStatusCode) *ResponseBuilder[T] {
	b.response.SetHttpStatus(statusCode.Int())
	return b
}

// Build retorna la respuesta construida
func (b *ResponseBuilder[T]) Build() *ApiResponse[T] {
	return b.response
}

// ========================================
// EXTENSION METHODS (Direct on ApiResponse)
// ========================================

// WithError - Extension method directo en ApiResponse
func WithError[T any](
	resp *ApiResponse[T],
	code constants.MessageCode,
	message string,
	statusCode constants.HTTPStatusCode,
) *ApiResponse[T] {
	resp.SetHttpStatus(statusCode.Int())
	resp.AddError(code, message)
	return resp
}

// WithSuccess - Extension method directo en ApiResponse
func WithSuccess[T any](
	resp *ApiResponse[T],
	data T,
	statusCode constants.HTTPStatusCode,
) *ApiResponse[T] {
	resp.SetData(data)
	resp.AddSuccessMessage()
	resp.SetHttpStatus(statusCode.Int())
	return resp
}

// ========================================
// FACTORY METHODS (Shortcuts)
// ========================================

// OkResponse crea una respuesta 200 OK con datos
func OkResponse[T any](data T) *ApiResponse[T] {
	return NewResponseBuilder[T]().
		WithSuccess(data, constants.StatusOK).
		Build()
}

// CreatedResponse crea una respuesta 201 Created
func CreatedResponse[T any](data T) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusCreated.Int())
	resp.SetData(data)
	resp.AddMessageWithType(
		constants.CodeCreatedSuccess,
		constants.GetDescription(constants.CodeCreatedSuccess),
		constants.MessageTypeSuccess,
	)
	return resp
}

// NoContentResponse crea una respuesta 204 No Content
func NoContentResponse[T any]() *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusNoContent.Int())
	return resp
}

// BadRequestResponse crea una respuesta 400 Bad Request
func BadRequestResponse[T any](code constants.MessageCode, message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusBadRequest.Int())
	resp.AddError(code, message)
	return resp
}

// UnauthorizedResponse crea una respuesta 401 Unauthorized
func UnauthorizedResponse[T any](message string) *ApiResponse[T] {
	return NewResponseBuilder[T]().
		WithError(
			constants.CodeUnauthorized,
			message,
			constants.StatusUnauthorized,
		).
		Build()
}

// ForbiddenResponse crea una respuesta 403 Forbidden
func ForbiddenResponse[T any](message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusForbidden.Int())
	resp.AddError(constants.CodeUnauthorized, message)
	return resp
}

// NotFoundResponse crea una respuesta 404 Not Found
func NotFoundResponse[T any](message string) *ApiResponse[T] {
	return NewResponseBuilder[T]().
		WithError(
			constants.CodeResourceNotFound,
			message,
			constants.StatusNotFound,
		).
		Build()
}

// ConflictResponse crea una respuesta 409 Conflict
func ConflictResponse[T any](code constants.MessageCode, message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusConflict.Int())
	resp.AddError(code, message)
	// NO establecer datos - Go serializa como null automáticamente
	return resp
}

// ValidationErrorResponse crea una respuesta de error de validación
func ValidationErrorResponse[T any](errors map[string]string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusBadRequest.Int())

	for field, errorMsg := range errors {
		resp.AddMessageWithType(
			constants.CodeValidationError,
			field+": "+errorMsg,
			constants.MessageTypeError,
		)
	}

	return resp
}

// TooManyRequestsResponse crea una respuesta 429 Too Many Requests
func TooManyRequestsResponse[T any](message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusTooManyRequests.Int())
	resp.AddError(constants.CodeRateLimitExceeded, message)
	return resp
}

// InternalServerErrorResponse crea una respuesta 500
func InternalServerErrorResponse[T any](message string) *ApiResponse[T] {
	resp := NewApiResponse[T]()
	resp.SetHttpStatus(constants.StatusInternalServerError.Int())
	resp.AddError(constants.CodeInternalError, message)
	return resp
}

// ========================================
// CONVERSION HELPERS
// ========================================

// ToMap convierte ApiResponse a map (útil para logging)
func ToMap[T any](resp *ApiResponse[T]) map[string]interface{} {
	return map[string]interface{}{
		"meta": map[string]interface{}{
			"mensajes":      resp.Meta.Messages,
			"idTransaccion": resp.Meta.IdTransaction,
			"resultado":     resp.Meta.Result,
			"timestamp":     resp.Meta.Timestamp,
		},
		"datos": resp.Data,
	}
}
