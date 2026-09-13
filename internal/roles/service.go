// Package roles
package roles

import (
	"context"
	"errors"

	"github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/alisonsandrade/go-start-project/pkg/pagination"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleService defines the business operations for managing roles
type RoleService interface {
	Create(ctx context.Context, role *domain.RoleEntity) (*domain.RoleEntity, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.RoleEntity, error)
	List(ctx context.Context, params pagination.Params) (pagination.PageResult[domain.RoleEntity], error)
	Update(ctx context.Context, role *domain.RoleEntity) (*domain.RoleEntity, error)
	Delete(ctx context.Context, id uuid.UUID) error

	ReplacePermissions(
		ctx context.Context,
		roleID uuid.UUID,
		permissionIDs []uuid.UUID,
	) error
}

// roleService is the concrete implementaion. It depends on the repository
// INTERFACE, never on the concrete Gorm struct - that's what makes it testable.
type roleService struct {
	repo RoleRepository
}

// NewRoleService wires a RoleService with its dependency
func NewRoleService(repo RoleRepository) RoleService {
	return &roleService{repo: repo}
}

// Create creates a new role ensuring name uniqueness.
func (s *roleService) Create(ctx context.Context, role *domain.RoleEntity) (*domain.RoleEntity, error) {
	role.Name = domain.NormalizeRoleName(role.Name)

	_, err := s.repo.GetByName(ctx, role.Name)

	if err == nil {
		return nil, ErrRoleAlreadyExists
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err := s.repo.Create(ctx, role); err != nil {
		return nil, err
	}

	return role, nil
}

// GetByID returns a role by its UUID, translating persistence errors into
// errors. A missing record becomes ErrRoleNotFound
func (s *roleService) GetByID(ctx context.Context, id uuid.UUID) (*domain.RoleEntity, error) {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}

	return role, nil
}

// List returns all roles
func (s *roleService) List(ctx context.Context, params pagination.Params) (pagination.PageResult[domain.RoleEntity], error) {
	roles, total, err := s.repo.List(ctx, params.Limit, params.Offset())
	if err != nil {
		return pagination.PageResult[domain.RoleEntity]{}, err
	}
	return pagination.NewPageResult(roles, total, params), nil
}

// Update updates an existing role.
func (s *roleService) Update(ctx context.Context, role *domain.RoleEntity) (*domain.RoleEntity, error) {
	existingRole, err := s.repo.GetByID(ctx, role.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}

		return nil, err
	}

	if existingRole.IsSystem {
		return nil, ErrSystemRoleImmutable
	}

	if err := s.repo.Update(ctx, role); err != nil {
		return nil, err
	}

	return s.repo.GetByID(ctx, role.ID)
}

// Delete removes a role.
func (s *roleService) Delete(ctx context.Context, id uuid.UUID) error {
	role, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}

		return err
	}

	if role.IsSystem {
		return ErrSystemRoleImmutable
	}

	return s.repo.Delete(ctx, id)
}

// ReplacePermissions replaces all permissions assigned to a role.
func (s *roleService) ReplacePermissions(
	ctx context.Context,
	roleID uuid.UUID,
	permissionIDs []uuid.UUID,
) error {
	role, err := s.repo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRoleNotFound
		}

		return err
	}

	if role.IsSystem {
		return ErrSystemRoleImmutable
	}

	uniquePermissionIDs := uniqueUUIDs(permissionIDs)

	if len(uniquePermissionIDs) == 0 {
		return s.repo.ReplacePermissions(ctx, roleID, uniquePermissionIDs)
	}

	count, err := s.repo.CountPermissionsByIDs(ctx, uniquePermissionIDs)
	if err != nil {
		return err
	}

	if count != int64(len(uniquePermissionIDs)) {
		return ErrInvalidPermissions
	}

	return s.repo.ReplacePermissions(
		ctx,
		roleID,
		uniquePermissionIDs,
	)
}

// uniqueUUIDs removes duplicate UUIDs while preserving their original order.
func uniqueUUIDs(ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	unique := make([]uuid.UUID, 0, len(ids))

	for _, id := range ids {
		if _, exists := seen[id]; exists {
			continue
		}

		seen[id] = struct{}{}
		unique = append(unique, id)
	}

	return unique
}
