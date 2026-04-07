package links

import (
	"aaroncunliffe/url-shortener/internal/database"
	"context"
	"log/slog"
)

type Core struct {
	Logger *slog.Logger
	Store  Storer
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

func (h Core) CreateLink(ctx context.Context, shorthPath string, originURL string) (string, error) {
	err := h.Store.InsertLink(ctx, shorthPath, originURL)
	if err != nil {
		return "", err
	}

	return shorthPath, nil
}
