package debug

import (
	"errors"
	"expvar"
	"log/slog"
	"net/http"
	"net/http/pprof"

	"github.com/arl/statsviz"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Creates a standard mux for internal use
// Healthchecks - readiness / liveness
// debugging / performance profiling - pprof, expvar, and statviz
func NewAPI(logger *slog.Logger, db *pgxpool.Pool) http.Handler {
	mux := http.NewServeMux() // Don't need chi here

	// pprof
	mux.HandleFunc("/debug/pprof", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	// expvar
	mux.Handle("/debug/vars", expvar.Handler())

	// Statviz tool
	if err := statsviz.Register(mux); err != nil {
		logger.Error("registering statviz failed", "error", err)
	}

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
