package debug

import (
	"errors"
	"expvar"
	"log/slog"
	"net/http"
	"net/http/pprof"

	"github.com/aaroncunliffe/go-template-url-shortener/internal/telemetry"
	"github.com/arl/statsviz"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Creates a standard mux for internal use
// Healthchecks - readiness / liveness
// debugging / performance profiling - pprof, expvar, and statviz
func NewAPI(logger *slog.Logger, db *pgxpool.Pool, tel *telemetry.Telemetry) http.Handler {
	mux := http.NewServeMux()

	// pprof
	mux.HandleFunc("/debug/pprof", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	mux.Handle("/debug/vars", expvar.Handler())

	if err := statsviz.Register(mux); err != nil {
		logger.Error("registering statviz failed", "error", err)
	}

	mux.Handle("/debug/metrics", tel.MetricsHandler())

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

// Simple response format
func response(statusCode int, w http.ResponseWriter) {
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(http.StatusText(statusCode)))
}
