// Package apiresponse where will be all errors responses
package apiresponse

type ErrorResponse struct {
	Message string            `json:"message" example:"mensagem de erro"`
	Code    string            `json:"code,omitempty" example:"VALIDATION_FAILED"`
	Details map[string]string `json:"details,omitempty"`
}

// NewError cria uma nova resposta de erro simples com mensagem e código.
func NewError(message, code string) ErrorResponse {
	return ErrorResponse{
		Message: message,
		Code:    code,
	}
}

// NewValidationError cria uma resposta de erro de validação com código VALIDATION_FAILED e detalhes por campo.
func NewValidationError(message string, details map[string]string) ErrorResponse {
	return ErrorResponse{
		Message: message,
		Code:    "VALIDATION_FAILED",
		Details: details,
	}
}
