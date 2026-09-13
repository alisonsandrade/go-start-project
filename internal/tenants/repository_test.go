package tenants_test

import (
	"context"
	"testing"

	"github.com/alisonsandrade/go-start-project/internal/tenants"
	"github.com/alisonsandrade/go-start-project/internal/tenants/domain"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTenantTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.Tenant{}))
	return db
}

func TestRepositoryCRUD(t *testing.T) {
	db := setupTenantTestDB(t)
	repo := tenants.NewRepository(db)
	ctx := context.Background()
	tenant := &domain.Tenant{Name: "Example Ltda", Document: "123456789", Settings: []byte(`{"locale":"pt-BR"}`)}

	require.NoError(t, repo.Create(ctx, tenant))
	assert.NotEqual(t, uuid.Nil, tenant.ID)

	byID, err := repo.FindByID(ctx, tenant.ID)
	require.NoError(t, err)
	require.NotNil(t, byID)
	assert.Equal(t, tenant.Document, byID.Document)
	assert.JSONEq(t, `{"locale":"pt-BR"}`, string(byID.Settings))

	byDocument, err := repo.FindByDocument(ctx, tenant.Document)
	require.NoError(t, err)
	require.NotNil(t, byDocument)
	assert.Equal(t, tenant.ID, byDocument.ID)

	tenant.Name = "Updated Ltda"
	tenant.Settings = []byte(`{"locale":"en-US"}`)
	require.NoError(t, repo.Update(ctx, tenant))
	updated, err := repo.FindByID(ctx, tenant.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Ltda", updated.Name)
	assert.JSONEq(t, `{"locale":"en-US"}`, string(updated.Settings))

	second := &domain.Tenant{Name: "Second Ltda", Document: "987654321", Settings: []byte(`{}`)}
	require.NoError(t, repo.Create(ctx, second))
	items, total, err := repo.List(ctx, 1, 0)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 1)

	missing, err := repo.FindByID(ctx, uuid.New())
	require.NoError(t, err)
	assert.Nil(t, missing)
}
