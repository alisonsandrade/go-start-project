// Package roles
package roles

import (
	"net/http"

	"github.com/alisonsandrade/go-start-project/internal/auth"
	"github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/alisonsandrade/go-start-project/pkg/token"
)

// RequirePermission ensures the authenticated user's role grants the given permission
func RequirePermission(
	roleRepo RoleRepository,
	permission domain.PermissionCode,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(auth.UserClaimsKey).(*token.CustomClaims)
			if !ok || claims == nil {
				http.Error(w, `{"error": unauthenticated user}`, http.StatusUnauthorized)
				return
			}

			hasPermission, err := roleRepo.RoleHasPermission(claims.RoleID, permission)
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
