// Package platform
package platform

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// InitLogger configura o slog global para formato JSON
func InitLogger() {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	slog.SetDefault(slog.New(handler))
}

type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	bytes      int
}

func (rw *responseRecorder) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseRecorder) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.bytes += n
	return n, err
}

// SlogMiddleware registra cada requisição com tempo de resposta e request_id
func SlogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		reqID := chimiddleware.GetReqID(r.Context())

		rec := &responseRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rec, r)

		duration := time.Since(start)

		slog.InfoContext(
			r.Context(), "http_request",
			slog.String("request_id", reqID),
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", rec.statusCode),
			slog.Int("bytes", rec.bytes),
			slog.Duration("latency", duration),
			slog.String("remote_ip", r.RemoteAddr),
		)
	})
}
