package logger

import (
	"io"
	"os"

	"github.com/charmbracelet/log"
)

// Logger interface for logging operations.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

// Config holds logging configuration.
type Config struct {
	Level        string
	Output       io.Writer
	ReportCaller bool
	Prefix       string
}

// DefaultConfig returns sensible default configuration.
func DefaultConfig() *Config {
	return &Config{
		Level:        "info",
		Output:       os.Stderr,
		ReportCaller: false,
		Prefix:       "",
	}
}

// NewLogger creates a new logger with the given configuration.
//

func NewLogger(cfg *Config) Logger {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	level, err := log.ParseLevel(cfg.Level)
	if err != nil {
		level = log.InfoLevel
	}

	logger := log.NewWithOptions(cfg.Output, log.Options{
		Level:           level,
		ReportTimestamp: true,
		ReportCaller:    cfg.ReportCaller,
		Prefix:          cfg.Prefix,
		Formatter:       log.TextFormatter,
	})

	return &charmLogger{logger: logger}
}

// charmLogger implements Logger interface using charmbracelet/log.
type charmLogger struct {
	logger *log.Logger
}

func (l *charmLogger) Debug(msg string, args ...any) {
	if len(args) == 0 {
		l.logger.Debug(msg)

		return
	}

	l.logger.Debug(msg, args...)
}

func (l *charmLogger) Info(msg string, args ...any) {
	if len(args) == 0 {
		l.logger.Info(msg)

		return
	}

	l.logger.Info(msg, args...)
}

func (l *charmLogger) Warn(msg string, args ...any) {
	if len(args) == 0 {
		l.logger.Warn(msg)

		return
	}

	l.logger.Warn(msg, args...)
}

func (l *charmLogger) Error(msg string, args ...any) {
	if len(args) == 0 {
		l.logger.Error(msg)

		return
	}

	l.logger.Error(msg, args...)
}

// NoOpLogger provides a no-op logger implementation.
type NoOpLogger struct{}

func (l *NoOpLogger) Debug(msg string, args ...any) {}
func (l *NoOpLogger) Info(msg string, args ...any)  {}
func (l *NoOpLogger) Warn(msg string, args ...any)  {}
func (l *NoOpLogger) Error(msg string, args ...any) {}

// Default returns a no-op logger by default.
//
//nolint:gochecknoglobals // Default logger instance for convenience
var Default Logger = &NoOpLogger{}
