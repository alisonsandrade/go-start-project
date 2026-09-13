package tenants

import (
	"context"
	"errors"
	"testing"

	"github.com/alisonsandrade/go-start-project/internal/tenants/domain"
	"github.com/alisonsandrade/go-start-project/pkg/pagination"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockRepository struct{ mock.Mock }

func (m *mockRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	return m.Called(ctx, tenant).Error(0)
}

func (m *mockRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	return m.Called(ctx, tenant).Error(0)
}

func (m *mockRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	args := m.Called(ctx, id)
	if tenant := args.Get(0); tenant != nil {
		return tenant.(*domain.Tenant), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) FindByDocument(ctx context.Context, document string) (*domain.Tenant, error) {
	args := m.Called(ctx, document)
	if tenant := args.Get(0); tenant != nil {
		return tenant.(*domain.Tenant), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockRepository) List(ctx context.Context, limit, offset int) ([]domain.Tenant, int64, error) {
	args := m.Called(ctx, limit, offset)
	var tenants []domain.Tenant
	if value := args.Get(0); value != nil {
		tenants = value.([]domain.Tenant)
	}
	return tenants, args.Get(1).(int64), args.Error(2)
}

func TestServiceCreate(t *testing.T) {
	ctx := context.Background()

	t.Run("creates a tenant", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("FindByDocument", ctx, "123456789").Return(nil, nil).Once()
		repo.On("Create", ctx, mock.AnythingOfType("*domain.Tenant")).Return(nil).Once()

		result, err := NewService(repo).Create(ctx, domain.CreateTenantRequest{
			Name:     "Example Ltda",
			Document: "123.456.789",
			Settings: []byte(`{"locale":"pt-BR"}`),
		})

		require.NoError(t, err)
		assert.Equal(t, "123456789", result.Document)
		assert.JSONEq(t, `{"locale":"pt-BR"}`, string(result.Settings))
		repo.AssertExpectations(t)
	})

	t.Run("rejects a duplicate document", func(t *testing.T) {
		repo := new(mockRepository)
		existing := &domain.Tenant{Document: "123456789"}
		repo.On("FindByDocument", ctx, "123456789").Return(existing, nil).Once()

		result, err := NewService(repo).Create(ctx, domain.CreateTenantRequest{
			Name: "Example Ltda", Document: "123.456.789",
		})

		assert.Nil(t, result)
		assert.ErrorIs(t, err, ErrAlreadyExists)
		repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
	})

	t.Run("propagates repository errors", func(t *testing.T) {
		repo := new(mockRepository)
		repo.On("FindByDocument", ctx, "123456789").Return(nil, errors.New("database down")).Once()

		_, err := NewService(repo).Create(ctx, domain.CreateTenantRequest{
			Name: "Example Ltda", Document: "123.456.789",
		})

		assert.EqualError(t, err, "database down")
	})
}

func TestServiceFindByID(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	repo := new(mockRepository)
	repo.On("FindByID", ctx, id).Return(nil, nil).Once()

	result, err := NewService(repo).FindByID(ctx, id)

	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestServiceUpdate(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	tenant := &domain.Tenant{Document: "123456789", Name: "Old Name"}
	repo := new(mockRepository)
	repo.On("FindByID", ctx, id).Return(tenant, nil).Once()
	repo.On("Update", ctx, tenant).Return(nil).Once()

	result, err := NewService(repo).Update(ctx, id, domain.UpdateTenantRequest{
		Name:     stringPointer("New Name"),
		Settings: []byte(`{"locale":"pt-BR"}`),
	})

	require.NoError(t, err)
	assert.Equal(t, "New Name", result.Name)
	assert.JSONEq(t, `{"locale":"pt-BR"}`, string(result.Settings))
	repo.AssertExpectations(t)
}

func TestServiceList(t *testing.T) {
	ctx := context.Background()
	params := pagination.Params{Page: 2, Limit: 2}
	repo := new(mockRepository)
	repo.On("List", ctx, params.Limit, params.Offset()).Return([]domain.Tenant{{Name: "Example"}}, int64(5), nil).Once()

	result, err := NewService(repo).List(ctx, params)

	require.NoError(t, err)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, 2, result.Meta.CurrentPage)
	assert.Equal(t, int64(5), result.Meta.TotalItems)
	assert.Equal(t, 3, result.Meta.TotalPages)
}

func TestServiceDelete(t *testing.T) {
	ctx := context.Background()
	id := uuid.New()
	repo := new(mockRepository)
	repo.On("FindByID", ctx, id).Return(&domain.Tenant{Document: "123456789"}, nil).Once()
	repo.On("Delete", ctx, id).Return(nil).Once()

	err := NewService(repo).Delete(ctx, id)

	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func stringPointer(value string) *string {
	return &value
}
