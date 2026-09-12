package users

import (
	"context"

	rolesDomain "github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/alisonsandrade/go-start-project/internal/users/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type mockUserRepository struct{ mock.Mock }

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *mockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockUserRepository) GetDefaultRoleID(ctx context.Context) (uuid.UUID, error) {
	args := m.Called(ctx)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if user := args.Get(0); user != nil {
		return user.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	if user := args.Get(0); user != nil {
		return user.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserRepository) List(ctx context.Context, limit, offset int) ([]domain.User, int64, error) {
	args := m.Called(ctx, limit, offset)
	var users []domain.User
	if value := args.Get(0); value != nil {
		users = value.([]domain.User)
	}
	return users, args.Get(1).(int64), args.Error(2)
}

type mockRoleRepository struct{ mock.Mock }

func (m *mockRoleRepository) RoleHasPermission(roleID uuid.UUID, code rolesDomain.PermissionCode) (bool, error) {
	args := m.Called(roleID, code)
	return args.Bool(0), args.Error(1)
}

func (m *mockRoleRepository) Create(role *rolesDomain.RoleEntity) error {
	return m.Called(role).Error(0)
}

func (m *mockRoleRepository) GetByID(id uuid.UUID) (*rolesDomain.RoleEntity, error) {
	args := m.Called(id)
	if role := args.Get(0); role != nil {
		return role.(*rolesDomain.RoleEntity), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRoleRepository) GetByName(name string) (*rolesDomain.RoleEntity, error) {
	args := m.Called(name)
	if role := args.Get(0); role != nil {
		return role.(*rolesDomain.RoleEntity), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRoleRepository) List(ctx context.Context, limit, offset int) ([]rolesDomain.RoleEntity, int64, error) {
	args := m.Called(ctx, limit, offset)
	var roles []rolesDomain.RoleEntity
	if value := args.Get(0); value != nil {
		roles = value.([]rolesDomain.RoleEntity)
	}
	return roles, args.Get(1).(int64), args.Error(2)
}

func (m *mockRoleRepository) Update(role *rolesDomain.RoleEntity) error {
	return m.Called(role).Error(0)
}

func (m *mockRoleRepository) Delete(id uuid.UUID) error {
	return m.Called(id).Error(0)
}

func (m *mockRoleRepository) CountPermissionsByIDs(ids []uuid.UUID) (int64, error) {
	args := m.Called(ids)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRoleRepository) ReplacePermissions(roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return m.Called(roleID, permissionIDs).Error(0)
}

var _ UserRepository = (*mockUserRepository)(nil)
var _ rolesDomainRepository = (*mockRoleRepository)(nil)

type rolesDomainRepository interface {
	RoleHasPermission(uuid.UUID, rolesDomain.PermissionCode) (bool, error)
	Create(*rolesDomain.RoleEntity) error
	GetByID(uuid.UUID) (*rolesDomain.RoleEntity, error)
	GetByName(string) (*rolesDomain.RoleEntity, error)
	List(context.Context, int, int) ([]rolesDomain.RoleEntity, int64, error)
	Update(*rolesDomain.RoleEntity) error
	Delete(uuid.UUID) error
	CountPermissionsByIDs([]uuid.UUID) (int64, error)
	ReplacePermissions(uuid.UUID, []uuid.UUID) error
}
