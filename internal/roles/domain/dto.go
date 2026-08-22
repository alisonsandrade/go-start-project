// Package domain
package domain

import "github.com/google/uuid"

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
