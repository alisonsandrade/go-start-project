// Package auth repository implementa a persistência de dados da aplicação interagindo com o banco./ repository
package auth

import (
	"context"

	"github.com/alisonsandrade/go-start-project/internal/auth/domain"
	usersDomain "github.com/alisonsandrade/go-start-project/internal/users/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *usersDomain.User) error
	FindByEmail(ctx context.Context, email string) (*usersDomain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*usersDomain.User, error)
	GetDefaultRoleID(ctx context.Context) (uuid.UUID, error)
}

type TokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	Delete(ctx context.Context, token string) error
}

type tokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{db}
}

func (r *tokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	return r.db.Create(token).Error
}

func (r *tokenRepository) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var rt domain.RefreshToken
	err := r.db.Where("token = ?", token).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *tokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.Where("user_id = ?", userID).Delete(&domain.RefreshToken{}).Error
}

func (r *tokenRepository) Delete(ctx context.Context, token string) error {
	return r.db.Where("token = ?", token).Delete(&domain.RefreshToken{}).Error
}
