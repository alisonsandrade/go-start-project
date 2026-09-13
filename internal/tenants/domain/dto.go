package domain

import (
	"encoding/json"

	"github.com/alisonsandrade/go-start-project/pkg/pagination"
)

type CreateTenantRequest struct {
	Name     string          `json:"name"`
	Document string          `json:"document"`
	Settings json.RawMessage `json:"settings" swaggertype:"object"`
}

type UpdateTenantRequest struct {
	Name     *string         `json:"name"`
	Document *string         `json:"document"`
	Settings json.RawMessage `json:"settings" swaggertype:"object"`
}

type TenantPageResponse struct {
	Data []Tenant        `json:"data"`
	Meta pagination.Meta `json:"meta"`
}
