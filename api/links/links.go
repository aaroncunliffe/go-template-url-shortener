package links

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aaroncunliffe/go-template-url-shortener/internal/business/links"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/web"

	"github.com/go-chi/chi/v5"
)

// Combined validation and business logic
// In larger apps business logic should be abstracted to internal

type Handler struct {
	Logger *slog.Logger
	Links  links.Core
}

func (h Handler) LinkRedirect(w http.ResponseWriter, r *http.Request) error {
	path := chi.URLParam(r, "path")
	h.Logger.Info("looking up url for path", slog.String("path", path))

	redirect, err := h.Links.ResolveLink(r.Context(), path)
	if err != nil {
		if errors.Is(err, links.ErrNotFound) {
			return web.NewRequestError(http.StatusNotFound, links.ErrNotFound, web.Trusted)
		}
		return fmt.Errorf("resolving link: %w", err)
	}

	// Sensible check for DB tampering
	if err := web.ValidRedirectURL(redirect); err != nil {
		return web.NewRequestError(http.StatusBadRequest, err, web.Untrusted)
	}

	h.Logger.Info("performing redirect", slog.String("path", path), slog.String("redirect", redirect))

	// Validated above, redirect here is intentional
	//nolint:gosec
	http.Redirect(w, r, redirect, http.StatusFound)
	return nil
}

func (h Handler) CreateLink(w http.ResponseWriter, r *http.Request) error {
	var input CreateLinkRequest
	if err := web.DecodeJSON(r, &input); err != nil {
		return web.NewRequestError(http.StatusBadRequest, fmt.Errorf("decode %w", err), web.Untrusted)
	}

	if err := web.ValidateStruct(input); err != nil {
		return web.NewRequestError(http.StatusBadRequest, fmt.Errorf("validating struct %w", err), web.Untrusted)
	}

	// Create link
	h.Logger.Info("creating link", slog.String("origin", input.OriginURL))
	shortPath, err := h.Links.CreateLink(r.Context(), input.ShortPath, input.OriginURL)
	if err != nil {
		if errors.Is(err, links.ErrConflict) {
			return web.NewRequestError(http.StatusConflict, links.ErrConflict, web.Trusted)
		}
		return fmt.Errorf("creating link: %w", err)
	}

	web.JSON(w, http.StatusCreated, CreateLinkResponse{ShortPath: shortPath})
	return nil
}
