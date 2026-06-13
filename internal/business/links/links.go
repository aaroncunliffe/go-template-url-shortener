package links

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

type Core struct {
	Logger *slog.Logger
	Store  Storer

	Generate func() (string, error)
}

// Storer interface
// Abstraction of database specific implementation
// without needing to change business logic.
type Storer interface {
	GetLinkByPath(ctx context.Context, shortPath string) (Link, error)
	InsertLink(ctx context.Context, link Link) error
}

func (c Core) ResolveLink(ctx context.Context, path string) (string, error) {
	link, err := c.Store.GetLinkByPath(ctx, path)
	if err != nil {
		return "", err
	}
	return link.OriginalURL, nil
}

// Create links using a generated short path if required
func (c Core) CreateLink(ctx context.Context, shortPath string, originalURL string) (string, error) {
	link := Link{
		ShortPath:   shortPath,
		OriginalURL: originalURL,
	}

	// Easy route - attempt to store the manually set path first
	if shortPath != "" {
		if err := c.Store.InsertLink(ctx, link); err != nil {
			return "", err
		}
		return shortPath, nil
	}

	for attempt := range maxGenerateAttempts {
		generatedShortPath, err := c.Generate()
		link.ShortPath = generatedShortPath
		if err != nil {
			return "", fmt.Errorf("generate short code: %w", err)
		}

		err = c.Store.InsertLink(ctx, link)
		if err == nil {
			return generatedShortPath, nil
		}

		// Break loop if not conflict
		if !errors.Is(err, ErrConflict) {
			return "", err
		}

		c.Logger.WarnContext(ctx, "generated short code conflict",
			slog.Int("attempt", attempt+1),
		)
	}

	return "", ErrShortCodeGenerationFailed
}
