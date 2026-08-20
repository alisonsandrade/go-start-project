// Package middleware provides HTTP interceptors to manage the application's
// request and response lifecycle.
//
// The functions in this package allow applying reusable cross-cutting concerns,
// such as authentication, panic recovery, structured logging, and CORS handling,
// before the request reaches the final API handlers.
package middleware

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/alisonsandrade/go-start-project/internal/config"
	"github.com/alisonsandrade/go-start-project/internal/domain"
	"github.com/alisonsandrade/go-start-project/internal/repository"
	"github.com/alisonsandrade/go-start-project/pkg/token"
)

type contextKey string

const (
	UserClaimsKey contextKey = "userClaims"
)

func AuthMiddleware(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error": "cabeçalho de autorização ausente"}`, http.StatusUnauthorized)
				return
			}

			// Extração do token
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error": "formato de token inválido. Use Bearer <token>"}`, http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimSpace(parts[1])

			// Validação do token
			claims, err := token.ValidateToken(tokenString, cfg.JWTSecret)
			if err != nil {
				http.Error(w, `{"error": "token inválido ou expirado"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(allowedRoles ...domain.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserClaimsKey).(*token.CustomClaims)
			if !ok || claims == nil {
				http.Error(w, `{"error": "não autenticado"}`, http.StatusUnauthorized)
				return
			}

			// Verificação de roles do usuário
			userRole := domain.Role(claims.Role)
			if !slices.Contains(allowedRoles, userRole) {
				http.Error(w, `{"error": "acesso proibido: permissão insuficiente"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission ensures the authenticated user's role grants the given permission
func RequirePermission(
	roleRepo repository.RoleRepository,
	permission domain.PermissionCode,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserClaimsKey).(*token.CustomClaims)
			if !ok || claims == nil {
				http.Error(w, `{"error": unauthenticated user}`, http.StatusUnauthorized)
				return
			}

			hasPermission, err := roleRepo.RoleHasPermission(claims.Role, permission)
			if err != nil {
				http.Error(w, `{"error": "verified permission error"}`, http.StatusInternalServerError)
				return
			}

			if !hasPermission {
				http.Error(w, `{"error": "unauthorized access: insufficient permission"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
