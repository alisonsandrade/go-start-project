// Package users
package users

import (
	"context"
	"errors"

	"github.com/alisonsandrade/go-start-project/internal/platform/database"
	rolesDomain "github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/alisonsandrade/go-start-project/internal/users/domain"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	ChangeTenant(ctx context.Context, id, tenantID, roleID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetDefaultRoleID(ctx context.Context) (uuid.UUID, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	List(ctx context.Context, limit, offset int) ([]domain.User, int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Create(user).Error
}

func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Save(user).Error
}

func (r *userRepository) ChangeTenant(ctx context.Context, id, tenantID, roleID uuid.UUID) error {
	claims, ok := ctx.Value(token.ClaimsContextKey).(*token.CustomClaims)
	if !ok || claims == nil || claims.TenantID == uuid.Nil {
		return errors.New("tenant_id é obrigatório")
	}
	if tenantID == claims.TenantID {
		return errors.New("o usuário já pertence a este tenant")
	}

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}

	var tenant struct{ ID uuid.UUID }
	if err := tx.Table("tenants").Where("id = ? AND deleted_at IS NULL", tenantID).First(&tenant).Error; err != nil {
		tx.Rollback()
		return err
	}

	var role rolesDomain.RoleEntity
	if err := tx.Table("roles").Where("id = ? AND tenant_id = ?", roleID, tenantID).First(&role).Error; err != nil {
		tx.Rollback()
		return err
	}

	result := tx.Model(&domain.User{}).
		Where("id = ? AND tenant_id = ?", id, claims.TenantID).
		Updates(map[string]interface{}{"tenant_id": tenantID, "role_id": roleID})
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return gorm.ErrRecordNotFound
	}

	return tx.Commit().Error
}

func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	// Inicia uma transação. Se algo falhar, o banco sofre Rollback.
	tx := r.db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Begin()

	if err := tx.Table("users").Where("id = ?", id).Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Delete(&domain.User{}, "id = ?", id).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

func (r *userRepository) GetDefaultRoleID(ctx context.Context) (uuid.UUID, error) {
	var role rolesDomain.RoleEntity

	err := r.db.
		WithContext(ctx).
		Table("roles").
		Scopes(database.TenantScope(ctx)).
		Select("id").
		Where("name = ?", "USER").
		First(&role).
		Error

	return role.ID, err
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User

	err := r.db.
		WithContext(ctx).
		Scopes(database.TenantScope(ctx)).
		Preload("Role").
		Where("email = ?", email).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var user domain.User
	err := r.db.
		WithContext(ctx).
		Scopes(database.TenantScope(ctx)).
		Preload("Role").
		First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) List(ctx context.Context, limit, offset int) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	if err := r.db.WithContext(ctx).Scopes(database.TenantScope(ctx)).Model(&domain.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.
		WithContext(ctx).
		Scopes(database.TenantScope(ctx)).
		Preload("Role").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&users).Error
	return users, total, err
}
