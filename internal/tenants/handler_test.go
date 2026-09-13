package tenants_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alisonsandrade/go-start-project/internal/auth"
	baseDomain "github.com/alisonsandrade/go-start-project/internal/domain"
	"github.com/alisonsandrade/go-start-project/internal/tenants"
	"github.com/alisonsandrade/go-start-project/internal/tenants/domain"
	"github.com/alisonsandrade/go-start-project/pkg/pagination"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockService struct{ mock.Mock }

func (m *mockService) Create(ctx context.Context, request domain.CreateTenantRequest) (*domain.Tenant, error) {
	args := m.Called(ctx, request)
	if tenant := args.Get(0); tenant != nil {
		return tenant.(*domain.Tenant), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockService) Update(ctx context.Context, id uuid.UUID, request domain.UpdateTenantRequest) (*domain.Tenant, error) {
	args := m.Called(ctx, id, request)
	if tenant := args.Get(0); tenant != nil {
		return tenant.(*domain.Tenant), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockService) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockService) FindByID(ctx context.Context, id uuid.UUID) (*domain.Tenant, error) {
	args := m.Called(ctx, id)
	if tenant := args.Get(0); tenant != nil {
		return tenant.(*domain.Tenant), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockService) List(ctx context.Context, params pagination.Params) (pagination.PageResult[domain.Tenant], error) {
	args := m.Called(ctx, params)
	return args.Get(0).(pagination.PageResult[domain.Tenant]), args.Error(1)
}

func tenantRequest(method, path string, body []byte) *http.Request {
	return httptest.NewRequest(method, path, bytes.NewReader(body))
}

func withTenantRouteID(request *http.Request, id string) *http.Request {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add("id", id)
	return request.WithContext(context.WithValue(request.Context(), chi.RouteCtxKey, routeContext))
}

func TestHandlerCreate(t *testing.T) {
	requestBody := domain.CreateTenantRequest{Name: "Example Ltda", Document: "123456789", Settings: json.RawMessage(`{}`)}
	payload, err := json.Marshal(requestBody)
	require.NoError(t, err)

	t.Run("creates tenant", func(t *testing.T) {
		service := new(mockService)
		tenant := &domain.Tenant{Name: requestBody.Name, Document: requestBody.Document}
		service.On("Create", mock.Anything, requestBody).Return(tenant, nil).Once()

		recorder := httptest.NewRecorder()
		tenants.NewHandler(service).Create(recorder, tenantRequest(http.MethodPost, "/api/tenants", payload))

		assert.Equal(t, http.StatusCreated, recorder.Code)
		assert.Contains(t, recorder.Body.String(), requestBody.Name)
	})

	t.Run("rejects malformed JSON", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		tenants.NewHandler(new(mockService)).Create(recorder, tenantRequest(http.MethodPost, "/api/tenants", []byte("{")))

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("returns conflict for duplicate tenant", func(t *testing.T) {
		service := new(mockService)
		service.On("Create", mock.Anything, requestBody).Return(nil, tenants.ErrAlreadyExists).Once()

		recorder := httptest.NewRecorder()
		tenants.NewHandler(service).Create(recorder, tenantRequest(http.MethodPost, "/api/tenants", payload))

		assert.Equal(t, http.StatusConflict, recorder.Code)
	})
}

func TestHandlerGetByID(t *testing.T) {
	id := uuid.New()

	t.Run("returns tenant", func(t *testing.T) {
		service := new(mockService)
		service.On("FindByID", mock.Anything, id).Return(&domain.Tenant{BaseModel: baseDomain.BaseModel{ID: id}, Name: "Example Ltda"}, nil).Once()
		recorder := httptest.NewRecorder()
		request := withTenantRouteID(tenantRequest(http.MethodGet, "/api/tenants/"+id.String(), nil), id.String())

		tenants.NewHandler(service).GetByID(recorder, request)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "Example Ltda")
	})

	t.Run("returns not found", func(t *testing.T) {
		service := new(mockService)
		service.On("FindByID", mock.Anything, id).Return(nil, tenants.ErrNotFound).Once()
		recorder := httptest.NewRecorder()
		request := withTenantRouteID(tenantRequest(http.MethodGet, "/api/tenants/"+id.String(), nil), id.String())

		tenants.NewHandler(service).GetByID(recorder, request)

		assert.Equal(t, http.StatusNotFound, recorder.Code)
	})

	t.Run("rejects invalid UUID", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		request := withTenantRouteID(tenantRequest(http.MethodGet, "/api/tenants/invalid", nil), "invalid")

		tenants.NewHandler(new(mockService)).GetByID(recorder, request)

		assert.Equal(t, http.StatusBadRequest, recorder.Code)
	})

	t.Run("returns internal error", func(t *testing.T) {
		service := new(mockService)
		service.On("FindByID", mock.Anything, id).Return(nil, errors.New("database down")).Once()
		recorder := httptest.NewRecorder()
		request := withTenantRouteID(tenantRequest(http.MethodGet, "/api/tenants/"+id.String(), nil), id.String())

		tenants.NewHandler(service).GetByID(recorder, request)

		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})
}

func TestHandlerList(t *testing.T) {
	service := new(mockService)
	result := pagination.PageResult[domain.Tenant]{Data: []domain.Tenant{{Name: "Example Ltda"}}}
	service.On("List", mock.Anything, pagination.Params{Page: 2, Limit: 5}).Return(result, nil).Once()
	recorder := httptest.NewRecorder()
	request := tenantRequest(http.MethodGet, "/api/tenants?page=2&limit=5", nil)

	tenants.NewHandler(service).List(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Example Ltda")
}

func TestHandlerUpdate(t *testing.T) {
	id := uuid.New()
	requestBody := domain.UpdateTenantRequest{Name: stringPointer("Updated Ltda"), Settings: json.RawMessage(`{}`)}
	payload, err := json.Marshal(requestBody)
	require.NoError(t, err)

	service := new(mockService)
	service.On("Update", mock.Anything, id, requestBody).Return(&domain.Tenant{Name: "Updated Ltda"}, nil).Once()
	recorder := httptest.NewRecorder()
	request := withTenantRouteID(tenantRequest(http.MethodPut, "/api/tenants/"+id.String(), payload), id.String())

	tenants.NewHandler(service).Update(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Updated Ltda")
}

func TestHandlerGetCurrent(t *testing.T) {
	tenantID := uuid.New()
	service := new(mockService)
	service.On("FindByID", mock.Anything, tenantID).Return(&domain.Tenant{Name: "Current Ltda"}, nil).Once()
	recorder := httptest.NewRecorder()
	request := tenantRequest(http.MethodGet, "/api/tenants/me", nil).WithContext(
		context.WithValue(context.Background(), auth.UserClaimsKey, &token.CustomClaims{TenantID: tenantID}),
	)

	tenants.NewHandler(service).GetCurrent(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "Current Ltda")
}

func TestHandlerDelete(t *testing.T) {
	id := uuid.New()
	service := new(mockService)
	service.On("Delete", mock.Anything, id).Return(nil).Once()
	recorder := httptest.NewRecorder()
	request := withTenantRouteID(tenantRequest(http.MethodDelete, "/api/tenants/"+id.String(), nil), id.String())

	tenants.NewHandler(service).Delete(recorder, request)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func stringPointer(value string) *string {
	return &value
}
