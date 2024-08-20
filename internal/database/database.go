package database

import (
	"context"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	User       string
	Pass       string
	Host       string
	Name       string
	TlsEnabled bool
}

func Open(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	q := url.Values{}
	q.Set("timezone", "utc")
	q.Set("sslmode", "disable")
	if cfg.TlsEnabled {
		q.Set("sslmode", "require")
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.User, cfg.Pass),
		Host:     cfg.Host,
		Path:     cfg.Name,
		RawQuery: q.Encode(),
	}

	conn, err := pgxpool.New(ctx, u.String())
	if err != nil {
		return nil, err
	}

	return conn, nil
}
