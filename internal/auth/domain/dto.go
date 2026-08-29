// Package domain users dto.
package domain

import (
	"errors"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

type RegisterRequest struct {
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Phone     string    `json:"phone"`
	AvatarURL string    `json:"avatar_url"`
	JobTitle  string    `json:"job_title"`
	Bio       string    `json:"bio"`
	RoleID    uuid.UUID `json:"role_id"`
}

type RegisterAdminRequest struct {
	RegisterRequest
	RoleID uuid.UUID `json:"role" example:"admin"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
