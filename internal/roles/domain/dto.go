// Package domain
package domain

import (
	"github.com/alisonsandrade/go-start-project/pkg/pagination"
	"github.com/google/uuid"
)

type CreateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=50"`
	Description string `json:"description"`
}

type UpdateRoleRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=50"`
	Description string `json:"description"`
}

type ReplacePermissionsRequest struct {
	PermissionIDs []uuid.UUID `json:"permission_ids"`
}

// RolePageResponse define a estrutura de resposta documentada no Swagger
type RolePageResponse struct {
	Data []RoleEntity    `json:"data"`
	Meta pagination.Meta `json:"meta"`
}
