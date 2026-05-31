package api

import (
	"log/slog"
	"net/http"

	"github.com/aaroncunliffe/go-template-url-shortener/internal/database"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/telemetry"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/web"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/web/middleware"

	"github.com/go-chi/chi/v5"
)

type Config struct {
	Logger    *slog.Logger
	DB        *database.Queries
	Telemetry *telemetry.Telemetry
}

func NewAPI(config Config) http.Handler {

	// Add nil checks for dependencies

	mux := chi.NewRouter()
	mux.Use(middleware.Logger(config.Logger))
	mux.Use(middleware.Metrics(config.Telemetry))

	routes(&web.Router{Mux: mux, Logger: config.Logger}, config)

	return mux
}
