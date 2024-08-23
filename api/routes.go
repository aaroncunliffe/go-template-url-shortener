package api

import (
	"aaroncunliffe/url-shortener/api/links"

	"github.com/go-chi/chi"
)

// Define API Routes
// Defined here to be easier to read with much larger applications
func routes(mux *chi.Mux) {
	mux.Get("/", links.LinkRedirect)
}
