package api

import (
	"aaroncunliffe/url-shortener/api/links"

	linksCore "aaroncunliffe/url-shortener/internal/business/links"
	"aaroncunliffe/url-shortener/internal/business/links/pgstore"

	"github.com/go-chi/chi"
)

// Define API Routes
// Defined here to be easier to read with much larger applications
func routes(mux *chi.Mux, config Config) {

	linksHandler := links.Handler{
		Logger: config.Logger,
		Links: linksCore.Core{
			Logger: config.Logger,
			Store:  pgstore.PGStore{DB: config.DB},
		},
	}
	mux.Get("/{path}", linksHandler.LinkRedirect)

	// API
	// Individual REST API Routes here
	mux.Post("/api/link", linksHandler.CreateLink)

}
