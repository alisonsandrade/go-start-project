package roles_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alisonsandrade/go-start-project/internal/roles"
	"github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/alisonsandrade/go-start-project/pkg/pagination"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

var ctx = context.Background()

func TestRoleService_Create(t *testing.T) {
	t.Run("creates a role successfully and normalizes its name", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)

		role := &domain.RoleEntity{Name: " admin "} // Nome bagunçado

		// Simulate the database finding nobody with the normalized name "ADMIN".
		repo.On("GetByName", "ADMIN").Return(nil, gorm.ErrRecordNotFound).Once()
		// Simula o banco salvando com sucesso
		repo.On("Create", mock.AnythingOfType("*domain.RoleEntity")).Return(nil).Once()

		created, err := svc.Create(ctx, role)

		assert.NoError(t, err)
		assert.Equal(t, "ADMIN", created.Name) // Ensures the normalization rule ran.
		repo.AssertExpectations(t)
	})

	t.Run("returns an error when role already exists", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)

		role := &domain.RoleEntity{Name: "USER"}

		// Simula que o banco ACHOU um papel com esse nome
		repo.On("GetByName", "USER").Return(&domain.RoleEntity{}, nil).Once()

		_, err := svc.Create(ctx, role)

		assert.ErrorIs(t, err, roles.ErrRoleAlreadyExists)
	})

	t.Run("fails when database returns an error other than not found", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)
		role := &domain.RoleEntity{Name: "DB_DOWN"}

		repo.On("GetByName", "DB_DOWN").Return(nil, errors.New("db error")).Once()
		_, err := svc.Create(ctx, role)
		assert.ErrorContains(t, err, "db error")
	})

	t.Run("fails when saving to database returns an error", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)
		role := &domain.RoleEntity{Name: "SAVE_ERROR"}

		repo.On("GetByName", "SAVE_ERROR").Return(nil, gorm.ErrRecordNotFound).Once()
		repo.On("Create", mock.Anything).Return(errors.New("insert failed")).Once()

		_, err := svc.Create(ctx, role)
		assert.ErrorContains(t, err, "insert failed")
	})
}

func TestRoleService_GetByID(t *testing.T) {
	t.Run("converts gorm error to domain error when role is not found", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)
		id := uuid.New()

		repo.On("GetByID", id).Return(nil, gorm.ErrRecordNotFound).Once()

		_, err := svc.GetByID(ctx, id)

		assert.ErrorIs(t, err, roles.ErrRoleNotFound) // Error translation rule.
	})

	t.Run("returns a generic database error", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)
		id := uuid.New()

		repo.On("GetByID", id).Return(nil, errors.New("db dead")).Once()
		_, err := svc.GetByID(ctx, id)
		assert.ErrorContains(t, err, "db dead")
	})

	t.Run("retrieves a role successfully by ID", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)
		id := uuid.New()
		expectedRole := &domain.RoleEntity{}
		expectedRole.ID = id
		expectedRole.Name = "OK"

		repo.On("GetByID", id).Return(expectedRole, nil).Once()
		res, err := svc.GetByID(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, "OK", res.Name)
	})
}

func TestRoleService_List(t *testing.T) {
	t.Run("returns the paginated result correctly", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)

		// Simulamos uma requisição pedindo a página 2, com limite de 10
		params := pagination.Params{Page: 2, Limit: 10}

		fakeRoles := []domain.RoleEntity{{Name: "ROLE_1"}, {Name: "ROLE_2"}}
		// offset = (page - 1) * limit = (2-1)*10 = 10
		repo.On("List", mock.Anything, 10, 10).Return(fakeRoles, int64(25), nil).Once()

		result, err := svc.List(context.Background(), params)

		assert.NoError(t, err)
		assert.Equal(t, int64(25), result.Meta.TotalItems)
		assert.Equal(t, 3, result.Meta.TotalPages) // 25 itens / 10 = 3 páginas
		assert.Len(t, result.Data, 2)
	})

	t.Run("returns an error when database fails", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)

		repo.On("List", mock.Anything, 10, 0).Return([]domain.RoleEntity{}, int64(0), errors.New("timeout")).Once()
		_, err := svc.List(context.Background(), pagination.Params{Page: 1, Limit: 10})
		assert.ErrorContains(t, err, "timeout")
	})
}

func TestRoleService_Update(t *testing.T) {
	t.Run("prevents editing a system role (IsSystem = true)", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)

		roleToUpdate := &domain.RoleEntity{}
		roleToUpdate.ID = uuid.New()
		roleToUpdate.Name = "ADMIN"

		// Simula que o papel existe no banco, mas tem a flag IsSystem = true
		repo.On("GetByID", roleToUpdate.ID).Return(&domain.RoleEntity{IsSystem: true}, nil).Once()

		_, err := svc.Update(ctx, roleToUpdate)

		assert.ErrorIs(t, err, roles.ErrSystemRoleImmutable) // A regra blindou o sistema!
	})

	t.Run("allows editing regular roles", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)

		roleToUpdate := &domain.RoleEntity{}
		roleToUpdate.ID = uuid.New()
		roleToUpdate.Name = "MANAGER"

		// Retorna IsSystem = false (permitido)
		repo.On("GetByID", roleToUpdate.ID).Return(&domain.RoleEntity{IsSystem: false}, nil).Twice()
		repo.On("Update", roleToUpdate).Return(nil).Once()

		_, err := svc.Update(ctx, roleToUpdate)

		assert.NoError(t, err)
	})

	t.Run("returns an error when role is not found during update", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)
		role := &domain.RoleEntity{}
		role.ID = uuid.New()

		repo.On("GetByID", role.ID).Return(nil, gorm.ErrRecordNotFound).Once()
		_, err := svc.Update(ctx, role)
		assert.ErrorIs(t, err, roles.ErrRoleNotFound)
	})

	t.Run("returns a generic database error during update", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)
		role := &domain.RoleEntity{}
		role.ID = uuid.New()

		repo.On("GetByID", role.ID).Return(&domain.RoleEntity{IsSystem: false}, nil).Twice()
		repo.On("Update", role).Return(errors.New("lock error")).Once()

		_, err := svc.Update(ctx, role)
		assert.ErrorContains(t, err, "lock error")
	})
}

func TestRoleService_Delete(t *testing.T) {
	t.Run("prevents deleting a system role", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)
		id := uuid.New()

		repo.On("GetByID", id).Return(&domain.RoleEntity{IsSystem: true}, nil).Once()

		err := svc.Delete(ctx, id)

		assert.ErrorIs(t, err, roles.ErrSystemRoleImmutable)
	})

	t.Run("allows deleting a regular role", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)
		id := uuid.New()

		repo.On("GetByID", id).Return(&domain.RoleEntity{IsSystem: false}, nil).Once()
		repo.On("Delete", id).Return(nil).Once()

		err := svc.Delete(ctx, id)

		assert.NoError(t, err)
	})

	t.Run("returns an error when role is not found during deletion", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)
		id := uuid.New()

		repo.On("GetByID", id).Return(nil, gorm.ErrRecordNotFound).Once()
		err := svc.Delete(ctx, id)
		assert.ErrorIs(t, err, roles.ErrRoleNotFound)
	})
}

func TestRoleService_ReplacePermissions(t *testing.T) {
	t.Run("rejects association when a permission does not exist in database", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)

		roleID := uuid.New()
		perm1, perm2 := uuid.New(), uuid.New()
		perms := []uuid.UUID{perm1, perm2}

		repo.On("GetByID", roleID).Return(&domain.RoleEntity{IsSystem: false}, nil).Once()

		// O usuário enviou 2 permissões, mas o Mock diz que o banco só achou 1
		repo.On("CountPermissionsByIDs", mock.Anything).Return(int64(1), nil).Once()

		err := svc.ReplacePermissions(ctx, roleID, perms)

		assert.ErrorIs(t, err, roles.ErrInvalidPermissions)
	})

	t.Run("associates permissions successfully (including deduplication)", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)

		roleID := uuid.New()
		perm1 := uuid.New()
		// Simulando envio de permissões repetidas no JSON da API
		perms := []uuid.UUID{perm1, perm1}

		repo.On("GetByID", roleID).Return(&domain.RoleEntity{IsSystem: false}, nil).Once()
		// O banco deve ser consultado apenas para 1 permissão (deduplicada)
		repo.On("CountPermissionsByIDs", []uuid.UUID{perm1}).Return(int64(1), nil).Once()
		repo.On("ReplacePermissions", roleID, []uuid.UUID{perm1}).Return(nil).Once()

		err := svc.ReplacePermissions(ctx, roleID, perms)

		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("returns an error when role is not found", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)
		id := uuid.New()

		repo.On("GetByID", id).Return(nil, gorm.ErrRecordNotFound).Once()
		err := svc.ReplacePermissions(ctx, id, []uuid.UUID{uuid.New()})
		assert.ErrorIs(t, err, roles.ErrRoleNotFound)
	})

	t.Run("prevents replacement on a system role", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)
		id := uuid.New()

		repo.On("GetByID", id).Return(&domain.RoleEntity{IsSystem: true}, nil).Once()
		err := svc.ReplacePermissions(ctx, id, []uuid.UUID{uuid.New()})
		assert.ErrorIs(t, err, roles.ErrSystemRoleImmutable)
	})

	t.Run("performs replacement with an empty array directly in database", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)
		id := uuid.New()

		repo.On("GetByID", id).Return(&domain.RoleEntity{IsSystem: false}, nil).Once()
		repo.On("ReplacePermissions", id, []uuid.UUID{}).Return(nil).Once()

		err := svc.ReplacePermissions(ctx, id, []uuid.UUID{})
		assert.NoError(t, err)
	})

	t.Run("returns an error when database count fails", func(t *testing.T) {
		repo := new(MockRoleRepo)
		svc := roles.NewRoleService(repo)
		id, permID := uuid.New(), uuid.New()

		repo.On("GetByID", id).Return(&domain.RoleEntity{IsSystem: false}, nil).Once()
		repo.On("CountPermissionsByIDs", []uuid.UUID{permID}).Return(int64(0), errors.New("db crash")).Once()

		err := svc.ReplacePermissions(ctx, id, []uuid.UUID{permID})
		assert.ErrorContains(t, err, "db crash")
	})
}
