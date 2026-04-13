package logging

import (
	"log/slog"
	"os"
	"strings"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level: parseLevel(os.Getenv("EU5_LOG_LEVEL")),
}))

func parseLevel(raw string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Debugf(format string, args ...any) {
	logger.Debug(format, args...)
}

func Infof(format string, args ...any) {
	logger.Info(format, args...)
}

func Warnf(format string, args ...any) {
	logger.Warn(format, args...)
}

func Errorf(format string, args ...any) {
	logger.Error(format, args...)
}
