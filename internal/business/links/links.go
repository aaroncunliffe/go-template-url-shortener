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

func (h Core) ResolveLink(path string) (string, error) {
	link, err := h.Store.GetLinkByPath(context.Background(), path)
	if err != nil {
		return "", err
	}
	return link.OriginalUrl, nil
}
