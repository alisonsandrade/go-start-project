// Package platform
package platform

import (
	"net"
	"net/http"
	"time"

	"github.com/go-chi/httprate"
)

// RateLimiter cria um limitador seguro usando o IP da conexão física (r.RemoteAddr).
func RateLimiter(requestLimit int, windowLength time.Duration) func(http.Handler) http.Handler {
	limiter := httprate.NewRateLimiter(
		requestLimit,
		windowLength,
		httprate.WithKeyFuncs(func(r *http.Request) (string, error) {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				return r.RemoteAddr, nil
			}
			return host, nil
		}),
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			ErrorJSON(w, http.StatusTooManyRequests, "muitas tentativas. Tente novamente mais tarde.")
		}),
	)

	return limiter.Handler
}
