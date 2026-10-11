// Package platform domain-agnostic infrastructure (shared internally)
package platform

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

// ErrorCode representa o código de erro strongly-typed da aplicação.
type ErrorCode string

const (
	CodeInternalError       ErrorCode = "INTERNAL_ERROR"
	CodeValidationFailed    ErrorCode = "VALIDATION_FAILED"
	CodeInvalidRequestBody  ErrorCode = "INVALID_REQUEST_BODY"
	CodeUnauthorized        ErrorCode = "UNAUTHORIZED"
	CodeForbidden           ErrorCode = "FORBIDDEN"
	CodeNotFound            ErrorCode = "NOT_FOUND"
	CodeConflict            ErrorCode = "CONFLICT"
	CodeBadRequest          ErrorCode = "BAD_REQUEST"
	CodeInvalidUUID         ErrorCode = "INVALID_UUID"
	CodeInvalidToken        ErrorCode = "INVALID_TOKEN"
	CodeInvalidTokenFormat  ErrorCode = "INVALID_TOKEN_FORMAT"
	CodeInvalidTenantID     ErrorCode = "INVALID_TENANT_ID"
	CodeEmailAlreadyExists  ErrorCode = "EMAIL_ALREADY_EXISTS"
	CodeInvalidCredentials  ErrorCode = "INVALID_CREDENTIALS"
	CodeInvalidRole         ErrorCode = "INVALID_ROLE"
	CodeInvalidEmail        ErrorCode = "INVALID_EMAIL"
	CodeInvalidResetToken   ErrorCode = "INVALID_RESET_TOKEN"
	CodeExpiredResetToken   ErrorCode = "EXPIRED_RESET_TOKEN"
	CodeIncorrectPassword   ErrorCode = "INCORRECT_PASSWORD"
	CodeUserNotFound        ErrorCode = "USER_NOT_FOUND"
	CodeTenantNotFound      ErrorCode = "TENANT_NOT_FOUND"
	CodeTenantAlreadyExists ErrorCode = "TENANT_ALREADY_EXISTS"
	CodeInvalidTenantData   ErrorCode = "INVALID_TENANT_DATA"
	CodeRoleNotFound        ErrorCode = "ROLE_NOT_FOUND"
	CodeRoleAlreadyExists   ErrorCode = "ROLE_ALREADY_EXISTS"
	CodeInvalidPermissions  ErrorCode = "INVALID_PERMISSIONS"
	CodeSystemRoleImmutable ErrorCode = "SYSTEM_ROLE_IMMUTABLE"
)

// ErrorResponse representa o formato padrão de resposta de erro da API.
type ErrorResponse struct {
	Message string            `json:"message" example:"mensagem de erro"`
	Code    string            `json:"code,omitempty" example:"VALIDATION_FAILED"`
	Details map[string]string `json:"details,omitempty"`
}

// MessageResponse representa uma resposta simples contendo uma mensagem de sucesso.
type MessageResponse struct {
	Message string `json:"message" example:"operação realizada com sucesso"`
}

// NewError cria uma nova resposta de erro simples com mensagem e código.
func NewError(message string, code ErrorCode) ErrorResponse {
	return ErrorResponse{
		Message: message,
		Code:    string(code),
	}
}

// NewValidationError cria uma resposta de erro de validação com código VALIDATION_FAILED e detalhes por campo.
func NewValidationError(message string, details map[string]string) ErrorResponse {
	return ErrorResponse{
		Message: message,
		Code:    string(CodeValidationFailed),
		Details: details,
	}
}

// JSON escreve dados no formato JSON na resposta HTTP.
func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// ErrorWithCodeJSON escreve uma resposta de erro estruturada com código fortemente tipado e mensagem.
func ErrorWithCodeJSON(w http.ResponseWriter, statusCode int, code ErrorCode, message string) {
	JSON(w, statusCode, NewError(message, code))
}

// ErrorJSON escreve uma resposta de erro com mensagem padrão usando ErrorCode derivado do status HTTP.
func ErrorJSON(w http.ResponseWriter, statusCode int, message string) {
	code := http.StatusText(statusCode)
	code = strings.ToUpper(strings.ReplaceAll(code, " ", "_"))
	ErrorWithCodeJSON(w, statusCode, ErrorCode(code), message)
}

// toSnakeCase converte strings CamelCase para snake_case.
func toSnakeCase(str string) string {
	var b strings.Builder
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteRune('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// TranslateValidationError traduz os erros do go-playground/validator para mensagens amigáveis em português.
func TranslateValidationError(err error) map[string]string {
	details := make(map[string]string)
	var valErrors validator.ValidationErrors
	if !errors.As(err, &valErrors) {
		details["error"] = err.Error()
		return details
	}

	for _, fe := range valErrors {
		fieldName := toSnakeCase(fe.Field())
		var msg string

		switch fe.Tag() {
		case "required":
			msg = "Este campo é obrigatório"
		case "email":
			msg = "Formato de e-mail inválido"
		case "min":
			msg = fmt.Sprintf("Deve ter no mínimo %s caracteres", fe.Param())
		case "max":
			msg = fmt.Sprintf("Deve ter no máximo %s caracteres", fe.Param())
		case "url":
			msg = "Deve ser uma URL válida"
		case "uuid":
			msg = "Deve ser um identificador UUID válido"
		case "nefield":
			msg = fmt.Sprintf("Deve ser diferente do campo %s", toSnakeCase(fe.Param()))
		default:
			msg = fmt.Sprintf("Falha na validação da regra '%s'", fe.Tag())
		}

		details[fieldName] = msg
	}

	return details
}

// ValidationErrorJSON envia uma resposta HTTP 422 estruturada com os campos que falharam na validação.
func ValidationErrorJSON(w http.ResponseWriter, err error) {
	details := TranslateValidationError(err)
	resp := NewValidationError("Dados fornecidos são inválidos", details)
	JSON(w, http.StatusUnprocessableEntity, resp)
}
