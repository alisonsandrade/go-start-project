// Package roles
package roles

import (
	"net/http"

	"github.com/alisonsandrade/go-start-project/internal/auth"
	"github.com/alisonsandrade/go-start-project/internal/roles/domain"
	"github.com/alisonsandrade/go-start-project/pkg/token"

	"github.com/alisonsandrade/go-start-project/internal/platform"
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
				platform.ErrorWithCodeJSON(w, http.StatusUnauthorized, platform.CodeUnauthorized, "Usuário não autenticado")
				return
			}

			hasPermission, err := roleRepo.RoleHasPermission(r.Context(), claims.RoleID, permission)
			if err != nil {
				platform.ErrorWithCodeJSON(w, http.StatusInternalServerError, platform.CodeInternalError, "Erro ao verificar a permissão")
				return
			}

			if !hasPermission {
				platform.ErrorWithCodeJSON(w, 401, platform.CodeUnauthorized, "Permissão insuficiente")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
