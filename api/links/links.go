package links

import (
	"aaroncunliffe/url-shortener/internal/business/links"
	"aaroncunliffe/url-shortener/internal/web"
	"errors"
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
		// Not found
		if errors.Is(err, links.ErrNotFound) {
			web.ErrorJSON(w, http.StatusNotFound, err.Error())
			return
		}

		h.Logger.Error("resolve link", "error", err)
		web.ErrorJSON(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.Logger.Info("performing redirect", slog.String("path", path), slog.String("redirect", redirect))
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (h Handler) CreateLink(w http.ResponseWriter, r *http.Request) {

	var input CreateLinkRequest
	err := web.DecodeJSON(r, &input)
	if err != nil {
		h.Logger.Error("decode json", "error", err)
		web.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	err = web.ValidateStruct(input)
	if err != nil {
		h.Logger.Error("validate struct", "error", err)
		web.ErrorJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	h.Logger.Info("creating link", slog.String("origin", input.OriginURL))
	shortPath, err := h.Links.CreateLink(r.Context(), input.ShortPath, input.OriginURL)
	if err != nil {
		if errors.Is(err, links.ErrConflict) {
			web.ErrorJSON(w, http.StatusConflict, err.Error())
			return
		}

		h.Logger.Error("create link", "error", err)
		web.ErrorJSON(w, http.StatusInternalServerError, "internal server error")
		return
	}

	web.JSON(w, http.StatusCreated, CreateLinkResponse{ShortPath: shortPath})
	return
}
