package links_test

import (
	"aaroncunliffe/url-shortener/api"
	"aaroncunliffe/url-shortener/internal/database"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const schemaPath = "../../configs/database/schema.sql"

type integrationEnv struct {
	server  *httptest.Server
	queries *database.Queries
	pool    *pgxpool.Pool
	ctx     context.Context

	container *postgres.PostgresContainer
}

var (
	sharedEnv  integrationEnv
	startupErr error
	schemaSQL  string
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		startupErr = fmt.Errorf("read schema sql: %w", err)
		os.Exit(m.Run())
	}
	schemaSQL = string(schemaBytes)

	provider, err := testcontainers.ProviderDocker.GetProvider()
	if err != nil {
		startupErr = err
		os.Exit(m.Run())
	}

	if err := provider.Health(ctx); err != nil {
		startupErr = err
		os.Exit(m.Run())
	}

	sharedEnv, err = startIntegrationEnv(ctx)
	if err != nil {
		startupErr = err
		fmt.Fprintf(os.Stderr, "failed to start integration test environment: %v\n", err)
		os.Exit(m.Run())
	}

	// Run tests
	code := m.Run()
	// Done running tests

	// Clean up
	sharedEnv.server.Close()
	sharedEnv.pool.Close()
	if err := testcontainers.TerminateContainer(sharedEnv.container); err != nil {
		fmt.Fprintf(os.Stderr, "failed to terminate integration container: %v\n", err)
		if code == 0 {
			code = 1
		}
	}

	os.Exit(code)
}

// Call from test handlers
func integrationEnvForTest(t *testing.T) integrationEnv {
	t.Helper()

	if startupErr != nil {
		t.Skipf("integration environment unavailable: %v", startupErr)
	}

	// Reset DB to clean slate
	if _, err := sharedEnv.pool.Exec(sharedEnv.ctx, `DROP SCHEMA public CASCADE;`); err != nil {
		t.Fatalf("drop test schema: %v", err)
	}
	if _, err := sharedEnv.pool.Exec(sharedEnv.ctx, schemaSQL); err != nil {
		t.Fatalf("rebuild test schema: %v", err)
	}

	return sharedEnv
}

func startIntegrationEnv(ctx context.Context) (integrationEnv, error) {
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("url_shortener_test"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("password"),
		postgres.WithInitScripts(schemaPath),
		testcontainers.WithAdditionalWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(5*time.Second),
			wait.ForListeningPort("5432/tcp").WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		return integrationEnv{}, fmt.Errorf("start postgres container: %w", err)
	}

	connString, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		testcontainers.TerminateContainer(postgresContainer)
		return integrationEnv{}, fmt.Errorf("build connection string: %w", err)
	}

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		testcontainers.TerminateContainer(postgresContainer)
		return integrationEnv{}, fmt.Errorf("open pgx pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		testcontainers.TerminateContainer(postgresContainer)
		return integrationEnv{}, fmt.Errorf("ping postgres: %w", err)
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := httptest.NewServer(api.NewAPI(api.Config{
		Logger: logger,
		DB:     database.New(pool),
	}))

	return integrationEnv{
		server:    server,
		queries:   database.New(pool),
		pool:      pool,
		ctx:       ctx,
		container: postgresContainer,
	}, nil
}
