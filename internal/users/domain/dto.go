// Package domain users dto.
package domain

import (
	"github.com/alisonsandrade/go-start-project/pkg/pagination"
	"github.com/google/uuid"
)

// UserBase for interface
type UserBase struct {
	Name      string `json:"name" validate:"required,min=3,max=100" example:"Alison Andrade"`
	Email     string `json:"email" validate:"required,email,max=150" example:"alison.andrade@email.com"`
	Phone     string `json:"phone" validate:"omitempty,max=20" example:"75988632510"`
	AvatarURL string `json:"avatar_url" validate:"omitempty,url,max=255" example:"https://github.com/alisonsandrade/avatar.jpg"`
	JobTitle  string `json:"job_title" validate:"omitempty,max=100" example:"Software Engineer"`
	Bio       string `json:"bio" validate:"omitempty,max=255" example:"This is my bio"`
}

type UpdateUserRequest struct {
	Name      string `json:"name,omitempty" validate:"omitempty,min=3,max=100" example:"Alison Andrade"`
	Phone     string `json:"phone,omitempty" validate:"omitempty,max=20" example:"7599999999"`
	AvatarURL string `json:"avatar_url,omitempty" validate:"omitempty,url,max=255" example:"https://github.com/alisonsandrade/avatar.jpg"`
	JobTitle  string `json:"job_title,omitempty" validate:"omitempty,max=100" example:"Tech Lead"`
	Bio       string `json:"bio,omitempty" validate:"omitempty,max=255" example:"Desenvolvedor Backend Go"`
}

// CreateUserRequest represents the payload for an admin to create any user's data.
type CreateUserRequest struct {
	UserBase
	Password string    `json:"password" validate:"required,min=8"`
	RoleID   uuid.UUID `json:"role_id" validate:"required,uuid"`
}

// CreateUserResponse represents the response admin to create new user.
type CreateUserResponse struct {
	UserBase
	RoleID uuid.UUID `json:"role_id"`
}

// AdminUpdateUserRequest represents the payload for an admin to update any user's data.
// Pointers are used to allow partial updates (PATCH behavior).
type AdminUpdateUserRequest struct {
	Name     *string    `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Email    *string    `json:"email,omitempty" validate:"omitempty,email,max=150"`
	RoleID   *uuid.UUID `json:"role_id,omitempty" validate:"omitempty,uuid"`
	IsActive *bool      `json:"is_active,omitempty"`
}

type ChangeUserTenantRequest struct {
	TenantID uuid.UUID `json:"tenant_id" validate:"required,uuid"`
	RoleID   uuid.UUID `json:"role_id" validate:"required,uuid"`
}

type UserResponse struct {
	ID     uuid.UUID `json:"id"`
	Name   string    `json:"name"`
	Email  string    `json:"email"`
	Phone  string    `json:"phone"`
	RoleID uuid.UUID `json:"role_id"`

	Role struct {
		ID   uuid.UUID `json:"id"`
		Name string    `json:"name"`
	} `json:"role"`
}

// UserPageResponse define a estrutura de resposta documentada no Swagger
type UserPageResponse struct {
	Data []User          `json:"data"`
	Meta pagination.Meta `json:"meta"`
}
