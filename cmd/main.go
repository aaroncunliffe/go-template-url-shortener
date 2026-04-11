package main

import (
	"aaroncunliffe/url-shortener/api"
	"aaroncunliffe/url-shortener/internal/database"
	"aaroncunliffe/url-shortener/internal/logging"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

// Injected vars
var version string = "local"
var buildDate string = "now"

type config struct {
	DBUser  string `env:"DB_USER" envDefault:"postgres"`
	DBPass  string `env:"DB_PASS" envDefault:"password"`
	DBName  string `env:"DB_NAME" envDefault:"postgres"`
	DBHost  string `env:"DB_HOST" envDefault:"db"`
	WebPort string `env:"WEB_PORT" envDefault:"8081"`
}

func main() {
	logger := logging.NewJsonLogger()
	if version == "local" {
		logger = logging.NewHumanReadableLogger()
	}

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
	// Ideally hide passwords and secrets from here
	logger.Info("config", "values", fmt.Sprintf("%+v", cfg))

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Channel for termination Signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

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
		return
	}
	logger.Info("startup", "status", "DB connected", "host", cfg.DBHost)
	defer func() {
		logger.Info("shutdown", "status", "stopping DB support", "host", cfg.DBHost)
		defer db.Close()
	}()

	api := api.NewAPI(api.Config{
		Logger: logger,
		DB:     database.New(db),
	})

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.WebPort),
		Handler: api,
		// Defaults for Timeouts should be considered in production
	}

	// Channel specifically for server errors
	serverErrors := make(chan error, 1)

	go func() {
		logger.Info("startup", "status", "web server listening", "port", cfg.WebPort)
		serverErrors <- server.ListenAndServe()
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
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("forced server shutdown", "error", err.Error())
		exitCode = 1
	} else {
		logger.Info("server shut down gracefully")
	}

	logger.Info("shutdown", "status", "service stopped")
	os.Exit(exitCode)
}
