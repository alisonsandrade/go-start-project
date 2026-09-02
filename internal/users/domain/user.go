// Package domain users define the entities users.
package domain

import (
	"errors"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	rolesDomain "github.com/alisonsandrade/go-start-project/internal/roles/domain"
	pkgDomain "github.com/alisonsandrade/go-start-project/pkg/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	ErrInvalidRole       = errors.New("invalid role")
	ErrUserNameTooShort  = errors.New("user name must be at least 2 characters long")
	ErrUserNameTooLong   = errors.New("user name cannot exceed 100 characters")
	ErrUserRoleRequired  = errors.New("role id is required")
	ErrUserPhoneTooLong  = errors.New("phone cannot exceed 20 characters")
	ErrUserJobTitleLong  = errors.New("job title cannot exceed 100 characters")
	ErrUserAvatarTooLong = errors.New("avatar url cannot exceed 255 characters")
	ErrUserAvatarInvalid = errors.New("avatar url format is invalid")
)

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

// Validate verified the business rules and limites the database (tags GORM)
func (u *User) Validate() error {
	// 1. Name field validation
	u.Name = strings.TrimSpace(u.Name)
	nameLen := utf8.RuneCountInString(u.Name)
	if nameLen < 3 {
		return ErrUserNameTooShort
	}
	if nameLen > 100 {
		return ErrUserNameTooLong
	}

	// 2. RoleID validation
	if u.RoleID == uuid.Nil {
		return ErrUserRoleRequired
	}

	// 3. Phone validatino
	u.Phone = strings.TrimSpace(u.Phone)
	if utf8.RuneCountInString(u.Phone) > 20 {
		return ErrUserPhoneTooLong
	}

	// 4. JobTitle validation
	u.JobTitle = strings.TrimSpace(u.JobTitle)
	if utf8.RuneCountInString(u.JobTitle) > 100 {
		return ErrUserJobTitleLong
	}

	// 5. AvatarURL field validation
	u.AvatarURL = strings.TrimSpace(u.AvatarURL)
	if u.AvatarURL != "" {
		if utf8.RuneCountInString(u.AvatarURL) > 255 {
			return ErrUserAvatarTooLong
		}
		if _, err := url.ParseRequestURI(u.AvatarURL); err != nil {
			return ErrUserAvatarInvalid
		}
	}

	return nil
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// BeforeSave hook Gorm that is calling before create or update user
// Injects the validation into the GORM lifecycle
func (u *User) BeforeSave(tx *gorm.DB) error {
	if err := u.Validate(); err != nil {
		return err
	}
	return nil
}
