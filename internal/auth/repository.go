// Package auth repository implementa a persistência de dados da aplicação interagindo com o banco.
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

type txContextKey struct{}

type TransactionManager interface {
	Transaction(ctx context.Context, fn func(txCtx context.Context) error) error
}

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

func getDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok && tx != nil {
		return tx
	}
	return defaultDB
}

type gormTransactionManager struct {
	db *gorm.DB
}

func NewTransactionManager(db *gorm.DB) TransactionManager {
	return &gormTransactionManager{db: db}
}

func (tm *gormTransactionManager) Transaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txContextKey{}, tx)
		return fn(txCtx)
	})
}

type tokenRepository struct {
	db *gorm.DB
}

func NewTokenRepository(db *gorm.DB) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	db := getDB(ctx, r.db)
	return db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Create(token).Error
}

func (r *tokenRepository) FindByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	db := getDB(ctx, r.db)
	var rt domain.RefreshToken
	err := db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Where("token = ?", token).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

func (r *tokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	db := getDB(ctx, r.db)
	return db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Where("user_id = ?", userID).Delete(&domain.RefreshToken{}).Error
}

func (r *tokenRepository) Delete(ctx context.Context, token string) error {
	db := getDB(ctx, r.db)
	return db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Where("token = ?", token).Delete(&domain.RefreshToken{}).Error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *usersDomain.User) error {
	db := getDB(ctx, r.db)
	return db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, user *usersDomain.User) error {
	db := getDB(ctx, r.db)
	return db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Save(user).Error
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*usersDomain.User, error) {
	db := getDB(ctx, r.db)
	var user usersDomain.User
	err := db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*usersDomain.User, error) {
	db := getDB(ctx, r.db)
	var user usersDomain.User
	err := db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetDefaultRoleID(ctx context.Context) (uuid.UUID, error) {
	db := getDB(ctx, r.db)
	var role rolesDomain.RoleEntity
	err := db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Where("name = ?", "user").First(&role).Error
	if err != nil {
		return uuid.Nil, err
	}
	return role.ID, nil
}
