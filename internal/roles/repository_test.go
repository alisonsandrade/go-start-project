package roles_test

import (
	"context"
	"testing"

	"github.com/alisonsandrade/go-start-project/internal/roles"
	"github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupRolesTestDB creates tables in in-memory SQLite without PostgreSQL-specific syntax.
func setupRolesTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	err = db.Exec(`
		CREATE TABLE roles (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			is_system BOOLEAN DEFAULT FALSE,
			created_at DATETIME,
			updated_at DATETIME
			,deleted_at DATETIME
		);
		CREATE TABLE permissions (
			id TEXT PRIMARY KEY,
			code TEXT NOT NULL UNIQUE,
			description TEXT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME
		);
		CREATE TABLE role_permissions (
			role_id TEXT NOT NULL,
			permission_id TEXT NOT NULL,
			tenant_id TEXT,
			PRIMARY KEY (role_id, permission_id)
		);
	`).Error
	assert.NoError(t, err)

	return db
}

func TestRoleRepository_Integration(t *testing.T) {
	db := setupRolesTestDB(t)
	repo := roles.NewRoleRepository(db)
	ctx := context.WithValue(context.Background(), token.ClaimsContextKey, &token.CustomClaims{TenantID: token.DefaultTenantID})

	roleID := uuid.New()
	role := &domain.RoleEntity{}
	role.ID = roleID
	role.Name = "MANAGER"
	role.Description = "Manager Role"

	t.Run("Create persists a new role", func(t *testing.T) {
		err := repo.Create(ctx, role)
		require.NoError(t, err)
	})

	t.Run("GetByID and GetByName return the created role", func(t *testing.T) {
		// Test GetByID.
		foundByID, err := repo.GetByID(ctx, roleID)
		require.NoError(t, err)
		assert.Equal(t, "MANAGER", foundByID.Name)

		// Test GetByName.
		foundByName, err := repo.GetByName(ctx, "MANAGER")
		require.NoError(t, err)
		assert.Equal(t, roleID, foundByName.ID)
	})

	t.Run("Update changes the role data", func(t *testing.T) {
		role.Description = "Updated Description"
		err := repo.Update(ctx, role)
		assert.NoError(t, err)

		updated, _ := repo.GetByID(ctx, roleID)
		assert.Equal(t, "Updated Description", updated.Description)
	})

	t.Run("List returns paginated roles", func(t *testing.T) {
		rolesList, total, err := repo.List(ctx, 10, 0)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, rolesList, 1)
	})

	t.Run("ReplacePermissions e RoleHasPermission", func(t *testing.T) {
		// 1. Inserimos uma permissão real no banco para testar o vínculo
		permID := uuid.New()
		err := db.Exec("INSERT INTO permissions (id, code, description) VALUES (?, ?, ?)", permID, "READ_USERS", "Can read users").Error
		assert.NoError(t, err)

		// 2. Vinculamos a permissão à role usando o repositório
		err = repo.ReplacePermissions(ctx, roleID, []uuid.UUID{permID})
		assert.NoError(t, err)

		// 3. Verify that the repository confirms the role has the permission.
		hasPerm, err := repo.RoleHasPermission(ctx, roleID, domain.PermissionCode("READ_USERS"))
		assert.NoError(t, err)
		assert.True(t, hasPerm)

		// 4. Verificamos CountPermissionsByIDs
		count, err := repo.CountPermissionsByIDs(ctx, []uuid.UUID{permID})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("Delete removes the role", func(t *testing.T) {
		err := repo.Delete(ctx, roleID)
		assert.NoError(t, err)

		_, err = repo.GetByID(ctx, roleID)
		assert.Error(t, err) // Deve dar erro de NotFound do GORM
	})
}

func TestPermissionRepository_Integration(t *testing.T) {
	db := setupRolesTestDB(t)
	repo := roles.NewPermissionRepository(db)
	ctx := context.Background()

	permID1 := uuid.New()
	permID2 := uuid.New()

	db.Exec("INSERT INTO permissions (id, code, description) VALUES (?, ?, ?)", permID1, "READ_ROLES", "Lê roles")
	db.Exec("INSERT INTO permissions (id, code, description) VALUES (?, ?, ?)", permID2, "WRITE_ROLES", "Escreve roles")

	t.Run("GetByIDs returns the correct permissions", func(t *testing.T) {
		perms, err := repo.GetByIDs(ctx, []uuid.UUID{permID1, permID2})
		assert.NoError(t, err)
		assert.Len(t, perms, 2)
	})

	t.Run("GetByIDs returns empty when the ID list is empty", func(t *testing.T) {
		perms, err := repo.GetByIDs(ctx, []uuid.UUID{})
		assert.NoError(t, err)
		assert.Empty(t, perms)
	})

	t.Run("List returns all permissions", func(t *testing.T) {
		perms, err := repo.List(ctx)
		assert.NoError(t, err)
		assert.Len(t, perms, 2)
	})
}
