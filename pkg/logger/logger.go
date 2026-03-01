package logger

import (
	"log/slog"
	"os"
)

// Logger is a wrapper around slog.Logger that provides additional functionality
type Logger struct {
	*slog.Logger
}

// New creates a new Logger instance based on the environment.
func New(env string) *Logger {
	var handler slog.Handler

	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// WithComponent adds a "component" field to the logger for better context in logs.
func (l *Logger) WithComponent(name string) *Logger {
	// We return a new *logger.Logger so we don't mutate the original
	return &Logger{
		Logger: l.Logger.With(slog.String("component", name)),
	}
}
