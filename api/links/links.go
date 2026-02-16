package links

import (
	"aaroncunliffe/url-shortener/internal/database"
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v5"
)

// Combined validation and business logic
// In larger apps business logic should be abstracted to internal

type Handler struct {
	Logger *slog.Logger
	DB     *database.Queries
}

func (h Handler) LinkRedirect(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "path")

	h.Logger.Info("looking up url for path", slog.String("path", path))

	link, err := h.DB.GetLink(context.Background(), path)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			h.Logger.Warn(err.Error())
			w.WriteHeader(404)
			return
		}
		h.Logger.Error(err.Error())
	}

	if (link == database.Link{}) {
		w.WriteHeader(404)
		return
	}

	h.Logger.Info("performing redirect", slog.String("path", path), slog.String("redirect", link.OriginalUrl))
	http.Redirect(w, r, link.OriginalUrl, 302)
}
