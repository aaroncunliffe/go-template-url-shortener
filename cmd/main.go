package main

import (
	"aaroncunliffe/url-shortener/internal/database"
	"aaroncunliffe/url-shortener/internal/logging"
	"context"
	"fmt"
	"os"

	"github.com/caarlos0/env"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

// Injected vars
var version string = "local"
var buildDate string = "now"

type config struct {
	DBUser string `env:"DB_USER" envDefault:"postgres"`
	DBPass string `env:"DB_PASS" envDefault:"password"`
	DBName string `env:"DB_NAME" envDefault:"postgres"`
	DBHost string `env:"DB_HOST" envDefault:"localhost"`
	DBPort string `env:"DB_PORT" envDefault:"5432"`
}

func main() {
	logger := logging.NewHumanReadableLogger()

	logger.Info("starting service", "version", version, "build date", buildDate)
	logger.Info("initialising dependencies...")
	defer logger.Info("shutdown complete")

	err := godotenv.Load()
	if err != nil {
		logger.Info("No .env file loaded")
	}

	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		logger.Error(err.Error())
	}
	// Ideally hide passwords and secrets from here
	logger.Info("config", "values", fmt.Sprintf("%+v", cfg))

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.Open(ctx, database.Config{
		User:       cfg.DBUser,
		Pass:       cfg.DBPass,
		Host:       cfg.DBHost,
		Name:       cfg.DBName,
		TlsEnabled: false,
	})
	if err != nil {
		logger.Error(err.Error())
		return
	}
	logger.Info("startup", "status", "DB connected", "host", cfg.DBHost)
	defer func() {
		logger.Info("shutdown", "status", "stopping DB support", "host", cfg.DBHost)
		defer db.Close()
	}()

	///

	q := database.New(db)
	link, err := q.GetLink(ctx, "test")
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetLink failed: %v\n", err)
		os.Exit(1)
	}

	logger.Info(fmt.Sprintf("got link: %+v", link))

}
