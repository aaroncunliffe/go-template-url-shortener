package pgstore

import (
	"context"
	"errors"

	"github.com/aaroncunliffe/go-template-url-shortener/internal/business/links"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/database"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Postgres concrete store implementation
type PGStore struct {
	DB *database.Queries
}

func (s PGStore) GetLinkByPath(ctx context.Context, path string) (links.Link, error) {
	row, err := s.DB.GetLinkByPath(ctx, path)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return links.Link{}, links.ErrNotFound
		}
		return links.Link{}, err
	}
	return links.Link{ShortPath: row.ShortPath, OriginalURL: row.OriginalUrl}, nil
}

func (s PGStore) InsertLink(ctx context.Context, link links.Link) error {
	err := s.DB.InsertLink(ctx, database.InsertLinkParams{
		ShortPath:   link.ShortPath,
		OriginalUrl: link.OriginalURL,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return links.ErrConflict
		}
		return err
	}
	return nil
}
