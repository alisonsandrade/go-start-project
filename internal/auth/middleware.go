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
