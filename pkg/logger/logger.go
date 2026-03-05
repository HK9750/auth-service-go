package logger

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	Level  string
	Format string
}

func New(cfg Config) *slog.Logger {
	level := parseLevel(cfg.Level)
	options := &slog.HandlerOptions{Level: level}

	if strings.EqualFold(cfg.Format, "text") {
		return slog.New(slog.NewTextHandler(os.Stdout, options))
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, options))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
