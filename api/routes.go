package api

import (
	"github.com/aaroncunliffe/go-template-url-shortener/api/links"

	linksCore "github.com/aaroncunliffe/go-template-url-shortener/internal/business/links"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/business/links/pgstore"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/web"
)

// Define API Routes
// Defined here to be easier to read with much larger applications
func routes(r *web.Router, config Config) {

	linksHandler := links.Handler{
		Logger:    config.Logger,
		Telemetry: config.Telemetry,
		Links: linksCore.Core{
			Logger: config.Logger,
			// Plug in concrete store - can be Postgres or Redis
			Store: pgstore.PGStore{DB: config.DB},
		},
	}
	r.Get("/{path}", linksHandler.LinkRedirect)
	r.Post("/api/link", linksHandler.CreateLink)

}
