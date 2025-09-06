package logger

import (
	"log/slog"
	"os"
)

// Logger wraps slog.Logger with additional methods
type Logger struct {
	*slog.Logger
}

// New creates a new logger instance
func New() *Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}

	var handler slog.Handler = slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &Logger{Logger: logger}
}

// Fatal logs a fatal error and exits the program
func (l *Logger) Fatal(msg string, args ...any) {
	l.Error(msg, args...)
	os.Exit(1)
}
