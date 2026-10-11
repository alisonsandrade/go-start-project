// Package auth middleware provides HTTP interceptors to manage the application's
// request and response lifecycle.
//
// The functions in this package allow applying reusable cross-cutting concerns,
// such as authentication, panic recovery, structured logging, and CORS handling,
// before the request reaches the final API handlers.
package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/alisonsandrade/go-start-project/internal/config"
	"github.com/alisonsandrade/go-start-project/internal/platform"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/google/uuid"
)

const (
	UserClaimsKey = token.ClaimsContextKey
)

func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				platform.ErrorWithCodeJSON(w, http.StatusUnauthorized, "UNAUTHORIZED", "cabeçalho de autorização ausente")
				return
			}

			// Extração do token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				platform.ErrorWithCodeJSON(w, http.StatusUnauthorized, "INVALID_TOKEN_FORMAT", "formato de token inválido. Use Bearer <token>")
				return
			}

			tokenString := strings.TrimSpace(parts[1])

			// Validação do token
			claims, err := token.ValidateToken(tokenString, cfg.JWTSecret)
			if err != nil {
				platform.ErrorWithCodeJSON(w, http.StatusUnauthorized, "INVALID_TOKEN", "token inválido ou expirado")
				return
			}

			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TenantContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := token.DefaultTenantID
		if rawTenantID := r.Header.Get("X-Tenant-ID"); rawTenantID != "" {
			parsedTenantID, err := uuid.Parse(rawTenantID)
			if err != nil || parsedTenantID == uuid.Nil {
				platform.ErrorWithCodeJSON(w, http.StatusBadRequest, "INVALID_TENANT_ID", "tenant_id inválido")
				return
			}
			tenantID = parsedTenantID
		}

		claims := &token.CustomClaims{TenantID: tenantID}
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
