package roles_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alisonsandrade/go-start-project/internal/auth"
	"github.com/alisonsandrade/go-start-project/internal/roles"
	"github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRequirePermission_Middleware(t *testing.T) {
	permRequired := domain.PermissionCode("READ_USERS")

	t.Run("returns 401 when context has no claims", func(t *testing.T) {
		repo := new(MockRoleRepo)
		mw := roles.RequirePermission(repo, permRequired)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		req := httptest.NewRequest(http.MethodGet, "/api/admin", nil)
		rr := httptest.NewRecorder()

		mw(next).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("returns 500 when database permission check fails", func(t *testing.T) {
		repo := new(MockRoleRepo)
		mw := roles.RequirePermission(repo, permRequired)
		roleID := uuid.New()

		repo.On("RoleHasPermission", roleID, permRequired).Return(false, errors.New("db is offline")).Once()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		req := httptest.NewRequest(http.MethodGet, "/api/admin", nil)
		ctx := context.WithValue(req.Context(), auth.UserClaimsKey, &token.CustomClaims{RoleID: roleID})
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mw(next).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusInternalServerError, rr.Code)
	})

	t.Run("returns 403 when user lacks the required permission", func(t *testing.T) {
		repo := new(MockRoleRepo)
		mw := roles.RequirePermission(repo, permRequired)
		roleID := uuid.New()

		repo.On("RoleHasPermission", roleID, permRequired).Return(false, nil).Once()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		req := httptest.NewRequest(http.MethodGet, "/api/admin", nil)
		ctx := context.WithValue(req.Context(), auth.UserClaimsKey, &token.CustomClaims{RoleID: roleID})
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mw(next).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusForbidden, rr.Code)
	})

	t.Run("forwards the request (200) when user has the permission", func(t *testing.T) {
		repo := new(MockRoleRepo)
		mw := roles.RequirePermission(repo, permRequired)
		roleID := uuid.New()

		repo.On("RoleHasPermission", roleID, permRequired).Return(true, nil).Once()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		req := httptest.NewRequest(http.MethodGet, "/api/admin", nil)
		ctx := context.WithValue(req.Context(), auth.UserClaimsKey, &token.CustomClaims{RoleID: roleID})
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mw(next).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusOK, rr.Code)
	})
}
