// Package apiresponse where will be all errors responses
package apiresponse

type ErrorResponse struct {
	Error string `json:"error" example:"invalid credentials"`
}
