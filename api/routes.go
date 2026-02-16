package api

import (
	"aaroncunliffe/url-shortener/api/links"

	"github.com/go-chi/chi"
)

// Define API Routes
// Defined here to be easier to read with much larger applications
func routes(mux *chi.Mux, config Config) {

	linksHandler := links.Handler{
		DB:     config.DB,
		Logger: config.Logger,
	}
	mux.Get("/{path}", linksHandler.LinkRedirect)

	// api
	mux.Route("/api", func(r chi.Router) {
		// Individual REST API Routes here
	})

}
