package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aaroncunliffe/go-template-url-shortener/api"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/database"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/logging"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/telemetry"
	"github.com/aaroncunliffe/go-template-url-shortener/internal/web/debug"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

const serviceName = "url-shortener-api"

// Injected vars
var version string = "local"
var buildDate string = "now"

type config struct {
	DBUser    string `env:"DB_USER" envDefault:"postgres"`
	DBPass    string `env:"DB_PASS" envDefault:"password"`
	DBName    string `env:"DB_NAME" envDefault:"postgres"`
	DBHost    string `env:"DB_HOST" envDefault:"db"`
	WebPort   string `env:"WEB_PORT" envDefault:"8080"`
	DebugPort string `env:"DEBUG_PORT" envDefault:"4040"`

	TraceEndpoint    string  `env:"TRACE_ENDPOINT" envDefault:"http://tempo:4318"`
	TraceProbability float64 `env:"TraceProbability" envDefault:"1.0"`
}

func main() {
	os.Exit(run())
}

func run() int {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger := logging.NewJsonLogger()

	// =========================================================================
	// Startup
	logger.Info("starting service", "version", version, "build date", buildDate)
	logger.Info("initialising dependencies...")
	defer logger.Info("shutdown complete")

	// Always want to run in UTC.
	time.Local = time.UTC

	err := godotenv.Load()
	if err != nil {
		logger.Info("No .env file loaded")
	}

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		logger.Error("env.parse", "error", err.Error())
	}
	logger.Info("config",
		"db_user", cfg.DBUser,
		"db_name", cfg.DBName,
		"db_host", cfg.DBHost,
		"web_port", cfg.WebPort,
		"debug_port", cfg.DebugPort,
		"trace_endpoint", cfg.TraceEndpoint,
		"trace_probability", cfg.TraceProbability,
	)

	// Channel for termination Signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	// =========================================================================
	// Initialise Telemetry
	tel, err := telemetry.New(ctx, telemetry.Config{
		ServiceName:      serviceName,
		ServiceBuild:     version,
		TraceEndpoint:    cfg.TraceEndpoint,
		TraceProbability: cfg.TraceProbability,
	})
	if err != nil {
		logger.Error("telemetry setup", "error", err.Error())
		return 1
	}

	// =========================================================================
	// Database Support
	db, err := database.Open(ctx, database.Config{
		User:       cfg.DBUser,
		Pass:       cfg.DBPass,
		Host:       cfg.DBHost,
		Name:       cfg.DBName,
		TlsEnabled: false,
	})
	if err != nil {
		logger.Error("database open", "error", err.Error())
		return 1
	}
	logger.Info("startup", "status", "DB connected", "host", cfg.DBHost)
	defer func() {
		logger.Info("shutdown", "status", "stopping DB support", "host", cfg.DBHost)
		db.Close()
	}()

	// -------------------------------------------------------------------------
	// Start Servers

	// Channel specifically for server errors
	serverErrors := make(chan error, 1)

	// -------------------------------------------------------------------------
	// Start Debug Service
	// Internal only endpoint for healthchecks, debugging, and performance profiling
	debugServer := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.DebugPort),
		Handler:           debug.NewAPI(logger, db, tel),
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		logger.Info("startup", "status", "debug server listening", "port", cfg.DebugPort)
		if err := debugServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- fmt.Errorf("debug server error: %w", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Start API Service
	api := api.NewAPI(api.Config{
		Logger:    logger,
		DB:        database.New(db),
		Telemetry: tel,
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.WebPort),
		Handler: api,
		// Timeouts should be considered in production
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("startup", "status", "api server listening", "port", cfg.WebPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- fmt.Errorf("api server error: %w", err)
		}

	}()

	// =========================================================================
	// Shutdown
	select {
	case err := <-serverErrors:
		logger.Error("server error", "error", err.Error())
	case signal := <-signalCh:
		logger.Info("shutdown signal received", "signal", signal)
	case <-ctx.Done():
		logger.Info("context canceled")
	}

	// Trigger original context cancel for other services to listen to
	cancel()

	// Create new context to attempt graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer shutdownCancel()

	exitCode := 0

	logger.Info("shutdown", "status", "stopping api server")
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("forced api server shutdown", "error", err.Error())
		exitCode = 1
	} else {
		logger.Info("api server shut down gracefully")
	}

	logger.Info("shutdown", "status", "stopping debug server")
	if err := debugServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("forced debug server shutdown", "error", err.Error())
		exitCode = 1
	} else {
		logger.Info("debug server shut down gracefully")
	}

	logger.Info("shutdown", "status", "stopping telemetry")
	if err := tel.Shutdown(shutdownCtx); err != nil {
		logger.Error("forced telemetry shutdown", "error", err.Error())
		exitCode = 1
	} else {
		logger.Info("telemetry down gracefully")
	}

	logger.Info("shutdown", "status", "service stopped")
	return exitCode
}
