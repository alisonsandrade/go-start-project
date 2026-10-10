package domain

import (
	"encoding/json"

	"github.com/alisonsandrade/go-start-project/pkg/pagination"
)

type CreateTenantRequest struct {
	Name     string          `json:"name" validate:"required,min=2,max=150"`
	Document string          `json:"document" validate:"required,min=3,max=30"`
	Settings json.RawMessage `json:"settings" swaggertype:"object"`
}

type UpdateTenantRequest struct {
	Name     *string         `json:"name" validate:"omitempty,min=2,max=150"`
	Document *string         `json:"document" validate:"omitempty,min=3,max=30"`
	Settings json.RawMessage `json:"settings" swaggertype:"object"`
}

type TenantPageResponse struct {
	Data []Tenant        `json:"data"`
	Meta pagination.Meta `json:"meta"`
}
