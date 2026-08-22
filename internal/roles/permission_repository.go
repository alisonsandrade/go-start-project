// Package roles provides persistence operations for the application.
package roles

import (
	"github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PermissionRepository defines persistence operations for permissions.
type PermissionRepository interface {
	GetByIDs(ids []uuid.UUID) ([]domain.Permission, error)
	List() ([]domain.Permission, error)
}

type permissionRepository struct {
	db *gorm.DB
}

// NewPermissionRepository creates a PermissionRepository backed by the given database.
func NewPermissionRepository(db *gorm.DB) PermissionRepository {
	return &permissionRepository{db: db}
}

// GetByIDs returns the permissions matching the provided UUIDs.
func (r *permissionRepository) GetByIDs(ids []uuid.UUID) ([]domain.Permission, error) {
	if len(ids) == 0 {
		return []domain.Permission{}, nil
	}

	var permissions []domain.Permission

	if err := r.db.
		Where("id IN ?", ids).
		Find(&permissions).Error; err != nil {
		return nil, err
	}

	return permissions, nil
}

// List returns all permissions ordered by code.
func (r *permissionRepository) List() ([]domain.Permission, error) {
	var permissions []domain.Permission

	if err := r.db.
		Order("code ASC").
		Find(&permissions).Error; err != nil {
		return nil, err
	}

	return permissions, nil
}
