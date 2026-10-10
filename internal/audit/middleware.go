// Package audit
package audit

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/alisonsandrade/go-start-project/internal/auth"
	baseDomain "github.com/alisonsandrade/go-start-project/internal/domain"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/google/uuid"
)

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       bytes.Buffer
}

func (rw *statusResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *statusResponseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

// Middleware registra ações de alteração de estado no banco de dados.
// Recebe jwtSecret para extrair a identidade do usuário de forma independente da ordem do context.
func Middleware(repo AuditRepository, jwtSecret string) func(http.Handler) http.Handler {
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

			var userID *uuid.UUID
			tenantID := uuid.Nil

			// 1. Tenta extrair do context da requisição
			if claims, ok := r.Context().Value(auth.UserClaimsKey).(*token.CustomClaims); ok && claims != nil {
				tenantID = claims.TenantID
				if claims.UserID != uuid.Nil {
					uid := claims.UserID
					userID = &uid
				}
			}

			// 2. Se não encontrou no context, tenta extrair do cabeçalho Bearer Token
			if userID == nil {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
					if claims, err := token.ValidateToken(tokenStr, jwtSecret); err == nil && claims != nil {
						tenantID = claims.TenantID
						uid := claims.UserID
						userID = &uid
					}
				}
			}

			// 3. Caso seja rota de login ou refresh bem-sucedida, o token e o usuário estão no response body
			if (tenantID == uuid.Nil || userID == nil) && (strings.HasSuffix(path, "/login") || strings.HasSuffix(path, "/refresh")) {
				var authResp struct {
					AccessToken string `json:"access_token"`
					User        struct {
						ID       uuid.UUID `json:"id"`
						TenantID uuid.UUID `json:"tenant_id"`
					} `json:"user"`
				}

				if err := json.Unmarshal(rw.body.Bytes(), &authResp); err == nil {
					if authResp.AccessToken != "" {
						if claims, err := token.ValidateToken(authResp.AccessToken, jwtSecret); err == nil && claims != nil {
							tenantID = claims.TenantID
							uid := claims.UserID
							userID = &uid
						}
					}
					if tenantID == uuid.Nil && authResp.User.TenantID != uuid.Nil {
						tenantID = authResp.User.TenantID
					}
					if userID == nil && authResp.User.ID != uuid.Nil {
						uid := authResp.User.ID
						userID = &uid
					}
				}
			}

			// Se a rota não pertencer a nenhum tenant identificado, não persiste auditoria multi-tenant
			if tenantID == uuid.Nil {
				return
			}

			record := &Log{
				BaseModelTenant: baseDomain.BaseModelTenant{TenantID: tenantID},
				UserID:          userID,
				Action:          r.Method,
				Resource:        path,
				IPAddress:       ip,
				UserAgent:       r.UserAgent(),
			}

			// Prepara contexto desacoplado com as claims para satisfazer o TenantScope
			claimsForAudit := &token.CustomClaims{
				TenantID: tenantID,
			}
			if userID != nil {
				claimsForAudit.UserID = *userID
			}

			detachedContext := context.WithoutCancel(r.Context())
			auditCtx := context.WithValue(detachedContext, token.ClaimsContextKey, claimsForAudit)
			auditCtx = context.WithValue(auditCtx, auth.UserClaimsKey, claimsForAudit)

			go func(ctx context.Context, entry *Log) {
				timeoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				defer cancel()

				if err := repo.Create(timeoutCtx, entry); err != nil {
					log.Printf("[AUDIT ERROR] Falha ao persistir auditoria: %v", err)
				}
			}(auditCtx, record)
		})
	}
}
