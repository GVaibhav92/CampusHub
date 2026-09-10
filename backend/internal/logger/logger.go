package logger

import (
	"log/slog"
	"os"
)

// returns a JSON structured logger
func New() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(handler)
}
