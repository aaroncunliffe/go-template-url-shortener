package api

import (
	"log/slog"
	"net/http"

	"github.com/aaroncunliffe/go-template-url-shortener/internal/database"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/telemetry"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/web"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/web/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type Config struct {
	Logger    *slog.Logger
	DB        *database.Queries
	Telemetry *telemetry.Telemetry
}

func NewAPI(config Config) http.Handler {
	// Add nil checks for dependencies

	mux := chi.NewRouter()
	mux.Use(chimiddleware.RequestID)
	mux.Use(otelhttp.NewMiddleware(config.Telemetry.ServiceName))
	mux.Use(middleware.Logger(config.Logger))
	mux.Use(middleware.Metrics(config.Telemetry))

	routes(&web.Router{Mux: mux, Logger: config.Logger}, config)

	return mux
}
