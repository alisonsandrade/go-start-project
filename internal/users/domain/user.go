// Package domain users define the entities users.
package domain

import (
	"strings"

	baseDomain "github.com/alisonsandrade/go-start-project/internal/domain"
	rolesDomain "github.com/alisonsandrade/go-start-project/internal/roles/domain"
	pkgDomain "github.com/alisonsandrade/go-start-project/pkg/domain"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var validate = validator.New()

type User struct {
	baseDomain.BaseModelTenant `swaggerignore:"true"`
	Name                       string                 `gorm:"size:100;not null" json:"name" validate:"required,min=3,max=100"`
	Email                      pkgDomain.Email        `gorm:"size:150;uniqueIndex;not null" json:"email" swaggertype:"string"`
	Password                   pkgDomain.Password     `gorm:"not null" json:"-" swaggerignore:"true"`
	Phone                      string                 `gorm:"size:20" json:"phone" validate:"omitempty,max=20"`
	RoleID                     uuid.UUID              `gorm:"type:uuid;not null" json:"role_id" format:"uuid" validate:"required"`
	Role                       rolesDomain.RoleEntity `gorm:"foreignKey:RoleID" json:"role" swaggertype:"object"`
	AvatarURL                  string                 `gorm:"size:255" json:"avatar_url" validate:"omitempty,url,max=255"`
	JobTitle                   string                 `gorm:"size:100" json:"job_title" validate:"omitempty,max=100"`
	Bio                        string                 `gorm:"type:text" json:"bio"`
	IsActive                   bool                   `gorm:"default:true;not null" json:"is_active"`
	EmailVerified              bool                   `gorm:"default:false;not null" json:"email_verified"`
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
		Name:     strings.TrimSpace(name),
		Email:    email,
		Password: password,
		RoleID:   roleID,
		IsActive: true,
	}, nil
}

// Validate verifies the business rules and constraints using go-playground/validator.
func (u *User) Validate() error {
	u.Name = strings.TrimSpace(u.Name)
	u.Phone = strings.TrimSpace(u.Phone)
	u.JobTitle = strings.TrimSpace(u.JobTitle)
	u.AvatarURL = strings.TrimSpace(u.AvatarURL)

	return validate.Struct(u)
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	return u.BaseModelTenant.BeforeCreate(tx)
}

// BeforeSave hook Gorm that is called before create or update user.
// Injects the validation into the GORM lifecycle.
func (u *User) BeforeSave(tx *gorm.DB) error {
	if err := u.Validate(); err != nil {
		return err
	}
	return nil
}
