package debug

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Creates a standard approach for a debug interface
func NewAPI(build string, logger *slog.Logger, db *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux() // Don't need chi here

	// Basic health check
	mux.HandleFunc("/debug/liveness", func(w http.ResponseWriter, r *http.Request) {

		logger.Info("request",
			"path", r.RequestURI,
			"protocol", r.Proto,
		)

		response(http.StatusOK, w)
	})

	// Readiness to accept traffic check
	mux.HandleFunc("/debug/readiness", func(w http.ResponseWriter, r *http.Request) {
		var errs []error

		// Check status of individual services
		if err := db.Ping(r.Context()); err != nil {
			errs = append(errs, err)
		}

		logger.Info("request",
			"path", r.RequestURI,
			"protocol", r.Proto,
		)

		if len(errs) == 0 {
			response(http.StatusOK, w)
			return
		}

		response(http.StatusInternalServerError, w)
		logger.Error("readiness check failed", "count", len(errs), "errors", errors.Join(errs...))
	})

	return mux
}

// Own simple response format
func response(statusCode int, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(http.StatusText(statusCode)))
}
