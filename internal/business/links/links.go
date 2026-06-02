package links

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/aaroncunliffe/go-template-url-shortener/internal/database"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/shortcode"
)

type Core struct {
	Logger *slog.Logger
	Store  Storer

	generate func() (string, error)
}

// Storer interface
// Abstraction of database specific implementation
// without needing to change business logic.
type Storer interface {
	GetLinkByPath(ctx context.Context, shortPath string) (database.Link, error)
	InsertLink(ctx context.Context, shortPath string, originalURL string) error
}

func (h Core) ResolveLink(ctx context.Context, path string) (string, error) {
	link, err := h.Store.GetLinkByPath(ctx, path)
	if err != nil {
		return "", err
	}
	return link.OriginalUrl, nil
}

// Create links using a generated short path if required
func (h Core) CreateLink(ctx context.Context, shortPath string, originURL string) (string, error) {

	// Easy route - attempt to store the manually set path first
	if shortPath != "" {
		if err := h.Store.InsertLink(ctx, shortPath, originURL); err != nil {
			return "", err
		}
		return shortPath, nil
	}

	// Generate short code with a sensible backoff
	generate := h.pathGenerator()
	for attempt := range maxGenerateAttempts {
		shortPath, err := generate()
		if err != nil {
			return "", fmt.Errorf("generate short code: %w", err)
		}

		err = h.Store.InsertLink(ctx, shortPath, originURL)
		if err == nil {
			return shortPath, nil
		}

		// Break loop if not conflict
		if !errors.Is(err, ErrConflict) {
			return "", err
		}

		h.Logger.Warn("generated short code conflict",
			slog.Int("attempt", attempt+1),
		)
	}

	return "", ErrShortCodeGenerationFailed
}

// pathGenerator returns the generator used for auto-created short paths.
// Tests can inject h.generate; production falls back to shortcode.Generate.
func (h Core) pathGenerator() func() (string, error) {
	if h.generate != nil {
		return h.generate
	}
	return shortcode.Generate
}
