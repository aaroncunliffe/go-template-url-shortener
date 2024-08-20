package logging

import (
	"log/slog"
	"os"

	"github.com/charmbracelet/log"
)

func NewJsonLogger() *slog.Logger {
	opts := slog.HandlerOptions{}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &opts))

	return logger
}

func NewHumanReadableLogger() *slog.Logger {
	handler := log.New(os.Stdout)
	logger := slog.New(handler)

	return logger
}
