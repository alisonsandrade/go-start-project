// Package users
package users

import (
	"errors"

	"github.com/alisonsandrade/go-start-project/internal/users/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(user *domain.User) error
	Update(user *domain.User) error
	Delete(ui uuid.UUID) error
	FindByEmail(email string) (*domain.User, error)
	FindByID(ui uuid.UUID) (*domain.User, error)
	ListAll() ([]domain.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uuid.UUID) error {
	// Inicia uma transação. Se algo falhar, o banco sofre Rollback.
	tx := r.db.Begin()

	if err := tx.Model(&domain.User{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&domain.User{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User

	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) ListAll() ([]domain.User, error) {
	var users []domain.User
	err := r.db.Find(&users).Error
	return users, err
}
