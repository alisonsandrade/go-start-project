// Package platform domain-agnostic infrastructure (shared internally)
package platform

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"

	"github.com/alisonsandrade/go-start-project/pkg/apiresponse"
	"github.com/go-playground/validator/v10"
)

func JSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// ErrorWithCodeJSON escreve uma resposta de erro estruturada com código e mensagem.
func ErrorWithCodeJSON(w http.ResponseWriter, statusCode int, code, message string) {
	JSON(w, statusCode, apiresponse.NewError(message, code))
}

// ErrorJSON escreve uma resposta de erro com mensagem padrão usando ErrorResponse.
func ErrorJSON(w http.ResponseWriter, statusCode int, message string) {
	code := http.StatusText(statusCode)
	code = strings.ToUpper(strings.ReplaceAll(code, " ", "_"))
	ErrorWithCodeJSON(w, statusCode, code, message)
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
	resp := apiresponse.NewValidationError("Dados fornecidos são inválidos", details)
	JSON(w, http.StatusUnprocessableEntity, resp)
}
