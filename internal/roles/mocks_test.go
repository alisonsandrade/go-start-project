package roles_test

import (
	"context"

	"github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/alisonsandrade/go-start-project/pkg/pagination"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// MockRoleRepo centralizado para ser usado por todos os testes
type MockRoleRepo struct{ mock.Mock }

func (m *MockRoleRepo) RoleHasPermission(roleID uuid.UUID, code domain.PermissionCode) (bool, error) {
	args := m.Called(roleID, code)
	return args.Bool(0), args.Error(1)
}
func (m *MockRoleRepo) Create(role *domain.RoleEntity) error { return m.Called(role).Error(0) }
func (m *MockRoleRepo) GetByID(id uuid.UUID) (*domain.RoleEntity, error) {
	args := m.Called(id)
	if r := args.Get(0); r != nil {
		return r.(*domain.RoleEntity), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockRoleRepo) GetByName(name string) (*domain.RoleEntity, error) {
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
func (m *MockRoleRepo) Update(role *domain.RoleEntity) error { return m.Called(role).Error(0) }
func (m *MockRoleRepo) Delete(id uuid.UUID) error            { return m.Called(id).Error(0) }

// A função que o Go está sentindo falta precisa estar declarada exatamente assim:
func (m *MockRoleRepo) CountPermissionsByIDs(ids []uuid.UUID) (int64, error) {
	args := m.Called(ids)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRoleRepo) ReplacePermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return m.Called(roleID, permissionIDs).Error(0)
}

/*
* Mock Service
 */

// MockRoleService simula a camada de serviço para o Handler
type MockRoleService struct{ mock.Mock }

func (m *MockRoleService) Create(role *domain.RoleEntity) (*domain.RoleEntity, error) {
	args := m.Called(role)
	if r := args.Get(0); r != nil {
		return r.(*domain.RoleEntity), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockRoleService) GetByID(id uuid.UUID) (*domain.RoleEntity, error) {
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
func (m *MockRoleService) Update(role *domain.RoleEntity) (*domain.RoleEntity, error) {
	args := m.Called(role)
	if r := args.Get(0); r != nil {
		return r.(*domain.RoleEntity), args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockRoleService) Delete(id uuid.UUID) error {
	return m.Called(id).Error(0)
}
func (m *MockRoleService) ReplacePermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return m.Called(roleID, permissionIDs).Error(0)
}
