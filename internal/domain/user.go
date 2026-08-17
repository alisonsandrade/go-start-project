// Package domain define as entidades de negócio e DTOs.
package domain

import (
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin Role = "ADMIN"
	RoleUser  Role = "USER"
)

type User struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Email     string         `gorm:"size:150;uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"not null" json:"-"`
	Role      Role           `gorm:"type:varchar(20);default:'USER';not null" json:"role"`
	Phone     string         `gorm:"size:20" json:"phone"`
	AvatarURL string         `gorm:"size:255" json:"avatar_url"`
	JobTitle  string         `gorm:"size:100" json:"job_title"`
	Bio       string         `gorm:"type:text" json:"bio"`
	IsActive  bool           `gorm:"default:true;not null" json:"is_active"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (u *User) HashPassword() error {
	hasheBytes, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hasheBytes)
	return nil
}

func (u *User) CheckPassword(plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(plainPassword))
	return err == nil
}

/*************************************************************************************************
* 										DTOs de Entrada e saída
*************************************************************************************************/

type CreateUserDTO struct {
	Name      string `json:"name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Phone     string `json:"phone"`
	AvatarURL string `json:"avatar_url"`
	JobTitle  string `json:"job_title"`
	Bio       string `json:"bio"`
	Role      Role   `json:"role"`
}

type CreateAdminDTO struct {
	CreateUserDTO
	Role Role `json:"role" example:"ADMIN"`
}

type UpdateUserDTO struct {
	Name      string `json:"name,omitempty" example:"Alison Andrade"`
	Phone     string `json:"phone,omitempty" example:"7599999999"`
	AvatarURL string `json:"avatar_url,omitempty" example:"https://github.com/alisonsandrade/avatar.jpg"`
	JobTitle  string `json:"job_title,omitempty" example:"Tech Lead"`
	Bio       string `json:"bio,omitempty" example:"Desenvolvedor Backend Go"`
}

type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type MessageResponseDTO struct {
	Message string `json:"message" example:"Operação realizada com sucesso"`
}

type ErrorResponseDTO struct {
	Error string `json:"error" example:"Messagem explicativa do error"`
}
