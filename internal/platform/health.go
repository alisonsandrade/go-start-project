// Package platform
package platform

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"gorm.io/gorm"
)

type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// HealthCheckHandler expõe os endpoints de liveness e readiness
type HealthCheckHandler struct {
	db *gorm.DB
}

func NewHealthCheckHandler(db *gorm.DB) *HealthCheckHandler {
	return &HealthCheckHandler{db: db}
}

// Liveness (/healthz): Retorna 200 OK se a aplicação estiver respondendo
func (h *HealthCheckHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HealthResponse{
		Status:    "up",
		Timestamp: time.Now().UTC(),
	})
}

// Readiness (/readyz): Avalia se os serviços dependentes (PostgreSQL) aceitam tráfego
func (h *HealthCheckHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	sqlDB, err := h.db.DB()
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(HealthResponse{
			Status:    "down",
			Timestamp: time.Now().UTC(),
			Checks:    map[string]string{"database": "failed to obtain sql connection"},
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(HealthResponse{
			Status:    "down",
			Timestamp: time.Now().UTC(),
			Checks:    map[string]string{"database": err.Error()},
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(HealthResponse{
		Status:    "ready",
		Timestamp: time.Now().UTC(),
		Checks:    map[string]string{"database": "up"},
	})
}
