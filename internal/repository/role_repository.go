// Package repository
package repository

import (
	"github.com/alisonsandrade/go-start-project/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleRepository defines persistence operations for roles and permissions.
type RoleRepository interface {
	RoleHasPermission(roleName string, code domain.PermissionCode) (bool, error)

	Create(role *domain.RoleEntity) error
	GetByID(id uuid.UUID) (*domain.RoleEntity, error)
	GetByName(name string) (*domain.RoleEntity, error)
	List() ([]domain.RoleEntity, error)
	Update(role *domain.RoleEntity) error
	Delete(id uuid.UUID) error

	CountPermissionsByIDs(ids []uuid.UUID) (int64, error)
	ReplacePermissions(
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
func (r *roleRepository) Create(role *domain.RoleEntity) error {
	role.IsSystem = false // clasule security
	return r.db.Create(role).Error
}

// GetByID returns a role by its UUID with its permissions eagerly loaded
func (r *roleRepository) GetByID(id uuid.UUID) (*domain.RoleEntity, error) {
	var role domain.RoleEntity

	err := r.db.
		Preload("Permissions").
		First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return &role, nil
}

// GetByName returns a role by its unique name with its permissions eagerly loaded.
// This mirrors GetByID but keys off the human-readable name (e.g. "ADMIN", "USER"),
// which is how seeds and the enforcement layer usually reference roles.
func (r *roleRepository) GetByName(name string) (*domain.RoleEntity, error) {
	var role domain.RoleEntity

	err := r.db.
		Preload("Permissions").
		First(&role, "name = ?", name).Error
	if err != nil {
		return nil, err
	}

	return &role, nil
}

// List returns all roles with their permissions eagerly loaded.
func (r *roleRepository) List() ([]domain.RoleEntity, error) {
	var roles []domain.RoleEntity

	err := r.db.
		Preload("Permissions").
		Order("name ASC").
		Find(&roles).Error
	if err != nil {
		return nil, err
	}

	return roles, nil
}

// Update persists changes to an existing role.
// Note: this saves the role's own columns. Managing the permission
// associations (attach/detach) is done explicitly in a dedicated method,
// not as a side effect of Update.
func (r *roleRepository) Update(role *domain.RoleEntity) error {
	return r.db.Model(role).
		Select("name", "description").
		Updates(role).Error
}

// Delete removes a role by its UUID.
// This repository method is intentionally "dumb": it does not check whether
// the role is a system role. That business rule lives in the service layer.
func (r *roleRepository) Delete(id uuid.UUID) error {
	var role domain.RoleEntity
	return r.db.Delete(&role, "id = ?", id).Error
}

// RoleHasPermission reports wheter the role (by name) grants the permission (by code)
func (r *roleRepository) RoleHasPermission(roleName string, code domain.PermissionCode) (bool, error) {
	var count int64

	err := r.db.
		Table("role_permissions AS rp").
		Joins("JOIN roles r ON r.id = rp.role_id").
		Joins("JOIN permissions p ON p.id = rp.permission_id").
		Where("r.name = ? AND p.code = ?", roleName, string(code)).
		Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ReplacePermissions replaces all permissions assigned to a role.
func (r *roleRepository) ReplacePermissions(
	roleID uuid.UUID,
	permissionIDs []uuid.UUID,
) error {
	var role domain.RoleEntity

	if err := r.db.
		First(&role, "id = ?", roleID).
		Error; err != nil {
		return err
	}

	permissions := make([]domain.Permission, 0, len(permissionIDs))

	for _, id := range permissionIDs {
		permissions = append(
			permissions,
			domain.Permission{
				ID: id,
			},
		)
	}

	return r.db.
		Model(&role).
		Association("Permissions").
		Replace(permissions)
}

// CountPermissionsByIDs returns how many permissions exist for the provided IDs.
func (r *roleRepository) CountPermissionsByIDs(
	ids []uuid.UUID,
) (int64, error) {
	var count int64

	err := r.db.
		Model(&domain.Permission{}).
		Where("id IN ?", ids).
		Count(&count).Error
	if err != nil {
		return 0, err
	}

	return count, nil
}
