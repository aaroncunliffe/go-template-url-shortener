package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/aaroncunliffe/go-template-url-shortener/internal/telemetry"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func Metrics(t *telemetry.Telemetry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			t1 := time.Now()

			next.ServeHTTP(ww, r)

			status := strconv.Itoa(ww.Status())
			duration := time.Since(t1).Seconds()

			route := chi.RouteContext(r.Context()).RoutePattern()
			if route == "" {
				route = "unknown" // Sensible default to not explode prometheus metrics
			}

			t.RecordHTTPRequest(r.Context(), r.Method, route, status, duration)
		})
	}
}
