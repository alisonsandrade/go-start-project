// Package auth repository implementa a persistência de dados da aplicação interagindo com o banco./ repository
package auth

import (
	"github.com/alisonsandrade/go-start-project/internal/auth/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TokenRepository interface {
	Create(token *domain.RefreshToken) error
	FindByToken(token string) (*domain.RefreshToken, error)
	DeleteByUserID(userID uuid.UUID) error
	Delete(token string) error
}

type tokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{db}
}

func (r *tokenRepository) Create(token *domain.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *tokenRepository) FindByToken(token string) (*domain.RefreshToken, error) {
	var rt domain.RefreshToken
	err := r.db.Where("token = ?", token).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *tokenRepository) DeleteByUserID(userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&domain.RefreshToken{}).Error
}

func (r *tokenRepository) Delete(token string) error {
	return r.db.Where("token = ?", token).Delete(&domain.RefreshToken{}).Error
}
