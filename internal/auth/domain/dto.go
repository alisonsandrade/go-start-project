// Package domain users dto.
package domain

import (
	"errors"
	"regexp"
	"strings"

	"github.com/alisonsandrade/go-start-project/internal/users/domain"
	"github.com/google/uuid"
)

type RegisterRequest struct {
	Name      string    `json:"name" example:"Antônio Silva"`
	Email     string    `json:"email" example:"antonio.silva@email.com"`
	Password  string    `json:"password" example:"senha123!"`
	Phone     string    `json:"phone" example:"75988094321"`
	AvatarURL string    `json:"avatar_url" example:"https://github.com/antonio.silva/avatar.jpg"`
	JobTitle  string    `json:"job_title" example:"Desenvolvedor Backend Go"`
	Bio       string    `json:"bio" example:"Minha bio"`
	RoleID    uuid.UUID `json:"role_id" example:"216aaf06-34b6-3d26-8866-ad2f94babeea"`
}

type RegisterAdminRequest struct {
	RegisterRequest
	RoleID uuid.UUID `json:"role" example:"215aaf06-24b6-4d26-8855-ad2f94bartea"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponseDTO struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         domain.User `json:"user"`
}

// ForgotPasswordDTO is the payload for get recuperation
type ForgotPasswordDTO struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordDTO is the payload for to redefined the password
type ResetPasswordDTO struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type RefreshTokenDTO struct {
	RefreshToken string `json:"refresh_token" binding:"required" example:"dGhpcy1pcy1hLXJlZnJlc2gtdG9rZW4..."`
}

// ChangePasswordDTO is the payload for change user logged password
type ChangePasswordDTO struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func (d *RegisterRequest) Validate() error {
	d.Name = strings.TrimSpace(d.Name)
	d.Email = strings.ToLower(strings.TrimSpace(d.Email))

	if len(d.Name) < 3 {
		return errors.New("name must be at least 3 characters long")
	}
	if !emailRegex.MatchString(d.Email) {
		return errors.New("invalid email address")
	}
	if len(d.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	return nil
}
