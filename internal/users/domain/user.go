// Package domain users define the entities users.
package domain

import (
	"errors"
	"strings"
	"time"

	pkgDomain "github.com/alisonsandrade/go-start-project/pkg/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

var ErrInvalidRole = errors.New("invalid role")

// ParseRole normaliza para minúsculo e valida contra a enum
func ParseRole(roleStr string) (Role, error) {
	normalized := Role(strings.ToLower(strings.TrimSpace(roleStr)))

	switch normalized {
	case RoleAdmin, RoleUser:
		return normalized, nil
	case "":
		// Define um fallback padrão se vier vazio
		return RoleUser, nil
	default:
		return "", ErrInvalidRole
	}
}

type User struct {
	ID        uuid.UUID          `gorm:"type:uuid;primary_key" json:"id"`
	Name      string             `gorm:"size:100;not null" json:"name"`
	Email     pkgDomain.Email    `gorm:"size:150;uniqueIndex;not null" json:"email"`
	Password  pkgDomain.Password `gorm:"not null" json:"-"`
	Role      Role               `gorm:"type:varchar(20);default:'user';not null" json:"role"`
	Phone     string             `gorm:"size:20" json:"phone"`
	AvatarURL string             `gorm:"size:255" json:"avatar_url"`
	JobTitle  string             `gorm:"size:100" json:"job_title"`
	Bio       string             `gorm:"type:text" json:"bio"`
	IsActive  bool               `gorm:"default:true;not null" json:"is_active"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
	DeletedAt gorm.DeletedAt     `gorm:"index" json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// NewUser creates a new user instance with the provided details.
func NewUser(name, rawEmail, rawPassword string, role Role) (*User, error) {
	email, err := pkgDomain.NewEmail(rawEmail)
	if err != nil {
		return nil, err
	}

	password, err := pkgDomain.NewPassword(rawPassword)
	if err != nil {
		return nil, err
	}

	return &User{
		Name:      strings.TrimSpace(name),
		Email:     email,
		Password:  password,
		Role:      role,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}, nil
}
