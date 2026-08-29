// Package domain users define the entities users.
package domain

import (
	"errors"
	"strings"
	"time"

	rolesDomain "github.com/alisonsandrade/go-start-project/internal/roles/domain"
	pkgDomain "github.com/alisonsandrade/go-start-project/pkg/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var ErrInvalidRole = errors.New("invalid role")

type User struct {
	ID        uuid.UUID              `gorm:"type:uuid;primary_key" json:"id"`
	Name      string                 `gorm:"size:100;not null" json:"name"`
	Email     pkgDomain.Email        `gorm:"size:150;uniqueIndex;not null" json:"email" swaggertype:"string"`
	Password  pkgDomain.Password     `gorm:"not null" json:"-" swaggerignore:"true"`
	Phone     string                 `gorm:"size:20" json:"phone"`
	RoleID    uuid.UUID              `gorm:"type:uuid;not null" json:"role_id" format:"uuid"`
	Role      rolesDomain.RoleEntity `gorm:"foreingKey:RoleID" json:"role" swaggertype:"object"`
	AvatarURL string                 `gorm:"size:255" json:"avatar_url"`
	JobTitle  string                 `gorm:"size:100" json:"job_title"`
	Bio       string                 `gorm:"type:text" json:"bio"`
	IsActive  bool                   `gorm:"default:true;not null" json:"is_active"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	DeletedAt gorm.DeletedAt         `gorm:"index" json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// NewUser creates a new user instance with the provided details.
func NewUser(name, rawEmail, rawPassword string, roleID uuid.UUID) (*User, error) {
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
		RoleID:    roleID,
		IsActive:  true,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}, nil
}
