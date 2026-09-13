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

func (m *mockRoleRepository) RoleHasPermission(ctx context.Context, roleID uuid.UUID, code rolesDomain.PermissionCode) (bool, error) {
	args := m.Called(ctx, roleID, code)
	return args.Bool(0), args.Error(1)
}

func (m *mockRoleRepository) Create(ctx context.Context, role *rolesDomain.RoleEntity) error {
	return m.Called(ctx, role).Error(0)
}

func (m *mockRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*rolesDomain.RoleEntity, error) {
	args := m.Called(ctx, id)
	if role := args.Get(0); role != nil {
		return role.(*rolesDomain.RoleEntity), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRoleRepository) GetByName(ctx context.Context, name string) (*rolesDomain.RoleEntity, error) {
	args := m.Called(ctx, name)
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

func (m *mockRoleRepository) Update(ctx context.Context, role *rolesDomain.RoleEntity) error {
	return m.Called(ctx, role).Error(0)
}

func (m *mockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockRoleRepository) CountPermissionsByIDs(ctx context.Context, ids []uuid.UUID) (int64, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).(int64), args.Error(1)
}

func (m *mockRoleRepository) ReplacePermissions(ctx context.Context, roleID uuid.UUID, permissionIDs []uuid.UUID) error {
	return m.Called(ctx, roleID, permissionIDs).Error(0)
}

var _ UserRepository = (*mockUserRepository)(nil)
var _ rolesDomainRepository = (*mockRoleRepository)(nil)

type rolesDomainRepository interface {
	RoleHasPermission(context.Context, uuid.UUID, rolesDomain.PermissionCode) (bool, error)
	Create(context.Context, *rolesDomain.RoleEntity) error
	GetByID(context.Context, uuid.UUID) (*rolesDomain.RoleEntity, error)
	GetByName(context.Context, string) (*rolesDomain.RoleEntity, error)
	List(context.Context, int, int) ([]rolesDomain.RoleEntity, int64, error)
	Update(context.Context, *rolesDomain.RoleEntity) error
	Delete(context.Context, uuid.UUID) error
	CountPermissionsByIDs(context.Context, []uuid.UUID) (int64, error)
	ReplacePermissions(context.Context, uuid.UUID, []uuid.UUID) error
}
