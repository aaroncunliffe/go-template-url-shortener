package middleware

import (
	"fmt"
	"net/http"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5/middleware"
)

// Custom Logger middleware that supports slog for uniform logging
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Proxy to allow hooks
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()

			// In
			scheme := "http"
			if r.TLS != nil {
				scheme = "https"
			}

			logger.Info("request",
				"timestamp", time.Now().Format(time.RFC3339),
				"scheme", scheme,
				"method", r.Method,
				"request", fmt.Sprintf("%s%s", r.Host, r.RequestURI),
				"protocol", r.Proto,
			)

			// Handler
			next.ServeHTTP(ww, r)

			// Out
			logger.Info("response",
				"timestamp", time.Now().Format(time.RFC3339),
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration", time.Since(t1).Seconds(),
			)
		})
	}
}
