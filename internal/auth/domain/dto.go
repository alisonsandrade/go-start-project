// Package domain users dto.
package domain

import (
	"github.com/alisonsandrade/go-start-project/internal/users/domain"
	"github.com/google/uuid"
)

type RegisterRequest struct {
	Name      string    `json:"name" validate:"required,min=3,max=100" example:"Antônio Silva"`
	Email     string    `json:"email" validate:"required,email" example:"antonio.silva@email.com"`
	Password  string    `json:"password" validate:"required,min=8" example:"senha123!"`
	Phone     string    `json:"phone" validate:"omitempty,max=20" example:"75988094321"`
	AvatarURL string    `json:"avatar_url" validate:"omitempty,url,max=255" example:"https://github.com/antonio.silva/avatar.jpg"`
	JobTitle  string    `json:"job_title" validate:"omitempty,max=100" example:"Desenvolvedor Backend Go"`
	Bio       string    `json:"bio" validate:"omitempty,max=255" example:"Minha bio"`
	RoleID    uuid.UUID `json:"role_id" validate:"omitempty,uuid" example:"216aaf06-34b6-3d26-8866-ad2f94babeea"`
}

type RegisterAdminRequest struct {
	RegisterRequest
	RoleID uuid.UUID `json:"role" validate:"required,uuid" example:"215aaf06-24b6-4d26-8855-ad2f94bartea"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type AuthResponseDTO struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         domain.User `json:"user"`
}

// ForgotPasswordDTO is the payload for request password recovery
type ForgotPasswordDTO struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordDTO is the payload to redefine the password
type ResetPasswordDTO struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type RefreshTokenDTO struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"dGhpcy1pcy1hLXJlZnJlc2gtdG9rZW4..."`
}

// ChangePasswordDTO is the payload for logged user password changes
type ChangePasswordDTO struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8,nefield=CurrentPassword"`
}
