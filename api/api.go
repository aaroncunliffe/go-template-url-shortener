package api

import (
	"log/slog"
	"net/http"

	"github.com/aaroncunliffe/go-template-url-shortener/internal/database"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/web"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/web/middleware"

	"github.com/go-chi/chi"
)

// Set up API with dependencies to be passed to required handlers

type Config struct {
	Logger *slog.Logger
	DB     *database.Queries
}

func NewAPI(config Config) http.Handler {
	mux := chi.NewRouter()

	// Global Middleware
	// Custom logger middleware for uniform logging
	mux.Use(middleware.Logger(config.Logger))

	// No requirement for cors, but this can be added with chi easily here
	// https://github.com/go-chi/cors

	// Attach routes
	routes(&web.Router{Mux: mux, Logger: config.Logger}, config)

	return mux
}
