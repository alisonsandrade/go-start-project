// Package auth repository implementa a persistência de dados da aplicação interagindo com o banco./ repository
package auth

import (
	"context"

	"github.com/alisonsandrade/go-start-project/internal/auth/domain"
	"github.com/alisonsandrade/go-start-project/internal/platform/database"
	rolesDomain "github.com/alisonsandrade/go-start-project/internal/roles/domain"
	usersDomain "github.com/alisonsandrade/go-start-project/internal/users/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *usersDomain.User) error
	Update(ctx context.Context, user *usersDomain.User) error
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

type UnitOfWork interface {
	Do(ctx context.Context, fn func(userRepo UserRepository, tokenRepo TokenRepository) error) error
}

type tokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	return r.db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Create(token).Error
}

func (r *tokenRepository) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var rt domain.RefreshToken
	err := r.db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Where("token = ?", token).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *tokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Where("user_id = ?", userID).Delete(&domain.RefreshToken{}).Error
}

func (r *tokenRepository) Delete(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Where("token = ?", token).Delete(&domain.RefreshToken{}).Error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *usersDomain.User) error {
	return r.db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, user *usersDomain.User) error {
	return r.db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Save(user).Error
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*usersDomain.User, error) {
	var user usersDomain.User
	err := r.db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*usersDomain.User, error) {
	var user usersDomain.User
	err := r.db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetDefaultRoleID(ctx context.Context) (uuid.UUID, error) {
	var role rolesDomain.RoleEntity
	err := r.db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Where("name = ?", "user").First(&role).Error
	if err != nil {
		return uuid.Nil, err
	}
	return role.ID, nil
}

type gormUnitOfWork struct {
	db *gorm.DB
}

func NewUnitOfWork(db *gorm.DB) UnitOfWork {
	return &gormUnitOfWork{db: db}
}

func (u *gormUnitOfWork) Do(ctx context.Context, fn func(userRepo UserRepository, tokenRepo TokenRepository) error) error {
	return u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txUserRepo := &userRepository{db: tx}
		txTokenRepo := &tokenRepository{db: tx}
		return fn(txUserRepo, txTokenRepo)
	})
}
