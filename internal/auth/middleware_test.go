package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alisonsandrade/go-start-project/internal/auth"
	"github.com/alisonsandrade/go-start-project/internal/config"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	cfg := &config.Config{JWTSecret: "minha-chave-secreta-de-testes-32-bits!"}
	mw := auth.AuthMiddleware(cfg)

	t.Run("retorna 401 se cabecalho Authorization estiver ausente", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
		rr := httptest.NewRecorder()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		mw(next).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("retorna 401 se formato do token nao comecar com Bearer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
		req.Header.Set("Authorization", "Token meu-jwt-invalido")
		rr := httptest.NewRecorder()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		mw(next).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("retorna 401 se token for invalido ou expirado", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer token-completamente-invalido")
		rr := httptest.NewRecorder()

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		mw(next).ServeHTTP(rr, req)
		assert.Equal(t, http.StatusUnauthorized, rr.Code)
	})

	t.Run("injeta claims no contexto e segue quando token for valido", func(t *testing.T) {
		userID := uuid.New()
		tokenStr, err := token.GenerateToken(userID, "user@test.com", uuid.New(), cfg.JWTSecret, 1)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenStr)
		rr := httptest.NewRecorder()

		var extractedClaims *token.CustomClaims
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if claims, ok := r.Context().Value(auth.UserClaimsKey).(*token.CustomClaims); ok {
				extractedClaims = claims
			}
			w.WriteHeader(http.StatusOK)
		})

		mw(next).ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NotNil(t, extractedClaims)
		assert.Equal(t, userID, extractedClaims.UserID)
	})
}
