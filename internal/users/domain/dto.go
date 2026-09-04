// Package domain users dto.
package domain

import (
	"github.com/alisonsandrade/go-start-project/pkg/pagination"
	"github.com/google/uuid"
)

// UserBase for interface
type UserBase struct {
	Name      string `json:"name" example:"Alison Andrade"`
	Email     string `json:"email" example:"alison.andrade@email.com"`
	Phone     string `json:"phone" example:"75988632510"`
	AvatarURL string `json:"avatar_url" example:"https://github.com/alisonsandrade/avatar.jpg"`
	JobTitle  string `json:"job_title" example:"Software Engineer"`
	Bio       string `json:"bio" example:"This is my bio"`
}

type UpdateUserRequest struct {
	Name      string `json:"name,omitempty" example:"Alison Andrade"`
	Phone     string `json:"phone,omitempty" example:"7599999999"`
	AvatarURL string `json:"avatar_url,omitempty" example:"https://github.com/alisonsandrade/avatar.jpg"`
	JobTitle  string `json:"job_title,omitempty" example:"Tech Lead"`
	Bio       string `json:"bio,omitempty" example:"Desenvolvedor Backend Go"`
}

// CreateUserRequest represents the payload for an admin to create any user's data.
type CreateUserRequest struct {
	UserBase
	Password string    `json:"password"`
	RoleID   uuid.UUID `json:"role_id"`
}

// CreateUserResponse represents the response admin to create new user.
type CreateUserResponse struct {
	UserBase
	RoleID uuid.UUID `json:"role_id"`
}

// AdminUpdateUserRequest represents the payload for an admin to update any user's data.
// Pointers are used to allow partial updates (PATCH behavior).
type AdminUpdateUserRequest struct {
	Name     *string    `json:"name,omitempty"`
	Email    *string    `json:"email,omitempty"`
	RoleID   *uuid.UUID `json:"role_id,omitempty"`
	IsActive *bool      `json:"is_active,omitempty"`
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
