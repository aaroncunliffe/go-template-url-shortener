package database

import (
	"context"
	"net/url"

	"github.com/exaring/otelpgx"
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

	poolCfg, err := pgxpool.ParseConfig(u.String())
	if err != nil {
		return nil, err
	}

	// Attach otel
	// Tracing coming soon
	poolCfg.ConnConfig.Tracer = otelpgx.NewTracer()

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}
