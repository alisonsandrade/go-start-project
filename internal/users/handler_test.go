package users

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
	"github.com/alisonsandrade/go-start-project/internal/users/domain"
	"github.com/alisonsandrade/go-start-project/pkg/pagination"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockUserService struct{ mock.Mock }

func (m *mockUserService) GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if user := args.Get(0); user != nil {
		return user.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserService) GetDefaultRoleID(ctx context.Context) (uuid.UUID, error) {
	args := m.Called(ctx)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockUserService) UpdateUser(ctx context.Context, userID uuid.UUID, dto domain.UpdateUserRequest) (*domain.User, error) {
	args := m.Called(ctx, userID, dto)
	if user := args.Get(0); user != nil {
		return user.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserService) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *mockUserService) ListUsers(ctx context.Context, params pagination.Params) (pagination.PageResult[domain.User], error) {
	args := m.Called(ctx, params)
	return args.Get(0).(pagination.PageResult[domain.User]), args.Error(1)
}

func (m *mockUserService) CreateUserAsAdmin(ctx context.Context, dto domain.CreateUserRequest) (*domain.User, error) {
	args := m.Called(ctx, dto)
	if user := args.Get(0); user != nil {
		return user.(*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockUserService) UpdateUserAsAdmin(ctx context.Context, userID uuid.UUID, dto domain.AdminUpdateUserRequest) error {
	return m.Called(ctx, userID, dto).Error(0)
}

func (m *mockUserService) SoftDeleteUserAsAdmin(ctx context.Context, userID uuid.UUID) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *mockUserService) SeedDefaultAdmin(ctx context.Context, name, email, password string) error {
	return m.Called(ctx, name, email, password).Error(0)
}

func userHandlerRequest(method, path string, body []byte, userID uuid.UUID) *http.Request {
	req := httptest.NewRequest(method, path, bytes.NewReader(body))
	claims := &token.CustomClaims{UserID: userID}
	return req.WithContext(context.WithValue(req.Context(), auth.UserClaimsKey, claims))
}

func withRouteID(req *http.Request, id string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestUserHandler_GetUser(t *testing.T) {
	userID := uuid.New()
	user := &domain.User{BaseModelTenant: baseDomain.BaseModelTenant{BaseModel: baseDomain.BaseModel{ID: userID}}, Name: "Alice"}
	service := new(mockUserService)
	service.On("GetUser", mock.Anything, userID).Return(user, nil).Once()

	rr := httptest.NewRecorder()
	NewUserHandler(service).GetUser(rr, userHandlerRequest(http.MethodGet, "/api/users/me", nil, userID))

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"name":"Alice"`)
}

func TestUserHandler_ListUsers(t *testing.T) {
	service := new(mockUserService)
	result := pagination.PageResult[domain.User]{Data: []domain.User{{Name: "Alice"}}}
	service.On("ListUsers", mock.Anything, pagination.Params{Page: 2, Limit: 5}).Return(result, nil).Once()
	req := userHandlerRequest(http.MethodGet, "/api/users?page=2&limit=5", nil, uuid.New())
	rr := httptest.NewRecorder()

	NewUserHandler(service).ListUsers(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), `"current_page":0`)
}

func TestUserHandler_CreateUser(t *testing.T) {
	userID := uuid.New()
	roleID := uuid.New()
	dto := domain.CreateUserRequest{UserBase: domain.UserBase{Name: "Alice", Email: "alice@example.com"}, Password: "StrongPass1", RoleID: roleID}
	user := &domain.User{BaseModelTenant: baseDomain.BaseModelTenant{BaseModel: baseDomain.BaseModel{ID: userID}}, Name: "Alice", RoleID: roleID}

	t.Run("returns unauthorized without claims", func(t *testing.T) {
		service := new(mockUserService)
		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString(`{}`))
		rr := httptest.NewRecorder()

		NewUserHandler(service).CreateUser(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("returns bad request for malformed JSON", func(t *testing.T) {
		service := new(mockUserService)
		rr := httptest.NewRecorder()

		NewUserHandler(service).CreateUser(rr, userHandlerRequest(http.MethodPost, "/api/users", []byte("{invalid"), userID))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("creates a user", func(t *testing.T) {
		service := new(mockUserService)
		service.On("CreateUserAsAdmin", mock.Anything, dto).Return(user, nil).Once()
		payload, err := json.Marshal(dto)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		NewUserHandler(service).CreateUser(rr, userHandlerRequest(http.MethodPost, "/api/users", payload, userID))

		assert.Equal(t, http.StatusCreated, rr.Code)
	})
}

func TestUserHandler_UpdateAndDeleteUser(t *testing.T) {
	userID := uuid.New()
	dto := domain.UpdateUserRequest{Name: "Updated Alice"}
	user := &domain.User{BaseModelTenant: baseDomain.BaseModelTenant{BaseModel: baseDomain.BaseModel{ID: userID}}, Name: dto.Name}

	t.Run("updates the authenticated user", func(t *testing.T) {
		service := new(mockUserService)
		service.On("UpdateUser", mock.Anything, userID, dto).Return(user, nil).Once()
		payload, err := json.Marshal(dto)
		require.NoError(t, err)
		rr := httptest.NewRecorder()

		NewUserHandler(service).UpdateUser(rr, userHandlerRequest(http.MethodPut, "/api/users/me", payload, userID))

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("returns bad request for malformed update JSON", func(t *testing.T) {
		rr := httptest.NewRecorder()

		NewUserHandler(new(mockUserService)).UpdateUser(rr, userHandlerRequest(http.MethodPut, "/api/users/me", []byte("{invalid"), userID))

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("deletes the authenticated user", func(t *testing.T) {
		service := new(mockUserService)
		service.On("DeleteUser", mock.Anything, userID).Return(nil).Once()
		rr := httptest.NewRecorder()

		NewUserHandler(service).DeleteUser(rr, userHandlerRequest(http.MethodDelete, "/api/users/me", nil, userID))

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestUserHandler_AdminEndpoints(t *testing.T) {
	userID := uuid.New()
	user := &domain.User{BaseModelTenant: baseDomain.BaseModelTenant{BaseModel: baseDomain.BaseModel{ID: userID}}, Name: "Alice"}

	t.Run("rejects an invalid user ID", func(t *testing.T) {
		rr := httptest.NewRecorder()
		req := withRouteID(userHandlerRequest(http.MethodGet, "/api/users/invalid", nil, userID), "invalid")

		NewUserHandler(new(mockUserService)).GetUserByID(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("gets a user by ID", func(t *testing.T) {
		service := new(mockUserService)
		service.On("GetUser", mock.Anything, userID).Return(user, nil).Once()
		rr := httptest.NewRecorder()
		req := withRouteID(userHandlerRequest(http.MethodGet, "/api/users/"+userID.String(), nil, uuid.New()), userID.String())

		NewUserHandler(service).GetUserByID(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("soft deletes a user by ID", func(t *testing.T) {
		service := new(mockUserService)
		service.On("SoftDeleteUserAsAdmin", mock.Anything, userID).Return(nil).Once()
		rr := httptest.NewRecorder()
		req := withRouteID(userHandlerRequest(http.MethodDelete, "/api/users/"+userID.String(), nil, uuid.New()), userID.String())

		NewUserHandler(service).SoftDeleteUserAsAdmin(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("maps service failures to internal server error", func(t *testing.T) {
		service := new(mockUserService)
		service.On("GetUser", mock.Anything, userID).Return(nil, errors.New("database down")).Once()
		rr := httptest.NewRecorder()
		req := withRouteID(userHandlerRequest(http.MethodGet, "/api/users/"+userID.String(), nil, uuid.New()), userID.String())

		NewUserHandler(service).GetUserByID(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})
}
