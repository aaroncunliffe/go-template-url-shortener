package pgstore

import (
	"aaroncunliffe/url-shortener/internal/business/links"
	"aaroncunliffe/url-shortener/internal/database"
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Postgres concrete store implementation
type PGStore struct {
	DB *database.Queries
}

func (s PGStore) GetLinkByPath(ctx context.Context, path string) (database.Link, error) {
	row, err := s.DB.GetLinkByPath(ctx, path)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return database.Link{}, links.ErrNotFound
		}
		return database.Link{}, err
	}
	return database.Link{ShortPath: row.ShortPath, OriginalUrl: row.OriginalUrl}, nil
}

func (s PGStore) InsertLink(ctx context.Context, shortPath string, originalURL string) error {
	err := s.DB.InsertLink(ctx, database.InsertLinkParams{
		ShortPath:   shortPath,
		OriginalUrl: originalURL,
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
