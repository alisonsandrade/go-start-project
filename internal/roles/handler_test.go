package roles_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alisonsandrade/go-start-project/internal/config"
	"github.com/alisonsandrade/go-start-project/internal/roles"
	"github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/alisonsandrade/go-start-project/pkg/pagination"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRoleHandler_GetRoleByID(t *testing.T) {
	mockSvc := new(MockRoleService)
	handler := roles.NewRoleHandler(mockSvc)

	t.Run("returns 400 when UUID is invalid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/roles/invalid-uuid", nil)

		// Inject Chi router path parameters.
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "invalid-uuid")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		handler.GetRoleByID(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("returns 404 when role is not found", func(t *testing.T) {
		id := uuid.New()
		req := httptest.NewRequest(http.MethodGet, "/api/roles/"+id.String(), nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", id.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		mockSvc.On("GetByID", id).Return(nil, roles.ErrRoleNotFound).Once()

		rr := httptest.NewRecorder()
		handler.GetRoleByID(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestRoleHandler_CreateRole(t *testing.T) {
	mockSvc := new(MockRoleService)
	handler := roles.NewRoleHandler(mockSvc)

	t.Run("returns 400 for malformed JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/roles", bytes.NewReader([]byte("{invalid")))
		rr := httptest.NewRecorder()

		handler.CreateRole(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("returns 409 when role already exists", func(t *testing.T) {
		dto := domain.CreateRoleRequest{Name: "ADMIN"}
		payload, _ := json.Marshal(dto)

		mockSvc.On("Create", mock.Anything).Return(nil, roles.ErrRoleAlreadyExists).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/roles", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		handler.CreateRole(rr, req)
		assert.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("returns 201 when role is created successfully", func(t *testing.T) {
		dto := domain.CreateRoleRequest{Name: "MANAGER"}
		payload, _ := json.Marshal(dto)

		mockSvc.On("Create", mock.Anything).Return(&domain.RoleEntity{Name: "MANAGER"}, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "/api/roles", bytes.NewReader(payload))
		rr := httptest.NewRecorder()

		handler.CreateRole(rr, req)
		assert.Equal(t, http.StatusCreated, rr.Code)
	})
}

func TestRoleHandler_DeleteRole(t *testing.T) {
	mockSvc := new(MockRoleService)
	handler := roles.NewRoleHandler(mockSvc)

	t.Run("returns 409 when deleting an immutable role (IsSystem)", func(t *testing.T) {
		id := uuid.New()
		req := httptest.NewRequest(http.MethodDelete, "/api/roles/"+id.String(), nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", id.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		mockSvc.On("Delete", id).Return(roles.ErrSystemRoleImmutable).Once()

		rr := httptest.NewRecorder()
		handler.DeleteRole(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
	})

	t.Run("returns 200 when role is deleted successfully", func(t *testing.T) {
		id := uuid.New()
		req := httptest.NewRequest(http.MethodDelete, "/api/roles/"+id.String(), nil)

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", id.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		mockSvc.On("Delete", id).Return(nil).Once()

		rr := httptest.NewRecorder()
		handler.DeleteRole(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestRoleHandler_ListRoles(t *testing.T) {
	mockSvc := new(MockRoleService)
	handler := roles.NewRoleHandler(mockSvc)

	t.Run("returns 500 when service fails", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/roles", nil)
		rr := httptest.NewRecorder()

		mockSvc.On("List", mock.Anything, mock.Anything).Return(pagination.PageResult[domain.RoleEntity]{}, errors.New("db error")).Once()

		handler.ListRoles(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("returns 200 with a list of roles", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/roles?page=1&limit=10", nil)
		rr := httptest.NewRecorder()

		mockResult := pagination.PageResult[domain.RoleEntity]{
			Data: []domain.RoleEntity{{Name: "ADMIN"}},
		}
		mockSvc.On("List", mock.Anything, mock.Anything).Return(mockResult, nil).Once()

		handler.ListRoles(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestRoleHandler_UpdateRole(t *testing.T) {
	mockSvc := new(MockRoleService)
	handler := roles.NewRoleHandler(mockSvc)
	roleID := uuid.New()

	t.Run("returns 400 when UUID is invalid", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/roles/uuid-invalido", nil)
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", "uuid-invalido")
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		handler.UpdateRole(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("returns 404 when role does not exist", func(t *testing.T) {
		dto := domain.UpdateRoleRequest{Name: "MANAGER"}
		payload, _ := json.Marshal(dto)
		req := httptest.NewRequest(http.MethodPut, "/api/roles/"+roleID.String(), bytes.NewReader(payload))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", roleID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		mockSvc.On("Update", mock.Anything).Return(nil, roles.ErrRoleNotFound).Once()

		rr := httptest.NewRecorder()
		handler.UpdateRole(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("returns 200 on success", func(t *testing.T) {
		dto := domain.UpdateRoleRequest{Name: "MANAGER"}
		payload, _ := json.Marshal(dto)
		req := httptest.NewRequest(http.MethodPut, "/api/roles/"+roleID.String(), bytes.NewReader(payload))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", roleID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		mockSvc.On("Update", mock.Anything).Return(&domain.RoleEntity{Name: "MANAGER"}, nil).Once()

		rr := httptest.NewRecorder()
		handler.UpdateRole(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

func TestRoleHandler_ReplacePermissions(t *testing.T) {
	mockSvc := new(MockRoleService)
	handler := roles.NewRoleHandler(mockSvc)
	roleID := uuid.New()

	t.Run("returns 400 for malformed JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/roles/"+roleID.String()+"/permissions", bytes.NewReader([]byte("{bad-json")))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", roleID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		rr := httptest.NewRecorder()
		handler.ReplacePermissions(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("returns 400 when permissions are invalid", func(t *testing.T) {
		dto := domain.ReplacePermissionsRequest{PermissionIDs: []uuid.UUID{uuid.New()}}
		payload, _ := json.Marshal(dto)
		req := httptest.NewRequest(http.MethodPut, "/api/roles/"+roleID.String()+"/permissions", bytes.NewReader(payload))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", roleID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		mockSvc.On("ReplacePermissions", roleID, dto.PermissionIDs).Return(roles.ErrInvalidPermissions).Once()

		rr := httptest.NewRecorder()
		handler.ReplacePermissions(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("returns 200 on success", func(t *testing.T) {
		dto := domain.ReplacePermissionsRequest{PermissionIDs: []uuid.UUID{uuid.New()}}
		payload, _ := json.Marshal(dto)
		req := httptest.NewRequest(http.MethodPut, "/api/roles/"+roleID.String()+"/permissions", bytes.NewReader(payload))

		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", roleID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		mockSvc.On("ReplacePermissions", roleID, dto.PermissionIDs).Return(nil).Once()

		rr := httptest.NewRecorder()
		handler.ReplacePermissions(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}

// Ensures route setup does not panic and returns a valid router.
func TestRoleHandler_Routes(t *testing.T) {
	mockSvc := new(MockRoleService)
	handler := roles.NewRoleHandler(mockSvc)
	mockRepo := new(MockRoleRepo)

	// Use an empty (or mocked) config to avoid a nil pointer in the auth middleware.
	router := handler.Routes(&config.Config{JWTSecret: "test"}, mockRepo)
	assert.NotNil(t, router)
}
