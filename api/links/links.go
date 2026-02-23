package links

import (
	"aaroncunliffe/url-shortener/internal/business/links"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"
)

// Combined validation and business logic
// In larger apps business logic should be abstracted to internal

type Handler struct {
	Logger *slog.Logger
	Links  links.Core
}

func (h Handler) LinkRedirect(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "path")

	h.Logger.Info("looking up url for path", slog.String("path", path))

	redirect, err := h.Links.ResolveLink(r.Context(), path)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		h.Logger.Error("%w", err)
		return
	}

	h.Logger.Info("performing redirect", slog.String("path", path), slog.String("redirect", redirect))
	http.Redirect(w, r, redirect, http.StatusFound)
}
