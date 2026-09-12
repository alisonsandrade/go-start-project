package users

import (
	"context"
	"testing"
	"time"

	rolesDomain "github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/alisonsandrade/go-start-project/internal/users/domain"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupUserTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:user_test_"+uuid.NewString()+"?mode=memory&cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE roles (
			id uuid PRIMARY KEY,
			name text NOT NULL,
			description text,
			is_system numeric NOT NULL DEFAULT false,
			created_at datetime,
			updated_at datetime
		)
	`).Error)
	require.NoError(t, db.AutoMigrate(&domain.User{}))
	return db
}

func createRepositoryTestUser(t *testing.T, roleID uuid.UUID, name, email string) *domain.User {
	t.Helper()
	user, err := domain.NewUser(name, email, "StrongPass1", roleID)
	require.NoError(t, err)
	return user
}

func TestUserRepository_CRUD(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)
	ctx := context.Background()
	role := &rolesDomain.RoleEntity{ID: uuid.New(), Name: "USER"}
	require.NoError(t, db.Create(role).Error)
	user := createRepositoryTestUser(t, role.ID, "Alice Smith", "alice@example.com")

	t.Run("creates and finds a user by email and ID", func(t *testing.T) {
		require.NoError(t, repo.Create(ctx, user))

		byEmail, err := repo.FindByEmail(ctx, user.Email.String())
		require.NoError(t, err)
		require.NotNil(t, byEmail)
		assert.Equal(t, user.ID, byEmail.ID)
		assert.Equal(t, role.ID, byEmail.Role.ID)

		byID, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, byID)
		assert.Equal(t, user.Email.String(), byID.Email.String())
	})

	t.Run("updates a user", func(t *testing.T) {
		user.Name = "Updated Name"
		require.NoError(t, repo.Update(ctx, user))

		updated, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updated.Name)
	})

	t.Run("returns nil when user does not exist", func(t *testing.T) {
		missing, err := repo.FindByID(ctx, uuid.New())
		assert.NoError(t, err)
		assert.Nil(t, missing)

		missing, err = repo.FindByEmail(ctx, "missing@example.com")
		assert.NoError(t, err)
		assert.Nil(t, missing)
	})
}

func TestUserRepository_GetDefaultRoleID(t *testing.T) {
	db := setupUserTestDB(t)
	roleID := uuid.New()
	require.NoError(t, db.Create(&rolesDomain.RoleEntity{ID: roleID, Name: "USER"}).Error)
	repo := NewUserRepository(db)

	result, err := repo.GetDefaultRoleID(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, roleID, result)
}

func TestUserRepository_List(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)
	role := &rolesDomain.RoleEntity{ID: uuid.New(), Name: "USER"}
	require.NoError(t, db.Create(role).Error)
	first := createRepositoryTestUser(t, role.ID, "First User", "first@example.com")
	first.CreatedAt = time.Now().UTC().Add(-time.Hour)
	second := createRepositoryTestUser(t, role.ID, "Second User", "second@example.com")
	second.CreatedAt = time.Now().UTC()
	require.NoError(t, repo.Create(context.Background(), first))
	require.NoError(t, repo.Create(context.Background(), second))

	users, total, err := repo.List(context.Background(), 1, 0)

	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	require.Len(t, users, 1)
	assert.Equal(t, second.ID, users[0].ID)
	assert.Equal(t, role.ID, users[0].Role.ID)
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupUserTestDB(t)
	repo := NewUserRepository(db)
	role := &rolesDomain.RoleEntity{ID: uuid.New(), Name: "USER"}
	require.NoError(t, db.Create(role).Error)
	user := createRepositoryTestUser(t, role.ID, "Delete User", "delete@example.com")
	require.NoError(t, repo.Create(context.Background(), user))

	err := repo.Delete(context.Background(), user.ID)

	require.NoError(t, err)
	found, err := repo.FindByID(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Nil(t, found)

	var deleted domain.User
	require.NoError(t, db.Unscoped().First(&deleted, user.ID).Error)
	assert.False(t, deleted.IsActive)
	assert.True(t, deleted.DeletedAt.Valid)
}
