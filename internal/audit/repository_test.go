// Package Audit
package audit_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alisonsandrade/go-start-project/internal/audit"
	"github.com/alisonsandrade/go-start-project/internal/config"
	"github.com/alisonsandrade/go-start-project/pkg/token"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// setupAuditTestDB configures an in-memory database with a SQLite-compatible table.
func setupAuditTestDB(t *testing.T) (*gorm.DB, *config.Config) {
	db, errDB := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, errDB)

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatal("falha ao carregar configs: %w", err)
	}

	// Explicitly create the table without PostgreSQL annotations.
	err = db.Exec(`
		CREATE TABLE audit_logs (
			id TEXT PRIMARY KEY,
			user_id TEXT,
			action TEXT NOT NULL,
			resource TEXT NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			created_at DATETIME
		);
	`).Error
	assert.NoError(t, err)

	return db, cfg
}

func TestAuditRepository_Create(t *testing.T) {
	db, _ := setupAuditTestDB(t)
	// Usa o construtor real declarado em audit.go
	repo := audit.NewAuditRepository(db)
	ctx := context.Background()

	uid := uuid.New()

	tests := []struct {
		name        string
		log         audit.Log
		expectError bool
	}{
		{
			name: "create a new log when save a new user",
			log: audit.Log{
				ID:        uuid.New(),
				UserID:    &uid,
				Action:    "POST",
				Resource:  "/api/users",
				IPAddress: "127.0.0.1",
				UserAgent: "Go-Test-Agent",
				CreatedAt: time.Now().UTC().Add(-1 * time.Minute),
			},
			expectError: false,
		},
		{
			name: "create a new log when update roles (sem user_id)",
			log: audit.Log{
				ID:        uuid.New(),
				UserID:    nil,
				Action:    "PUT",
				Resource:  "/api/roles",
				IPAddress: "192.168.0.1",
				UserAgent: "Go-Test-Agent",
				CreatedAt: time.Now().UTC(),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Executa a função Create de audit.go
			err := repo.Create(ctx, &tt.log)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuditRepository_List(t *testing.T) {
	db, cfg := setupAuditTestDB(t)
	repo := audit.NewAuditRepository(db)
	ctx := context.Background()

	// Insert two records to test ordered retrieval.
	log1 := &audit.Log{
		ID:        uuid.New(),
		Action:    "POST",
		Resource:  "/api/users",
		CreatedAt: time.Now().UTC().Add(-10 * time.Minute),
	}
	log2 := &audit.Log{
		ID:        uuid.New(),
		Action:    "DELETE",
		Resource:  "/api/users/1",
		CreatedAt: time.Now().UTC(), // Mais recente
	}

	assert.NoError(t, repo.Create(ctx, log1))
	assert.NoError(t, repo.Create(ctx, log2))

	t.Run("lists all records in descending order", func(t *testing.T) {
		// Executa a função List de audit.go
		logs, err := repo.List(ctx, 10, 0)

		assert.NoError(t, err)
		assert.Len(t, logs, 2)
		// O primeiro registro deve ser o log2 (mais recente pelo Order("created_at DESC"))
		assert.Equal(t, log2.ID, logs[0].ID)
		assert.Equal(t, "DELETE", logs[0].Action)
	})

	t.Run("applies limit and offset correctly", func(t *testing.T) {
		logs, err := repo.List(ctx, 1, 1)

		assert.NoError(t, err)
		assert.Len(t, logs, 1)
		// O offset pula o primeiro e retorna o log1
		assert.Equal(t, log1.ID, logs[0].ID)
	})

	t.Run("ignores requests with the GET method", func(t *testing.T) {
		repo := &MockAuditRepo{Signal: make(chan struct{}, 1)}
		mw := audit.Middleware(repo, cfg.JWTSecret)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
		rec := httptest.NewRecorder()
		mw(next).ServeHTTP(rec, req)

		select {
		case <-repo.Signal:
			t.Fatal("GET nao deve disparar auditoria")
		case <-time.After(30 * time.Millisecond):
			// Sucesso
		}
	})

	t.Run("ignores client or server error responses (status >= 400)", func(t *testing.T) {
		repo := &MockAuditRepo{Signal: make(chan struct{}, 1)}
		mw := audit.Middleware(repo, cfg.JWTSecret)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
		})

		req := httptest.NewRequest(http.MethodPost, "/api/users", nil)
		rec := httptest.NewRecorder()
		mw(next).ServeHTTP(rec, req)

		select {
		case <-repo.Signal:
			t.Fatal("Erros >= 400 nao devem ser auditados")
		case <-time.After(30 * time.Millisecond):
			// Sucesso
		}
	})

	t.Run("extracts userID through the Bearer token header fallback", func(t *testing.T) {
		repo := &MockAuditRepo{Signal: make(chan struct{}, 1)}
		mw := audit.Middleware(repo, cfg.JWTSecret)

		userID := uuid.New()
		tokenStr, err := token.GenerateToken(userID, "user@test.com", uuid.New(), cfg.JWTSecret, 1)
		assert.NoError(t, err)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodPost, "/api/users/profile", nil)
		req.Header.Set("Authorization", "Bearer "+tokenStr)
		rec := httptest.NewRecorder()
		mw(next).ServeHTTP(rec, req)

		select {
		case <-repo.Signal:
			assert.NotNil(t, repo.CapturedLog.UserID)
			assert.Equal(t, userID, *repo.CapturedLog.UserID)
			assert.Equal(t, "/api/users/profile", repo.CapturedLog.Resource)
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout aguardando auditoria via Bearer token")
		}
	})

	t.Run("covers Write when WriteHeader has already been called", func(t *testing.T) {
		repo := &MockAuditRepo{Signal: make(chan struct{}, 1)}
		mw := audit.Middleware(repo, cfg.JWTSecret)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Chama WriteHeader antes do Write para que rw.statusCode != 0
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("created"))
		})

		req := httptest.NewRequest(http.MethodPost, "/api/roles", nil)
		rec := httptest.NewRecorder()
		mw(next).ServeHTTP(rec, req)

		select {
		case <-repo.Signal:
			assert.Equal(t, http.StatusCreated, rec.Code)
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout aguardando gravacao do log")
		}
	})
}
