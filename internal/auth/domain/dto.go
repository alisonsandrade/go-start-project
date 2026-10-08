// Package domain users dto.
package domain

import (
	"errors"
	"regexp"
	"strings"

	"github.com/alisonsandrade/go-start-project/internal/users/domain"
	"github.com/google/uuid"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

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

type RegisterAdminRequest struct {
	RegisterRequest
	RoleID uuid.UUID `json:"role" example:"215aaf06-24b6-4d26-8855-ad2f94bartea"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (d *LoginRequest) Validate() error {
	d.Email = strings.ToLower(strings.TrimSpace(d.Email))
	d.Password = strings.TrimSpace(d.Password)

	if d.Email == "" {
		return errors.New("email is required")
	}
	if !emailRegex.MatchString(d.Email) {
		return errors.New("invalid email address")
	}
	if d.Password == "" {
		return errors.New("password is required")
	}

	return nil
}

type AuthResponseDTO struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	User         domain.User `json:"user"`
}

// ForgotPasswordDTO is the payload for request password recovery
type ForgotPasswordDTO struct {
	Email string `json:"email"`
}

func (d *ForgotPasswordDTO) Validate() error {
	d.Email = strings.ToLower(strings.TrimSpace(d.Email))
	if d.Email == "" {
		return errors.New("email is required")
	}
	if !emailRegex.MatchString(d.Email) {
		return errors.New("invalid email address")
	}
	return nil
}

// ResetPasswordDTO is the payload to redefine the password
type ResetPasswordDTO struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

func (d *ResetPasswordDTO) Validate() error {
	d.Token = strings.TrimSpace(d.Token)
	d.NewPassword = strings.TrimSpace(d.NewPassword)

	if d.Token == "" {
		return errors.New("token is required")
	}
	if len(d.NewPassword) < 8 {
		return errors.New("new password must be at least 8 characters long")
	}
	return nil
}

type RefreshTokenDTO struct {
	RefreshToken string `json:"refresh_token" example:"dGhpcy1pcy1hLXJlZnJlc2gtdG9rZW4..."`
}

func (d *RefreshTokenDTO) Validate() error {
	d.RefreshToken = strings.TrimSpace(d.RefreshToken)
	if d.RefreshToken == "" {
		return errors.New("refresh_token is required")
	}
	return nil
}

// ChangePasswordDTO is the payload for logged user password changes
type ChangePasswordDTO struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

func (d *ChangePasswordDTO) Validate() error {
	d.CurrentPassword = strings.TrimSpace(d.CurrentPassword)
	d.NewPassword = strings.TrimSpace(d.NewPassword)

	if d.CurrentPassword == "" {
		return errors.New("current password is required")
	}
	if len(d.NewPassword) < 8 {
		return errors.New("new password must be at least 8 characters long")
	}
	if d.CurrentPassword == d.NewPassword {
		return errors.New("new password cannot be identical to current password")
	}
	return nil
}
