package roles_test

import (
	"context"

	"github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/alisonsandrade/go-start-project/pkg/pagination"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockRoleRepo
type MockRoleRepo struct{ mock.Mock }

func (m *MockRoleRepo) RoleHasPermission(ctx context.Context, roleID uuid.UUID, code domain.PermissionCode) (bool, error) {
	args := m.Called(roleID, code)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoleRepo) Create(ctx context.Context, role *domain.RoleEntity) error {
	return m.Called(role).Error(0)
}

func (m *MockRoleRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.RoleEntity, error) {
	args := m.Called(id)
	if r := args.Get(0); r != nil {
		return r.(*domain.RoleEntity), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRoleRepo) GetByName(ctx context.Context, name string) (*domain.RoleEntity, error) {
	args := m.Called(name)
	if r := args.Get(0); r != nil {
		return r.(*domain.RoleEntity), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRoleRepo) List(ctx context.Context, limit, offset int) ([]domain.RoleEntity, int64, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]domain.RoleEntity), args.Get(1).(int64), args.Error(2)
}

func (m *MockRoleRepo) Update(ctx context.Context, role *domain.RoleEntity) error {
	return m.Called(role).Error(0)
}

func (m *MockRoleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func (m *MockRoleRepo) CountPermissionsByIDs(ctx context.Context, ids []uuid.UUID) (int64, error) {
	args := m.Called(ids)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRoleRepo) ReplacePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return m.Called(roleID, permissionIDs).Error(0)
}

// MockRoleService
type MockRoleService struct{ mock.Mock }

func (m *MockRoleService) Create(ctx context.Context, role *domain.RoleEntity) (*domain.RoleEntity, error) {
	args := m.Called(role)
	if r := args.Get(0); r != nil {
		return r.(*domain.RoleEntity), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRoleService) GetByID(ctx context.Context, id uuid.UUID) (*domain.RoleEntity, error) {
	args := m.Called(id)
	if r := args.Get(0); r != nil {
		return r.(*domain.RoleEntity), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRoleService) List(ctx context.Context, params pagination.Params) (pagination.PageResult[domain.RoleEntity], error) {
	args := m.Called(ctx, params)
	return args.Get(0).(pagination.PageResult[domain.RoleEntity]), args.Error(1)
}

func (m *MockRoleService) Update(ctx context.Context, role *domain.RoleEntity) (*domain.RoleEntity, error) {
	args := m.Called(role)
	if r := args.Get(0); r != nil {
		return r.(*domain.RoleEntity), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRoleService) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func (m *MockRoleService) ReplacePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return m.Called(roleID, permissionIDs).Error(0)
}
