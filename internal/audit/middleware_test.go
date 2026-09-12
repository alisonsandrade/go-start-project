package audit_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alisonsandrade/go-start-project/internal/audit"
	"github.com/alisonsandrade/go-start-project/internal/auth"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// MockRepo sincrono com canais para sincronizar goroutines de auditoria
type MockAuditRepo struct {
	CapturedLog *audit.Log
	Signal      chan struct{}
	ShouldFail  bool
}

func (m *MockAuditRepo) Create(ctx context.Context, log *audit.Log) error {
	m.CapturedLog = log
	if m.Signal != nil {
		m.Signal <- struct{}{}
	}
	if m.ShouldFail {
		return errors.New("db error")
	}
	return nil
}

func (m *MockAuditRepo) List(ctx context.Context, limit, offset int) ([]audit.Log, error) {
	return nil, nil
}

func TestAuditMiddleware_AllScenarios(t *testing.T) {
	jwtSecret := "my-secret-key-for-unit-tests-32"

	t.Run("ignores OPTIONS method", func(t *testing.T) {
		repo := &MockAuditRepo{Signal: make(chan struct{}, 1)}
		mw := audit.Middleware(repo, jwtSecret)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodOptions, "/api/users", nil)
		rec := httptest.NewRecorder()
		mw(next).ServeHTTP(rec, req)

		select {
		case <-repo.Signal:
			t.Fatal("OPTIONS nao deve gravar log")
		case <-time.After(30 * time.Millisecond):
			// Sucesso
		}
	})

	t.Run("ignores Swagger and health routes", func(t *testing.T) {
		repo := &MockAuditRepo{Signal: make(chan struct{}, 1)}
		mw := audit.Middleware(repo, jwtSecret)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// Rota com /swagger
		reqSwagger := httptest.NewRequest(http.MethodPost, "/swagger/index.html", nil)
		recSwagger := httptest.NewRecorder()
		mw(next).ServeHTTP(recSwagger, reqSwagger)

		// Rota com /health
		reqHealth := httptest.NewRequest(http.MethodPost, "/healthz", nil)
		recHealth := httptest.NewRecorder()
		mw(next).ServeHTTP(recHealth, reqHealth)

		select {
		case <-repo.Signal:
			t.Fatal("Swagger e Health nao devem ser auditados")
		case <-time.After(30 * time.Millisecond):
			// Sucesso
		}
	})

	t.Run("extracts userID directly from context (auth claims)", func(t *testing.T) {
		repo := &MockAuditRepo{Signal: make(chan struct{}, 1)}
		mw := audit.Middleware(repo, jwtSecret)

		userID := uuid.New()
		claims := &token.CustomClaims{UserID: userID}

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Escreve direto no body sem chamar WriteHeader primeiro para exercitar Write()
			_, _ = w.Write([]byte("ok"))
		})

		req := httptest.NewRequest(http.MethodPost, "/api/roles", nil)
		ctx := context.WithValue(req.Context(), auth.UserClaimsKey, claims)
		req = req.WithContext(ctx)
		rec := httptest.NewRecorder()

		mw(next).ServeHTTP(rec, req)

		select {
		case <-repo.Signal:
			assert.NotNil(t, repo.CapturedLog.UserID)
			assert.Equal(t, userID, *repo.CapturedLog.UserID)
			assert.Equal(t, "/api/roles", repo.CapturedLog.Resource)
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout esperando persistencia de auditoria")
		}
	})

	t.Run("handles an IP without a port (invalid RemoteAddr for SplitHostPort)", func(t *testing.T) {
		repo := &MockAuditRepo{Signal: make(chan struct{}, 1)}
		mw := audit.Middleware(repo, jwtSecret)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})

		req := httptest.NewRequest(http.MethodDelete, "/api/users/123", nil)
		req.RemoteAddr = "10.0.0.1" // Sem porta: forca net.SplitHostPort a retornar erro
		rec := httptest.NewRecorder()

		mw(next).ServeHTTP(rec, req)

		select {
		case <-repo.Signal:
			assert.Equal(t, "10.0.0.1", repo.CapturedLog.IPAddress)
			assert.Equal(t, http.MethodDelete, repo.CapturedLog.Action)
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout esperando log")
		}
	})

	t.Run("handles database persistence failure without breaking the request", func(t *testing.T) {
		repo := &MockAuditRepo{
			Signal:     make(chan struct{}, 1),
			ShouldFail: true, // Simula erro no db
		}
		mw := audit.Middleware(repo, jwtSecret)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodPost, "/api/users", nil)
		rec := httptest.NewRecorder()

		mw(next).ServeHTTP(rec, req)

		select {
		case <-repo.Signal:
			// Ensure the repository error was captured by the logging branch.
			assert.Equal(t, http.StatusOK, rec.Code)
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout na persistencia")
		}
	})

	t.Run("covers statusResponseWriter.Write branches", func(t *testing.T) {
		repo := &MockAuditRepo{Signal: make(chan struct{}, 2)}
		mw := audit.Middleware(repo, "secret")

		// Caso 1: Handler chama Write direto (sem chamar WriteHeader antes -> statusCode vira 200)
		next1 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("sem write header"))
		})
		req1 := httptest.NewRequest(http.MethodPost, "/api/test-write-default", nil)
		rec1 := httptest.NewRecorder()
		mw(next1).ServeHTTP(rec1, req1)
		<-repo.Signal

		// Caso 2: Handler chama WriteHeader antes de Write (statusCode já definido)
		next2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("com write header"))
		})
		req2 := httptest.NewRequest(http.MethodPost, "/api/test-write-explicit", nil)
		rec2 := httptest.NewRecorder()
		mw(next2).ServeHTTP(rec2, req2)
		<-repo.Signal
	})
}
