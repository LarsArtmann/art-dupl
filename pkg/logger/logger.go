package logger

import (
	"io"
	"os"

	"charm.land/log/v2"
)

// Logger is the canonical logging interface for the project.
// Both the SDK and CLI reference this interface so that logger
// conformance is compiler-checked.
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

func NewLogger(cfg *Config) *charmLogger {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	level, err := log.ParseLevel(cfg.Level)
	if err != nil {
		level = log.InfoLevel
	}

	logger := log.NewWithOptions(
		cfg.Output,
		log.Options{ //nolint:exhaustruct // third-party struct; rely on library defaults for unset fields
			Level:           level,
			ReportTimestamp: true,
			ReportCaller:    cfg.ReportCaller,
			Prefix:          cfg.Prefix,
			Formatter:       log.TextFormatter,
		},
	)

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

func (l *NoOpLogger) Debug(_ string, _ ...any) {}
func (l *NoOpLogger) Info(_ string, _ ...any)  {}
func (l *NoOpLogger) Warn(_ string, _ ...any)  {}
func (l *NoOpLogger) Error(_ string, _ ...any) {}

// Compile-time assertions that logger implementations satisfy the Logger contract.
var (
	_ Logger = (*charmLogger)(nil)
	_ Logger = (*NoOpLogger)(nil)
)

// Default returns a no-op logger by default.
//
//nolint:gochecknoglobals // Default logger instance for convenience
var Default = &NoOpLogger{}
