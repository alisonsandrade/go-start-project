// Package audit
package audit

import (
	"context"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/alisonsandrade/go-start-project/internal/auth"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/google/uuid"
)

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *statusResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *statusResponseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

// Middleware registra ações de alteração de estado no banco de dados.
// Recebe jwtSecret para extrair a identidade do usuário de forma independente da ordem do context.
func Middleware(repo Repository, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			path := r.URL.Path
			if r.Method == http.MethodGet || strings.Contains(path, "/swagger") || strings.Contains(path, "/health") {
				next.ServeHTTP(w, r)
				return
			}

			rw := &statusResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rw, r)

			// Não audita se a requisição resultou em erro de cliente ou servidor
			if rw.statusCode >= 400 {
				return
			}

			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			// 1. Tenta extrair do context
			var userID *uuid.UUID
			if claims, ok := r.Context().Value(auth.UserClaimsKey).(*token.CustomClaims); ok && claims != nil {
				uid := claims.UserID
				userID = &uid
			}

			// 2. Se o context for nulo devido à cópia rasa do Go, extrai direto do Bearer Token
			if userID == nil {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
					if claims, err := token.ValidateToken(tokenStr, jwtSecret); err == nil && claims != nil {
						uid := claims.UserID
						userID = &uid
					}
				}
			}

			record := &Log{
				UserID:    userID,
				Action:    r.Method,
				Resource:  path,
				IPAddress: ip,
				UserAgent: r.UserAgent(),
				CreatedAt: time.Now().UTC(),
			}

			go func(entry *Log) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				if err := repo.Create(ctx, entry); err != nil {
					log.Printf("[AUDIT ERROR] Falha ao persistir auditoria: %v", err)
				}
			}(record)
		})
	}
}
