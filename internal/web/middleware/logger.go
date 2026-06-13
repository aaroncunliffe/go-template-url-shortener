package middleware

import (
	"net/http"
	"time"

	"log/slog"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// Custom Logger middleware that supports slog for uniform logging
func Logger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Proxy to allow hooks
			ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()

			/*
				Check for X-Forwarded-For header, else we default to the remote address.
				This is not a bulletproof solution, there are better options depending on how this application is deployed

				In a real world scenario we should check that RemoteAddr is a trusted source (e.g. cloudflare, or a loadbalancer)
				before trusting X-Forwarded-For
			*/
			ipAddress := r.Header.Get("X-Forwarded-For") // && is from a trusted load balancer
			if ipAddress == "" {
				ipAddress = r.RemoteAddr
			}

			// In
			logger.InfoContext(r.Context(), "request started",
				"method", r.Method,
				"request_path", r.URL.Path,
				"ip_address", ipAddress,
				"user_agent", r.UserAgent(),
			)

			// Handler
			next.ServeHTTP(ww, r)

			// More likely resolve after request completion
			route := chi.RouteContext(r.Context()).RoutePattern()
			if route == "" {
				route = "unknown"
			}

			// Out
			logger.InfoContext(r.Context(), "request completed",
				"method", r.Method,
				"route_pattern", route,
				"request_path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration_ms", time.Since(t1).Milliseconds(),
			)
		})
	}
}
