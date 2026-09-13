// Package roles
package roles

import (
	"context"

	"github.com/alisonsandrade/go-start-project/internal/platform/database"
	"github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleRepository defines persistence operations for roles and permissions.
type RoleRepository interface {
	RoleHasPermission(ctx context.Context, roleID uuid.UUID, code domain.PermissionCode) (bool, error)

	Create(ctx context.Context, role *domain.RoleEntity) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.RoleEntity, error)
	GetByName(ctx context.Context, name string) (*domain.RoleEntity, error)
	List(ctx context.Context, limit, offset int) ([]domain.RoleEntity, int64, error)
	Update(ctx context.Context, role *domain.RoleEntity) error
	Delete(ctx context.Context, id uuid.UUID) error

	CountPermissionsByIDs(ctx context.Context, ids []uuid.UUID) (int64, error)
	ReplacePermissions(
		ctx context.Context,
		roleID uuid.UUID,
		permissionIDs []uuid.UUID,
	) error
}

type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository creates a RoleRepository backed by the give database
func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

// Create persists a new role and its many-to-many permission assciations
func (r *roleRepository) Create(ctx context.Context, role *domain.RoleEntity) error {
	role.IsSystem = false // safety clause
	return r.db.
		WithContext(ctx).
		Scopes(database.TenantScope(ctx)).
		Create(role).Error
}

// GetByID returns a role by its UUID with its permissions eagerly loaded
func (r *roleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.RoleEntity, error) {
	var role domain.RoleEntity

	err := r.db.
		WithContext(ctx).
		Preload("Permissions").
		Scopes(database.TenantScope(ctx)).
		First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return &role, nil
}

// GetByName returns a role by its unique name with its permissions eagerly loaded.
// This mirrors GetByID but keys off the human-readable name (e.g. "ADMIN", "USER"),
// which is how seeds and the enforcement layer usually reference roles.
func (r *roleRepository) GetByName(ctx context.Context, name string) (*domain.RoleEntity, error) {
	var role domain.RoleEntity

	err := r.db.
		WithContext(ctx).
		Preload("Permissions").
		Scopes(database.TenantScope(ctx)).
		First(&role, "name = ?", name).Error
	if err != nil {
		return nil, err
	}

	return &role, nil
}

// List returns all roles with their permissions eagerly loaded.
func (r *roleRepository) List(ctx context.Context, limit, offset int) ([]domain.RoleEntity, int64, error) {
	var roles []domain.RoleEntity
	var total int64

	if err := r.db.
		WithContext(ctx).
		Scopes(database.TenantScope(ctx)).
		Model(&domain.RoleEntity{}).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := r.db.
		WithContext(ctx).
		Scopes(database.TenantScope(ctx)).
		Preload("Permissions").
		Order("name ASC").
		Limit(limit).
		Offset(offset).
		Find(&roles).Error
	if err != nil {
		return nil, 0, err
	}

	return roles, total, err
}

// Update persists changes to an existing role.
// Note: this saves the role's own columns. Managing the permission
// associations (attach/detach) is done explicitly in a dedicated method,
// not as a side effect of Update.
func (r *roleRepository) Update(ctx context.Context, role *domain.RoleEntity) error {
	return r.db.
		WithContext(ctx).
		Scopes(database.TenantScope(ctx)).
		Model(role).
		Select("name", "description").
		Updates(role).Error
}

// Delete removes a role by its UUID.
// This repository method is intentionally "dumb": it does not check whether
// the role is a system role. That business rule lives in the service layer.
func (r *roleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	var role domain.RoleEntity
	return r.db.
		WithContext(ctx).
		Scopes(database.TenantScope(ctx)).
		Delete(&role, "id = ?", id).Error
}

// RoleHasPermission reports wheter the role (by name) grants the permission (by code)
func (r *roleRepository) RoleHasPermission(ctx context.Context, roleID uuid.UUID, code domain.PermissionCode) (bool, error) {
	var count int64

	err := r.db.
		WithContext(ctx).
		Table("role_permissions AS rp").
		Joins("JOIN roles r ON r.id = rp.role_id").
		Joins("JOIN permissions p ON p.id = rp.permission_id").
		Where("r.tenant_id = ? AND r.id = ? AND p.code = ?", tenantID(ctx), roleID, string(code)).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ReplacePermissions replaces all permissions assigned to a role.
func (r *roleRepository) ReplacePermissions(
	ctx context.Context,
	roleID uuid.UUID,
	permissionIDs []uuid.UUID,
) error {
	var role domain.RoleEntity

	if err := r.db.
		WithContext(ctx).
		Scopes(database.TenantScope(ctx)).
		First(&role, "id = ?", roleID).
		Error; err != nil {
		return err
	}

	permissions := make([]domain.Permission, 0, len(permissionIDs))

	for _, id := range permissionIDs {
		permission := domain.Permission{}
		permission.ID = id
		permissions = append(permissions, permission)
	}

	return r.db.
		WithContext(ctx).
		Model(&role).
		Association("Permissions").
		Replace(permissions)
}

func tenantID(ctx context.Context) uuid.UUID {
	claims, ok := ctx.Value(token.ClaimsContextKey).(*token.CustomClaims)
	if !ok || claims == nil {
		return uuid.Nil
	}
	return claims.TenantID
}

// CountPermissionsByIDs returns how many permissions exist for the provided IDs.
func (r *roleRepository) CountPermissionsByIDs(
	ctx context.Context,
	ids []uuid.UUID,
) (int64, error) {
	var count int64

	err := r.db.
		WithContext(ctx).
		Model(&domain.Permission{}).
		Where("id IN ?", ids).
		Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}
