package domain

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmim Role = "ADMIN"
	RoleUser  Role = "USER"
)

type User struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	Name       string         `gorm:"size:100;not null" json:"name"`
	Email      string         `gorm:"size:150;uniqueIndex;not null" json:"email"`
	Password   string         `gorm:"not null" json:"-"`
	Role       Role           `gorm:"type:varchar(20);default:'USER';not null" json:"role"`
	Phone      string         `gorm:"size:20" json:"phone"`
	AvatarURL  string         `gorm:"size:255" json:"avatar_url"`
	JobTitle   string         `gorm:"size:100" json:"job_title"`
	Bio        string         `gorm:"type:text" json:"bio"`
	IsActive   bool           `gorm:"default:true;not null" json:"is_active"`
	CreateadAt time.Time      `json:"created_at"`
	UpdateadAt time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

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

type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponseDTO struct {
	Token string `json:"token"`
	User  User   `json:"user"`
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
