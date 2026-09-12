package users

import (
	"context"
	"errors"
	"testing"

	rolesDomain "github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/alisonsandrade/go-start-project/internal/users/domain"
	pkgDomain "github.com/alisonsandrade/go-start-project/pkg/domain"
	"github.com/alisonsandrade/go-start-project/pkg/pagination"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newServiceTestUser(t *testing.T, roleID uuid.UUID) *domain.User {
	t.Helper()
	user, err := domain.NewUser("Alice Smith", "alice@example.com", "StrongPass1", roleID)
	require.NoError(t, err)
	user.ID = uuid.New()
	return user
}

func TestUserService_GetUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	t.Run("returns the user", func(t *testing.T) {
		repo := new(mockUserRepository)
		roleRepo := new(mockRoleRepository)
		user := newServiceTestUser(t, uuid.New())
		repo.On("FindByID", ctx, userID).Return(user, nil).Once()

		result, err := NewUserService(repo, roleRepo).GetUser(ctx, userID)

		require.NoError(t, err)
		assert.Same(t, user, result)
	})

	t.Run("returns not found when repository returns nil", func(t *testing.T) {
		repo := new(mockUserRepository)
		repo.On("FindByID", ctx, userID).Return(nil, nil).Once()

		result, err := NewUserService(repo, new(mockRoleRepository)).GetUser(ctx, userID)

		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		repo := new(mockUserRepository)
		repo.On("FindByID", ctx, userID).Return(nil, errors.New("database down")).Once()

		result, err := NewUserService(repo, new(mockRoleRepository)).GetUser(ctx, userID)

		assert.Nil(t, result)
		assert.EqualError(t, err, "database down")
	})
}

func TestUserService_GetDefaultRoleID(t *testing.T) {
	ctx := context.Background()
	roleID := uuid.New()
	repo := new(mockUserRepository)
	repo.On("GetDefaultRoleID", ctx).Return(roleID, nil).Once()

	result, err := NewUserService(repo, new(mockRoleRepository)).GetDefaultRoleID(ctx)

	assert.NoError(t, err)
	assert.Equal(t, roleID, result)
}

func TestUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	roleID := uuid.New()
	user := newServiceTestUser(t, roleID)
	repo := new(mockUserRepository)
	repo.On("FindByID", ctx, userID).Return(user, nil).Once()
	repo.On("Update", ctx, user).Return(nil).Once()
	dto := domain.UpdateUserRequest{
		Name:      "Updated Name",
		Phone:     "555-0100",
		AvatarURL: "https://example.com/avatar.jpg",
		JobTitle:  "Tech Lead",
		Bio:       "Updated bio",
	}

	result, err := NewUserService(repo, new(mockRoleRepository)).UpdateUser(ctx, userID, dto)

	require.NoError(t, err)
	assert.Same(t, user, result)
	assert.Equal(t, dto.Name, user.Name)
	assert.Equal(t, dto.Phone, user.Phone)
	assert.Equal(t, dto.AvatarURL, user.AvatarURL)
	assert.Equal(t, dto.JobTitle, user.JobTitle)
	assert.Equal(t, dto.Bio, user.Bio)
}

func TestUserService_DeleteUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	user := newServiceTestUser(t, uuid.New())
	repo := new(mockUserRepository)
	repo.On("FindByID", ctx, userID).Return(user, nil).Once()
	repo.On("Delete", ctx, userID).Return(nil).Once()

	err := NewUserService(repo, new(mockRoleRepository)).DeleteUser(ctx, userID)

	assert.NoError(t, err)
}

func TestUserService_ListUsers(t *testing.T) {
	ctx := context.Background()
	params := pagination.Params{Page: 2, Limit: 2}
	users := []domain.User{{ID: uuid.New(), Name: "Alice"}, {ID: uuid.New(), Name: "Bob"}}
	repo := new(mockUserRepository)
	repo.On("List", ctx, params.Limit, params.Offset()).Return(users, int64(5), nil).Once()

	result, err := NewUserService(repo, new(mockRoleRepository)).ListUsers(ctx, params)

	require.NoError(t, err)
	assert.Equal(t, users, result.Data)
	assert.Equal(t, 2, result.Meta.CurrentPage)
	assert.Equal(t, int64(5), result.Meta.TotalItems)
	assert.Equal(t, 3, result.Meta.TotalPages)
}

func TestUserService_CreateUserAsAdmin(t *testing.T) {
	ctx := context.Background()
	roleID := uuid.New()
	role := &rolesDomain.RoleEntity{ID: roleID, Name: "EDITOR"}

	t.Run("creates a user with a valid role and email", func(t *testing.T) {
		repo := new(mockUserRepository)
		roleRepo := new(mockRoleRepository)
		roleRepo.On("GetByID", roleID).Return(role, nil).Once()
		repo.On("FindByEmail", ctx, "new@example.com").Return(nil, gorm.ErrRecordNotFound).Once()
		repo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil).Once()
		dto := domain.CreateUserRequest{UserBase: domain.UserBase{Name: "New User", Email: "new@example.com"}, Password: "StrongPass1", RoleID: roleID}

		result, err := NewUserService(repo, roleRepo).CreateUserAsAdmin(ctx, dto)

		require.NoError(t, err)
		assert.Equal(t, dto.Name, result.Name)
		assert.Equal(t, dto.Email, result.Email.String())
		assert.Equal(t, roleID, result.RoleID)
		assert.True(t, result.IsActive)
	})

	t.Run("rejects an invalid role", func(t *testing.T) {
		roleRepo := new(mockRoleRepository)
		roleRepo.On("GetByID", roleID).Return(nil, errors.New("role not found")).Once()

		result, err := NewUserService(new(mockUserRepository), roleRepo).CreateUserAsAdmin(ctx, domain.CreateUserRequest{UserBase: domain.UserBase{Email: "new@example.com"}, Password: "StrongPass1", RoleID: roleID})

		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrInvalidRole)
	})

	t.Run("rejects an existing email", func(t *testing.T) {
		repo := new(mockUserRepository)
		roleRepo := new(mockRoleRepository)
		roleRepo.On("GetByID", roleID).Return(role, nil).Once()
		repo.On("FindByEmail", ctx, "new@example.com").Return(newServiceTestUser(t, roleID), nil).Once()

		result, err := NewUserService(repo, roleRepo).CreateUserAsAdmin(ctx, domain.CreateUserRequest{UserBase: domain.UserBase{Email: "new@example.com"}, Password: "StrongPass1", RoleID: roleID})

		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrEmailAlreadyExists)
	})

	t.Run("rejects an invalid email before checking the role", func(t *testing.T) {
		roleRepo := new(mockRoleRepository)

		result, err := NewUserService(new(mockUserRepository), roleRepo).CreateUserAsAdmin(ctx, domain.CreateUserRequest{UserBase: domain.UserBase{Email: "invalid"}, Password: "StrongPass1", RoleID: roleID})

		assert.Nil(t, result)
		assert.ErrorIs(t, err, pkgDomain.ErrInvalidEmail)
		roleRepo.AssertNotCalled(t, "GetByID", mock.Anything)
	})
}

func TestUserService_UpdateUserAsAdmin(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	user := newServiceTestUser(t, uuid.New())
	newName := "Updated Name"
	isActive := false
	repo := new(mockUserRepository)
	repo.On("FindByID", ctx, userID).Return(user, nil).Once()
	repo.On("Update", ctx, user).Return(nil).Once()

	err := NewUserService(repo, new(mockRoleRepository)).UpdateUserAsAdmin(ctx, userID, domain.AdminUpdateUserRequest{Name: &newName, IsActive: &isActive})

	require.NoError(t, err)
	assert.Equal(t, newName, user.Name)
	assert.False(t, user.IsActive)
}

func TestUserService_SoftDeleteUserAsAdmin(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := new(mockUserRepository)
	repo.On("Delete", ctx, userID).Return(nil).Once()

	err := NewUserService(repo, new(mockRoleRepository)).SoftDeleteUserAsAdmin(ctx, userID)

	assert.NoError(t, err)
}

func TestUserService_SeedDefaultAdmin(t *testing.T) {
	ctx := context.Background()
	role := &rolesDomain.RoleEntity{ID: uuid.New(), Name: "ADMIN"}

	t.Run("creates the default admin", func(t *testing.T) {
		repo := new(mockUserRepository)
		roleRepo := new(mockRoleRepository)
		roleRepo.On("GetByName", "ADMIN").Return(role, nil).Once()
		repo.On("FindByEmail", ctx, "admin@example.com").Return(nil, nil).Once()
		repo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil).Once()

		err := NewUserService(repo, roleRepo).SeedDefaultAdmin(ctx, "Admin", "admin@example.com", "StrongPass1")

		assert.NoError(t, err)
	})

	t.Run("is idempotent when admin already exists", func(t *testing.T) {
		repo := new(mockUserRepository)
		roleRepo := new(mockRoleRepository)
		roleRepo.On("GetByName", "ADMIN").Return(role, nil).Once()
		repo.On("FindByEmail", ctx, "admin@example.com").Return(newServiceTestUser(t, role.ID), nil).Once()

		err := NewUserService(repo, roleRepo).SeedDefaultAdmin(ctx, "Admin", "admin@example.com", "StrongPass1")

		assert.NoError(t, err)
		repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("returns an error when admin role lookup fails", func(t *testing.T) {
		roleRepo := new(mockRoleRepository)
		roleRepo.On("GetByName", "ADMIN").Return(nil, errors.New("role lookup failed")).Once()

		err := NewUserService(new(mockUserRepository), roleRepo).SeedDefaultAdmin(ctx, "Admin", "admin@example.com", "StrongPass1")

		assert.Error(t, err)
	})
}
